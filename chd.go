// Package chd implements the compress, hash, and displace (CHD) minimal perfect
// hash algorithm. It provides a map builder that manages adding of items and
// map creation. It also provides a fibonacci array that can be used to further
// optimize memory usage.
//
// See http://cmph.sourceforge.net/papers/esa09.pdf for more details.
package chd

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"io"
	"unsafe"
)

// ErrNotFound is returned when a given key
// does not correspond to any value in the map.
var ErrNotFound = errors.New("chd: item not found")

// Map represents a map that uses
// CHD minimal perfect hash algorithm.
type Map struct {
	seed [2]uint64

	data []byte

	// For variable key
	// and/or value sizes
	sizes    []int
	offsets  []int
	keySizes []int

	// For constant key
	// and/or value sizes
	keySize  int
	itemSize int

	keySentinel []byte

	length    int
	tableSize int

	hashes CompactArray
}

// NewMap returns an empty map.
// Call Map.Read to populate it.
func NewMap() *Map {
	return &Map{hashes: newIntArray(0)}
}

// Get returns the value with the given key.
func (m *Map) Get(key []byte) ([]byte, error) {
	if m.length == 0 {
		return nil, ErrNotFound
	}

	h1, h2, h3, _ := spookyHash(key, m.seed[0], m.seed[1])
	hlen := uint64(m.hashes.Len())
	hidx := uint64(m.hashes.Get(int(h1 % hlen)))

	tableSize := uint64(m.tableSize)
	d0 := hidx / tableSize
	d1 := hidx % tableSize
	idx := (h2 + (d0 * h3) + d1) % tableSize

	size := m.itemSize
	offset := int(idx) * size
	if size == -1 {
		size = m.sizes[idx]
		offset = m.offsets[idx]
	}

	if size < 0 {
		return nil, ErrNotFound
	}
	data := m.data[offset : offset+size]

	if len(data) < len(key) {
		return nil, ErrNotFound
	}

	if !bytes.Equal(key, data[:len(key)]) {
		return nil, ErrNotFound
	}

	return data[len(key):], nil
}

// Write serializes the map.
func (m *Map) Write(w io.Writer) error {
	enc := gob.NewEncoder(w)

	enc.Encode(m.seed)
	enc.Encode(m.data)
	enc.Encode(m.sizes)
	enc.Encode(m.offsets)
	enc.Encode(m.keySizes)
	enc.Encode(m.keySize)
	enc.Encode(m.itemSize)
	enc.Encode(m.keySentinel)
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
	dec.Decode(&m.data)
	dec.Decode(&m.sizes)
	dec.Decode(&m.offsets)
	dec.Decode(&m.keySizes)
	dec.Decode(&m.keySize)
	dec.Decode(&m.itemSize)
	dec.Decode(&m.keySentinel)
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

// Iterator returns a map iterator.
func (m *Map) Iterator() *Iterator {
	it := &Iterator{m, 0, nil, nil}
	return it.Next()
}

// Len returns the number of items stored.
func (m *Map) Len() int {
	return m.length
}

// Size returns the size in bytes.
func (m *Map) Size() int {
	sizeofInt := int(unsafe.Sizeof(int(0)))

	size := len(m.data)
	size += len(m.sizes) * sizeofInt
	size += len(m.offsets) * sizeofInt
	size += len(m.keySizes) * sizeofInt
	size += m.hashes.Size()

	return size
}

// Iterator represents a map iterator.
type Iterator struct {
	m     *Map
	index int

	key   []byte
	value []byte
}

// Key returns the current key
// pointed to by the iterator.
func (i *Iterator) Key() []byte {
	return i.key
}

// Value returns the current value
// pointed to by the iterator.
func (i *Iterator) Value() []byte {
	return i.value
}

// Next returns the next map iterator.
// Returns nil, if there are no more
// key-value pairs to traverse.
func (i *Iterator) Next() *Iterator {
	m := i.m

	key := []byte{}
	value := []byte{}

	var idx int
	for idx = i.index; idx < m.tableSize; idx++ {
		keySize := m.keySize
		if keySize == -1 {
			keySize = m.keySizes[idx]
		}

		size := m.itemSize
		offset := idx * size
		if size == -1 {
			size = m.sizes[idx]
			offset = m.offsets[idx]
		}

		if size < 0 {
			continue
		}

		data := m.data[offset : offset+size]
		if !bytes.Equal(m.keySentinel, data[:keySize]) {
			key = data[:keySize]

			if len(data) > len(key) {
				value = data[keySize:]
			}

			break
		}
	}

	if idx < m.tableSize {
		i.index = idx + 1
		i.key = key
		i.value = value

		return i
	}

	return nil
}
