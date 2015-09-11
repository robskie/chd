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

type ipos struct {
	Size   int
	Offset int
}

type idata struct {
	Size    int
	Offset  int
	KeySize int
}

type itemData struct {
	KeySizes []int
	ItemPos  []ipos
	ItemData []idata

	addSize    func(idx, v int)
	addOffset  func(idx, v int)
	addKeySize func(idx, v int)

	getSize    func(idx int) int
	getOffset  func(idx int) int
	getKeySize func(idx int) int
}

func newItemData(tableSize, itemSize, keySize int) *itemData {
	data := &itemData{}

	if itemSize > 0 && keySize > 0 {
		data.addSize = func(idx, v int) {}
		data.addOffset = func(idx, v int) {}
		data.addKeySize = func(idx, v int) {}

		data.getSize = func(idx int) int { return itemSize }
		data.getOffset = func(idx int) int { return idx * itemSize }
		data.getKeySize = func(idx int) int { return keySize }

	} else if itemSize > 0 && keySize <= 0 {
		data.KeySizes = make([]int, tableSize)

		data.addSize = func(idx, v int) {}
		data.addOffset = func(idx, v int) {}
		data.addKeySize = func(idx, v int) { data.KeySizes[idx] = v }

		data.getSize = func(idx int) int { return itemSize }
		data.getOffset = func(idx int) int { return idx * itemSize }
		data.getKeySize = func(idx int) int { return data.KeySizes[idx] }

	} else if itemSize <= 0 && keySize > 0 {
		data.ItemPos = make([]ipos, tableSize)

		data.addSize = func(idx, v int) { data.ItemPos[idx].Size = v }
		data.addOffset = func(idx, v int) { data.ItemPos[idx].Offset = v }
		data.addKeySize = func(idx, v int) {}

		data.getSize = func(idx int) int { return data.ItemPos[idx].Size }
		data.getOffset = func(idx int) int { return data.ItemPos[idx].Offset }
		data.getKeySize = func(idx int) int { return keySize }

	} else { // itemSize <= 0 && keySize <= 0
		data.ItemData = make([]idata, tableSize)

		data.addSize = func(idx, v int) { data.ItemData[idx].Size = v }
		data.addOffset = func(idx, v int) { data.ItemData[idx].Offset = v }
		data.addKeySize = func(idx, v int) { data.ItemData[idx].KeySize = v }

		data.getSize = func(idx int) int { return data.ItemData[idx].Size }
		data.getOffset = func(idx int) int { return data.ItemData[idx].Offset }
		data.getKeySize = func(idx int) int { return data.ItemData[idx].KeySize }
	}

	return data
}

func (data *itemData) size() int {
	sizeofInt := int(unsafe.Sizeof(int(0)))

	size := len(data.KeySizes) * sizeofInt
	size += len(data.ItemPos) * sizeofInt * 2
	size += len(data.ItemData) * sizeofInt * 3

	return size
}

// Map represents a map that uses
// CHD minimal perfect hash algorithm.
type Map struct {
	seed [2]uint64

	data  []byte
	items *itemData

	keySize     int
	itemSize    int
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

	hashes := m.hashes
	hlen := uint64(hashes.Len())
	hidx := uint64(hashes.Get(int(h1 % hlen)))

	tableSize := uint64(m.tableSize)
	d0 := hidx / tableSize
	d1 := hidx % tableSize
	idx := int((h2 + (d0 * h3) + d1) % tableSize)

	items := m.items
	size := items.getSize(idx)
	offset := items.getOffset(idx)
	if size < 0 {
		return nil, ErrNotFound
	}

	keySize := items.getKeySize(idx)
	data := m.data[offset : offset+size]
	if !bytes.Equal(key, data[:keySize]) {
		return nil, ErrNotFound
	}

	return data[keySize:], nil
}

// Write serializes the map.
func (m *Map) Write(w io.Writer) error {
	enc := gob.NewEncoder(w)

	enc.Encode(m.seed)
	enc.Encode(m.data)
	enc.Encode(m.keySize)
	enc.Encode(m.itemSize)
	enc.Encode(m.keySentinel)
	enc.Encode(m.length)
	enc.Encode(m.tableSize)
	enc.Encode(m.items)

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
	dec.Decode(&m.keySize)
	dec.Decode(&m.itemSize)
	dec.Decode(&m.keySentinel)
	dec.Decode(&m.length)
	dec.Decode(&m.tableSize)

	m.items = newItemData(0, m.itemSize, m.keySize)
	dec.Decode(m.items)

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
	size := len(m.data)
	size += m.items.size()
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
	items := m.items

	key := []byte{}
	value := []byte{}

	var idx int
	for idx = i.index; idx < m.tableSize; idx++ {
		size := items.getSize(idx)
		offset := items.getOffset(idx)
		if size < 0 {
			continue
		}

		keySize := items.getKeySize(idx)
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
