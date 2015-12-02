package chd

import (
	"bytes"
	"errors"
	"math"
	"math/big"
	"math/rand"
	"sort"
	"time"
)

// BuildOptions specifies the
// options for building a map.
type BuildOptions struct {
	// LoadFactor sets the load
	// factor. Lower values results
	// in faster build times. Default
	// value is 1.0
	//
	// If ForceBuild is enabled the
	// actual load factor may differ
	// significantly from the set value.
	LoadFactor float64

	// BucketSize sets the average
	// number of keys per bucket.
	// Default value is 5.
	BucketSize int

	// ForceBuild indicates that the
	// Builder.Build method will always
	// succeed. This is done by decreasing
	// the load factor every time it fails.
	// Default value is true.
	ForceBuild bool
}

// NewBuildOptions creates build
// options with default values.
func NewBuildOptions() *BuildOptions {
	return &BuildOptions{1.0, 5, true}
}

type item struct {
	key []byte

	// counter is used for
	// removing key duplicates
	counter int

	deleted bool
}

type items []item

func (it items) Len() int {
	return len(it)
}

func (it items) Less(i, j int) bool {
	cmp := bytes.Compare(it[i].key, it[j].key)
	if cmp < 0 {
		return true
	} else if cmp > 0 {
		return false
	}

	// if cmp == 0
	return it[i].counter > it[j].counter
}

func (it items) Swap(i, j int) {
	it[i], it[j] = it[j], it[i]
}

// Builder manages adding
// of items and map creation.
type Builder struct {
	items   items
	counter int
	opts    *BuildOptions
}

type hash struct {
	h1 uint64
	h2 uint64
}

type bucket struct {
	index  uint64
	hashes []hash
}

type buckets []bucket

func (b buckets) Len() int {
	return len(b)
}

func (b buckets) Less(i, j int) bool {
	return len(b[i].hashes) > len(b[j].hashes)
}

func (b buckets) Swap(i, j int) {
	b[i], b[j] = b[j], b[i]
}

// NewBuilder returns a new map builder given
// the build options. If opts is nil, the default
// values are used.
func NewBuilder(opts *BuildOptions) *Builder {
	if opts == nil {
		opts = NewBuildOptions()
	}

	if opts.LoadFactor > 1.0 || opts.LoadFactor <= 0.0 {
		panic("chd: invalid load factor")
	}

	return &Builder{opts: opts}
}

// Add adds a given key to the builder.
func (b *Builder) Add(key []byte) {
	item := item{key, b.counter, false}
	b.items = append(b.items, item)

	b.counter++
}

// Delete removes the item with the given key.
func (b *Builder) Delete(key []byte) {
	item := item{key, b.counter, true}
	b.items = append(b.items, item)
	b.counter++
}

// Build creates a map.
func (b *Builder) Build() (m *Map, err error) {
	if len(b.items) == 0 {
		return &Map{}, nil
	}

	rand.Seed(time.Now().UTC().UnixNano())

	// Sort items in ascending order
	// of keys and decreasing counter
	sort.Sort(b.items)

	// Remove duplicates and deleted items by
	// moving them to the front and then slicing
	front := 0
	pkey := make([]byte, len(b.items[0].key)+1)
	for i, item := range b.items {
		if bytes.Equal(pkey, item.key) || item.deleted {
			b.items[front], b.items[i] = b.items[i], b.items[front]
			front++
		}

		pkey = item.key
	}
	b.items = b.items[front:]

	loadFactor := b.opts.LoadFactor
	tableSize := int(float64(len(b.items)) / loadFactor)
	tableSize = nearestPrime(tableSize)

	// Try building the map
	for {
		const numTries = 3
		for i := 0; i < numTries; i++ {
			seed := [2]uint64{
				uint64(rand.Int63()),
				uint64(rand.Int63()),
			}

			m, err = b.build(seed, b.opts.BucketSize, tableSize, b.items)
			if err == nil {
				return m, nil
			}
		}

		// If ForceBuild is enabled, reduce load factor and try again
		if b.opts.ForceBuild {
			loadFactor *= 0.90

			tableSize = int(float64(len(b.items)) / loadFactor)
			tableSize = nearestPrime(tableSize)
		} else {
			return nil, err
		}
	}
}

// build tries to create a map and
// returns an error if unsuccessful.
func (b *Builder) build(
	seed [2]uint64,
	bucketSize,
	tableSize int,
	items []item) (*Map, error) {

	ts := uint64(tableSize)
	nbuckets := uint64(len(items)/bucketSize) + 1
	buckets := make(buckets, nbuckets)
	hashIdx := make([]int, nbuckets)

	// Calculate hashes and put them into their designated buckets
	for i := range items {
		h1, h2, h3, _ := spookyHash(items[i].key, seed[0], seed[1])

		h2 %= ts
		h3 %= ts
		hash := hash{h2, h3}

		bidx := h1 % nbuckets
		buckets[bidx].index = bidx
		buckets[bidx].hashes = append(buckets[bidx].hashes, hash)
	}

	// Sort buckets in decreasing size
	sort.Sort(buckets)

	maxHashIdx := min(tableSize*tableSize, 1<<20)
	occupied := make([]bool, tableSize)
	indices := make([]uint64, 0, len(buckets[0].hashes))

	// Process buckets and populate table
	for _, b := range buckets {
		if len(b.hashes) == 0 {
			continue
		}

		d0 := uint64(0)
		d1 := uint64(math.MaxUint64) // rolls back to 0 when 1 is added

		hidx := 0

	NextHashIdx:
		for {
			if hidx == maxHashIdx {
				return nil, errors.New("chd: can't find collission-free hash function")
			}

			d1++
			if d1 == ts {
				d0++
				d1 = 0
			}

			indices = indices[:0]
			for _, h := range b.hashes {
				idx := (h.h1 + (d0 * h.h2) + d1) % ts

				if occupied[idx] {
					// Collission has occured, clear
					// table of previously added items
					for _, n := range indices {
						occupied[n] = false
					}

					hidx++
					continue NextHashIdx
				}

				occupied[idx] = true
				indices = append(indices, idx)
			}

			hashIdx[b.index] = hidx
			break
		}
	}

	// Construct hash array
	array := newCompactArray()
	for _, idx := range hashIdx {
		array.Add(idx)
	}

	m := &Map{
		seed,
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

func min(a, b int) int {
	if a < b {
		return a
	}

	return b
}
