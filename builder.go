package chd

import (
	"errors"
	"math/big"
	"math/rand"
	"sort"
	"time"
)

type item struct {
	key  []byte
	data []byte
}

// Builder manages adding
// of items and map creation.
type Builder struct {
	items map[string]*item

	keySize    int
	itemSize   int
	maxKeySize int
}

type hash struct {
	h1   uint64
	h2   uint64
	h3   uint64
	item *item
}

type bucket struct {
	index  uint64
	hashes []*hash
}

type buckets []bucket

func (b buckets) Len() int           { return len(b) }
func (b buckets) Less(i, j int) bool { return len(b[i].hashes) > len(b[j].hashes) }
func (b buckets) Swap(i, j int)      { b[i], b[j] = b[j], b[i] }

// NewBuilder returns a new map builder.
func NewBuilder() *Builder {
	return &Builder{map[string]*item{}, 0, 0, 0}
}

// Add adds the key-value pair to the builder.
func (b *Builder) Add(key, value []byte) {
	data := append(key, value...)

	if len(b.items) == 0 {
		b.keySize = len(key)
		b.itemSize = len(data)
	}

	if b.keySize != len(key) {
		b.keySize = -1
	}

	if b.itemSize != len(data) {
		b.itemSize = -1
	}

	if len(key) > b.maxKeySize {
		b.maxKeySize = len(key)
	}

	b.items[string(key)] = &item{key, data}
}

// Build creates a map given a CompactArray.
// If array is nil, it will use a plain integer
// array instead. Note that array must be gob encodable.
func (b *Builder) Build(array CompactArray) *Map {
	rand.Seed(time.Now().UTC().UnixNano())

	items := make([]*item, 0, len(b.items))
	for _, v := range b.items {
		items = append(items, v)
	}

	bucketSize := 5
	loadFactor := 1.0
	tableSize := int(float64(len(items)) / loadFactor)
	tableSize = nearestPrime(tableSize)

	// Try and try until successful
	for {
		const numTries = 3
		for i := 0; i < numTries; i++ {
			seed := [2]uint64{
				uint64(rand.Int63()),
				uint64(rand.Int63()),
			}

			m, err := b.build(seed, bucketSize, tableSize, items, array)
			if err == nil {
				return m
			}
		}

		// If unsuccessful, reduce the bucket
		// size first and then the load factor
		if bucketSize > 1 {
			bucketSize--
		} else {
			bucketSize = 5
			loadFactor *= 0.90

			tableSize = int(float64(len(items)) / loadFactor)
			tableSize = nearestPrime(tableSize)
		}
	}
}

// build tries to create a map and
// returns an error if unsuccessful.
func (b *Builder) build(
	seed [2]uint64,
	bucketSize,
	tableSize int,
	items []*item,
	array CompactArray) (*Map, error) {

	nbuckets := uint64(len(items)/bucketSize) + 1
	buckets := make(buckets, nbuckets)

	hashes := make([]*hash, len(items))
	hashIdx := make([]uint64, nbuckets)

	// Calculate hashes
	for i := range hashes {
		h1, h2, h3, _ := spookyHash(items[i].key, seed[0], seed[1])
		hashes[i] = &hash{h1, h2, h3, items[i]}
	}

	// Put hashes into buckets
	for _, h := range hashes {
		bidx := h.h1 % nbuckets
		buckets[bidx].index = bidx
		buckets[bidx].hashes = append(buckets[bidx].hashes, h)
	}

	// Sort buckets in decreasing size
	sort.Sort(buckets)

	ts := uint64(tableSize)
	table := make([]*item, ts)

	maxHashIdx := ts * ts
	indices := make([]uint64, 0, len(buckets[0].hashes))

	// Process buckets and populate table
	for _, b := range buckets {
		if len(b.hashes) == 0 {
			continue
		}
		hidx := uint64(0)

	NextHashIdx:
		for {
			if hidx >= maxHashIdx {
				return nil, errors.New("chd: can't find collission-free hash function")
			}

			d0 := hidx / ts
			d1 := hidx % ts

			indices = indices[:0]
			for _, h := range b.hashes {
				idx := (h.h2 + (d0 * h.h3) + d1) % ts

				if table[idx] != nil {
					// Collission has occured, clear
					// table of previously added items
					for _, n := range indices {
						table[n] = nil
					}

					hidx++
					continue NextHashIdx
				}

				table[idx] = h.item
				indices = append(indices, idx)
			}

			hashIdx[b.index] = hidx
			break
		}
	}

	// Construct Map fields
	if array == nil {
		array = newIntArray(len(hashIdx))
	}
	for _, idx := range hashIdx {
		array.Add(int(idx))
	}

	offset := 0
	idata := newItemData(tableSize, b.itemSize, b.keySize)
	for i, itm := range table {
		if itm == nil {
			// Add sentinel
			idata.addSize(i, -1)
			continue
		}

		idata.addOffset(i, offset)
		idata.addSize(i, len(itm.data))
		idata.addKeySize(i, len(itm.key))

		offset += len(itm.data)
	}

	// sentinel is used by the iterator
	sentinel := []byte{}
	if b.maxKeySize > 0 {
		sentinel = randBytes(b.maxKeySize)
		for _, ok := b.items[string(sentinel)]; ok; {
			sentinel = randBytes(b.maxKeySize)
		}
	}

	var padding []byte
	if b.itemSize > 0 {
		padding = make([]byte, b.itemSize-len(sentinel))
		padding = append(sentinel, padding...)
	}

	data := make([]byte, 0, offset)
	for _, itm := range table {
		if itm == nil {
			data = append(data, padding...)
		} else {
			data = append(data, itm.data...)
		}
	}

	m := &Map{
		seed,
		data,
		idata,
		b.keySize,
		b.itemSize,
		sentinel,
		len(items),
		tableSize,
		array,
	}

	return m, nil
}

func nearestPrime(num int) int {
	if num&1 == 0 {
		num++
	}

	for !big.NewInt(int64(num)).ProbablyPrime(10) {
		num += 2
	}

	return num
}

func randBytes(size int) []byte {
	b := make([]byte, size)
	for i := range b {
		b[i] = byte(rand.Intn(256))
	}

	return b
}
