// Package chd implements the compress, hash, and displace (CHD) minimal perfect
// hash algorithm. It provides a map builder that manages adding of items and
// map creation.
//
// See http://cmph.sourceforge.net/papers/esa09.pdf for more details.
package chd


// Map represents a map that uses
// CHD minimal perfect hash algorithm.
type Map struct {
	seed [2]uint64
	length    int
	tableSize uint64
	indexLen uint64
	index []uint64
	values [][]byte
}

// NewMap returns an empty map.
// Call Map.Read to populate it.
func NewMap() *Map {
	return &Map{}
}

// GetIndex returns the index of a given key. This will
// always return a value in the range [0, Map.Cap())
// even if the key is not found. It is up to the user
// to validate the returned index.
func (m *Map) GetIndex(key []byte) int {
	if m.length == 0 {
		return 0
	}

	h1, h2, h3, _ := spookyHash(key, m.seed[0], m.seed[1])

	hlen := m.indexLen
	hidx := m.index[int(h1 % hlen)]

	tableSize := m.tableSize
	h2 %= tableSize
	h3 %= tableSize
	d0 := hidx / tableSize
	d1 := hidx % tableSize
	idx := int((h2 + (d0 * h3) + d1) % tableSize)

	return idx
}

func (m *Map) Get(key []byte) []byte {
	return m.values[m.GetIndex(key)]
}

// Len returns the total number of keys.
func (m *Map) Len() int {
	return m.length
}

// Size returns the size in bytes.
func (m *Map) Size() int {
	return 0
}

