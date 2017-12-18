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
	key   []byte
	value []byte

	// counter is used for
	// removing key duplicates
	counter int

	h1, h2, h3 uint64
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
func (b *Builder) Add(key, value []byte) {
	b.items = append(b.items, item{key, value, b.counter, 0, 0, 0})
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
	//sort.Sort(b.items)

	loadFactor := b.opts.LoadFactor
	tableSize := uint64(float64(len(b.items)) / loadFactor)
	tableSize = uint64(nearestPrime(int(tableSize)))

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

			tableSize = uint64(float64(len(b.items)) / loadFactor)
			tableSize = uint64(nearestPrime(int(tableSize)))
		} else {
			return nil, err
		}
	}
}

// build tries to create a map and
// returns an error if unsuccessful.
func (b *Builder) build(
	seed [2]uint64,
	bucketSize int,
	tableSize uint64,
	items []item) (*Map, error) {

	nbuckets := uint64(len(items)/bucketSize) + 1
	buckets := make(buckets, nbuckets)
	hashIdx := make([]uint64, nbuckets)

	// Calculate hashes and put them into their designated buckets
	var h1, h2, h3 uint64
	for i := range items {
		h1, h2, h3, _ = spookyHash(items[i].key, seed[0], seed[1])

		h2 %= tableSize
		h3 %= tableSize
		hash := hash{h2, h3}

		items[i].h1 = h1
		items[i].h2 = h2
		items[i].h3 = h3

		bidx := h1 % nbuckets
		buckets[bidx].index = bidx
		buckets[bidx].hashes = append(buckets[bidx].hashes, hash)
	}

	// Sort buckets in decreasing size
	sort.Sort(buckets)

	maxHashIdx := uint64(min(tableSize*tableSize, 1<<20))
	occupied := make([]bool, int(tableSize))
	indices := make([]uint64, 0, len(buckets[0].hashes))

	// Process buckets and populate table
	var d0, d1 uint64
	var idx uint64
	var hidx uint64
	for _, b := range buckets {
		if len(b.hashes) == 0 {
			continue
		}

		d0 = 0
		d1 = math.MaxUint64 // rolls back to 0 when 1 is added
		hidx = 0

	NextHashIdx:
		for {
			if hidx == maxHashIdx {
				return nil, errors.New("chd: can't find collission-free hash function")
			}

			d1++
			if d1 == tableSize {
				d0++
				d1 = 0
			}

			indices = indices[:0]
			for _, h := range b.hashes {
				idx = (h.h1 + (d0 * h.h2) + d1) % tableSize

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

	m := &Map{
		seed,
		uint64(len(items)),
		tableSize,
		nbuckets,
		hashIdx,
		make([][]byte, tableSize),
		make([][]byte, tableSize),
	}

	for _, item := range items {
		h2, h3 = item.h2, item.h3
		hidx = m.index[int(item.h1%nbuckets)]

		h2 %= tableSize
		h3 %= tableSize
		d0 = hidx / tableSize
		d1 = hidx % tableSize
		idx = (h2 + (d0 * h3) + d1) % tableSize
		m.keys[idx] = item.key
		m.values[idx] = item.value
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

func min(a, b uint64) uint64 {
	if a < b {
		return a
	}

	return b
}
