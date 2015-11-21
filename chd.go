// Package chd implements the compress, hash, and displace (CHD) minimal perfect
// hash algorithm. It provides a map builder that manages adding of items and
// map creation.
//
// See http://cmph.sourceforge.net/papers/esa09.pdf for more details.
package chd

import (
	"encoding/gob"
	"fmt"
	"io"
)

// Map represents a map that uses
// CHD minimal perfect hash algorithm.
type Map struct {
	seed [2]uint64

	length    int
	tableSize int

	hashes CompactArray
}

// NewMap returns an empty map.
// Call Map.Read to populate it.
func NewMap() *Map {
	return &Map{hashes: newIntArray(0)}
}

// Get returns the index of a given key. This will
// always return a value in the range [0, Map.Cap())
// even if the key is not found. It is up to the user
// to validate the returned index.
func (m *Map) Get(key []byte) int {
	if m.length == 0 {
		return 0
	}

	h1, h2, h3, _ := spookyHash(key, m.seed[0], m.seed[1])

	hashes := m.hashes
	hlen := uint64(hashes.Len())
	hidx := uint64(hashes.Get(int(h1 % hlen)))

	tableSize := uint64(m.tableSize)

	h2 %= tableSize
	h3 %= tableSize
	d0 := hidx / tableSize
	d1 := hidx % tableSize
	idx := int((h2 + (d0 * h3) + d1) % tableSize)

	return idx
}

// Write serializes the map.
func (m *Map) Write(w io.Writer) error {
	enc := gob.NewEncoder(w)

	enc.Encode(m.seed)
	enc.Encode(m.length)
	enc.Encode(m.tableSize)

	if err := enc.Encode(m.hashes); err != nil {
		return fmt.Errorf("chd: write failed (%v)", err)
	}

	return nil
}

// Read deserializes a map. Note that array must
// be a pointer and should have the same type as
// the one used in serializing the map. If array
// is nil, it will use a plain integer array instead.
func (m *Map) Read(r io.Reader, array CompactArray) error {
	dec := gob.NewDecoder(r)

	dec.Decode(&m.seed)
	dec.Decode(&m.length)
	dec.Decode(&m.tableSize)

	if array == nil {
		array = newIntArray(0)
	}
	m.hashes = array

	if err := dec.Decode(m.hashes); err != nil {
		return fmt.Errorf("chd: read failed (%v)", err)
	}

	return nil
}

// Len returns the total number of keys.
func (m *Map) Len() int {
	return m.length
}

// Cap returns the total number of
// required bins to store the keys.
func (m *Map) Cap() int {
	return m.tableSize
}

// Size returns the size in bytes.
func (m *Map) Size() int {
	return m.hashes.Size()
}
