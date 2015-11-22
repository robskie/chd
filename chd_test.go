package chd

import (
	"bytes"
	"encoding/binary"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMapGet(t *testing.T) {
	b := NewBuilder(nil)

	keys := [][]byte{}
	for i := 0; i < 1e4; i++ {
		k := encode(i)

		b.Add(k)
		keys = append(keys, k)
	}

	m, _ := b.Build()
	occupied := make([]bool, m.Cap())
	for _, k := range keys {
		idx := m.Get(k)

		// Key index must be unique for every key
		if !assert.False(t, occupied[idx]) {
			break
		}
		occupied[idx] = true
	}

	assert.EqualValues(t, len(keys), m.Len())
}

func TestMapDelete(t *testing.T) {
	b := NewBuilder(nil)

	// Delete items that weren't added
	for i := 0; i < 100; i++ {
		b.Delete(encode(i))
	}

	// Add items
	keys := map[string]struct{}{}
	for i := 0; i < 1e4; i++ {
		k := encode(i)

		b.Add(k)
		keys[string(k)] = struct{}{}
	}

	// Delete added items
	i := 0
	reinsert := map[string]struct{}{}
	for k := range keys {
		delete(keys, k)
		b.Delete([]byte(k))

		reinsert[k] = struct{}{}

		if i == 1000 {
			break
		}
		i++
	}

	// Reinsert some deleted items
	i = 0
	for k := range reinsert {
		keys[k] = struct{}{}
		b.Add([]byte(k))

		if i == 500 {
			break
		}
		i++
	}

	m, _ := b.Build()
	occupied := make([]bool, m.Cap())
	for k := range keys {
		idx := m.Get([]byte(k))

		// Key index must be unique for every key
		if !assert.False(t, occupied[idx]) {
			break
		}
		occupied[idx] = true
	}

	assert.EqualValues(t, len(keys), m.Len())
}

func TestMapWriteRead(t *testing.T) {
	b := NewBuilder(nil)

	keys := map[string]int{}
	for i := 0; i < 1e4; i++ {
		k := encode(i)

		keys[string(k)] = -1
		b.Add(k)
	}

	m, _ := b.Build()
	for k := range keys {
		keys[k] = m.Get([]byte(k))
	}

	buf := &bytes.Buffer{}
	err := m.Write(buf)
	assert.Nil(t, err)

	mm := NewMap()
	err = mm.Read(buf)
	assert.Nil(t, err)

	for k, i := range keys {
		if !assert.Equal(t, i, mm.Get([]byte(k))) {
			break
		}
	}

	assert.EqualValues(t, m.Len(), mm.Len())
}

// BenchmarkMapGet measures the average running
// time of Map.Get operations using an IntArray.
func BenchmarkMapGet100KKeys(b *testing.B) {
	builder := NewBuilder(nil)
	keys := make([][]byte, 1e5)

	for i := range keys {
		k := encode(i)

		keys[i] = k
		builder.Add(k)
	}
	m, _ := builder.Build()

	kidx := make([]int, b.N)
	for i := range kidx {
		kidx[i] = rand.Intn(len(keys))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Get(keys[kidx[i]])
	}
}

func encode(v int) []byte {
	buffer := make([]byte, 8)
	binary.LittleEndian.PutUint64(buffer, uint64(v))
	return buffer
}

func decode(b []byte) int {
	return int(binary.LittleEndian.Uint64(b))
}
