// Package chd implements the compress, hash, and displace (CHD) minimal perfect
// hash algorithm. It provides a map builder that manages adding of items and
// map creation.
//
// See http://cmph.sourceforge.net/papers/esa09.pdf for more details.
package chd

import (
	"bytes"
	"encoding/binary"
	"io"
)

// Map represents a map that uses
// CHD minimal perfect hash algorithm.
type Map struct {
	seed      [2]uint64
	length    uint64
	tableSize uint64
	indexLen  uint64
	index     []uint64
	keys      [][]byte
	values    [][]byte
}

// NewMap returns an empty map.
// Call Map.Read to populate it.
func NewMap() *Map {
	return &Map{}
}

func (m *Map) Get(key []byte) []byte {
	if m.length == 0 {
		return nil
	}
	idx := m.getIndex(key)
	if bytes.Compare(key, m.keys[idx]) != 0 {
		return nil
	}
	return m.values[idx]
}

// Get a random entry from the hash table
func (m *Map) GetRandomValue() []byte {
	if m.length == 0 || len(m.values) == 0 {
		return nil
	}

	return pickRandom(m.values)
}

// Get a random entry from the hash table
func (m *Map) GetRandomKey() []byte {
	if m.length == 0 || len(m.keys) == 0 {
		return nil
	}
	return pickRandom(m.keys)
}

func (m *Map) getIndex(key []byte) (idx uint64) {
	if m.length == 0 {
		return 0
	}

	h1, h2, h3, _ := spookyHash(key, m.seed[0], m.seed[1])

	hlen := m.indexLen
	hidx := m.index[int(h1%hlen)]

	// tableSize :=
	h2 %= m.tableSize
	h3 %= m.tableSize
	d0 := hidx / m.tableSize
	d1 := hidx % m.tableSize
	idx = (h2 + (d0 * h3) + d1) % m.tableSize
	return
}

// Len returns the total number of keys.
func (m *Map) Len() int {
	return int(m.length)
}

// Size returns the size in bytes.
func (m *Map) Size() int {
	return int(m.length)
}

func (m *Map) Read(p []byte) (n int, err error) {
	bi := &sliceReader{b: p}
	m.seed[0] = bi.ReadInt()
	m.seed[1] = bi.ReadInt()
	m.tableSize = bi.ReadInt()
	m.indexLen = bi.ReadInt()
	m.length = bi.ReadInt()

	// index
	m.index = make([]uint64, m.length)

	var i uint64
	for i = 0; i < m.indexLen; i++ {
		m.index[i] = bi.ReadInt()
	}

	var kl, vl uint64
	m.keys = make([][]byte, m.tableSize)
	m.values = make([][]byte, m.tableSize)
	for i = 0; i < m.tableSize; i++ {
		kl = bi.ReadInt()
		vl = bi.ReadInt()
		m.keys[i] = bi.Read(kl)
		m.values[i] = bi.Read(vl)
	}
	return
}

// Serialize to given Writer
func (m *Map) WriteTo(w io.Writer) (n int64, err error) {
	write := func(nd ...interface{}) (n int64, err error) {
		for _, d := range nd {
			if err = binary.Write(w, binary.LittleEndian, d); err != nil {
				return
			}
			n += int64(binary.Size(d))
		}
		return
	}

	n, err = write(
		m.seed[0],
		m.seed[1],
		m.tableSize,
		m.indexLen,
		uint64(m.length),
	)
	var n1 int64
	for i := range m.index {
		if n1, err = write(m.index[i]); err != nil {
			return
		}
		n += n1
	}

	var ni int
	var i uint64
	for i = 0; i < m.tableSize; i++ {
		if n1, err = write(uint64(len(m.keys[i]))); err != nil {
			return
		}
		n += n1
		if n1, err = write(uint64(len(m.values[i]))); err != nil {
			return
		}
		n += n1
		if ni, err = w.Write(m.keys[i]); err != nil {
			return
		}
		n += int64(ni)
		if ni, err = w.Write(m.values[i]); err != nil {
			return
		}
		n += int64(ni)
	}
	return
}
