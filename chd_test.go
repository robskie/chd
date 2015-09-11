package chd

import (
	"bytes"
	"encoding/binary"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMapGetEmpty(t *testing.T) {
	m := NewMap()
	assert.Equal(t, 0, m.Len())

	_, err := m.Get([]byte{0})
	assert.NotNil(t, err)
}

func TestMapNilKey(t *testing.T) {
	b := NewBuilder()
	v := []byte{1, 2, 3}

	b.Add(nil, v)
	m := b.Build(nil)
	assert.Equal(t, 1, m.Len())

	vv, err := m.Get(nil)
	assert.Nil(t, err)
	assert.Equal(t, v, vv)
}

func TestMapGetOneEntry(t *testing.T) {
	b := NewBuilder()
	k, v := []byte{0}, []byte{1}
	b.Add(k, v)

	m := b.Build(nil)
	assert.Equal(t, 1, m.Len())

	vv, err := m.Get(k)
	assert.Nil(t, err)
	assert.Equal(t, v, vv)
}

func TestMapGetNilValues(t *testing.T) {
	b := NewBuilder()
	for i := 0; i < 1e5; i++ {
		d := encode(i)
		b.Add(d, nil)
	}

	m := b.Build(nil)
	for i := 0; i < 1e5; i++ {
		k := encode(i)
		v, err := m.Get(k)

		if !assert.Nil(t, err) {
			break
		}

		if !assert.Empty(t, v) {
			break
		}
	}

	assert.EqualValues(t, 1e5, m.Len())
}

func TestMapGetFixedKeySize(t *testing.T) {
	b := NewBuilder()
	values := make([][]byte, 1e5)
	for i := 0; i < 1e5; i++ {
		k := encode(i)
		v := randBytes(rand.Intn(10))

		values[i] = v
		b.Add(k, v)
	}

	m := b.Build(nil)
	assert.NotEqual(t, -1, m.keySize)

	for i := 0; i < 1e5; i++ {
		k := encode(i)
		v, err := m.Get(k)

		if !assert.Nil(t, err) {
			break
		}

		if !assert.Equal(t, values[i], v) {
			break
		}
	}

	assert.EqualValues(t, 1e5, m.Len())
}

func TestMapGetFixedKeyValueSize(t *testing.T) {
	b := NewBuilder()
	for i := 0; i < 1e5; i++ {
		d := encode(i)
		b.Add(d, d)
	}

	m := b.Build(nil)
	assert.NotEqual(t, -1, m.keySize)
	assert.NotEqual(t, -1, m.itemSize)

	for i := 0; i < 1e5; i++ {
		k := encode(i)
		v, err := m.Get(k)

		if !assert.Nil(t, err) {
			break
		}

		if !assert.Equal(t, i, decode(v)) {
			break
		}
	}

	assert.EqualValues(t, 1e5, m.Len())
}

func TestMapGetRandKeyValueSize(t *testing.T) {
	b := NewBuilder()

	kv := map[string][]byte{}
	for i := 0; i < 1e5; i++ {
		k := randBytes(rand.Intn(10) + 1)
		v := randBytes(rand.Intn(10))

		kv[string(k)] = v
		b.Add(k, v)
	}

	m := b.Build(nil)
	assert.Equal(t, -1, m.keySize)
	assert.Equal(t, -1, m.itemSize)

	for k, v := range kv {
		vv, err := m.Get([]byte(k))

		if !assert.Nil(t, err) {
			break
		}

		if !assert.Equal(t, v, vv) {
			break
		}
	}

	assert.EqualValues(t, len(kv), m.Len())
}

func TestMapGetHitMiss(t *testing.T) {
	b := NewBuilder()

	kv := map[string][]byte{}
	for i := 0; i < 1e5; i++ {
		k := randBytes(rand.Intn(3) + 1)
		v := randBytes(rand.Intn(10))

		kv[string(k)] = v
		b.Add(k, v)
	}

	m := b.Build(nil)
	for i := 0; i < 1e6; i++ {
		k := randBytes(rand.Intn(3) + 1)

		v, ok := kv[string(k)]
		vv, err := m.Get(k)
		if ok {
			if !assert.Nil(t, err) {
				break
			}

			if !assert.Equal(t, v, vv) {
				break
			}
		} else {
			if !assert.NotNil(t, err) {
				break
			}
		}
	}

	assert.EqualValues(t, len(kv), m.Len())
}

func TestMapWriteRead(t *testing.T) {
	b := NewBuilder()
	values := make([][]byte, 1e5)
	for i := 0; i < 1e5; i++ {
		k := encode(i)
		v := randBytes(10)

		values[i] = v
		b.Add(k, v)
	}
	m := b.Build(nil)

	buf := &bytes.Buffer{}
	err := m.Write(buf)
	assert.Nil(t, err)

	mm := NewMap()
	err = mm.Read(buf, nil)
	assert.Nil(t, err)

	for i := 0; i < 1e5; i++ {
		k := encode(i)
		v, err := mm.Get(k)

		if !assert.Nil(t, err) {
			break
		}

		if !assert.Equal(t, values[i], v) {
			break
		}
	}

	assert.EqualValues(t, 1e5, mm.Len())
}

func TestIteratorEmpty(t *testing.T) {
	it := NewBuilder().Build(nil).Iterator()
	assert.Nil(t, it)

	it = NewMap().Iterator()
	assert.Nil(t, it)
}

func TestIteratorOneEntry(t *testing.T) {
	b := NewBuilder()
	k, v := []byte{0}, []byte{1}

	b.Add(k, v)
	it := b.Build(nil).Iterator()

	assert.NotNil(t, it)
	assert.Equal(t, k, it.Key())
	assert.Equal(t, v, it.Value())
	assert.Nil(t, it.Next())
}

func TestIteratorFixedKeySize(t *testing.T) {
	b := NewBuilder()
	values := make([][]byte, 1e5)
	for i := 0; i < 1e5; i++ {
		k := encode(i)
		v := randBytes(rand.Intn(10))

		values[i] = v
		b.Add(k, v)
	}
	m := b.Build(nil)

	n := 0
	for it := m.Iterator(); it != nil; it = it.Next() {
		k := decode(it.Key())
		v := it.Value()

		if !assert.Equal(t, values[k], v) {
			break
		}
		n++
	}

	assert.EqualValues(t, 1e5, n)
}

func TestIteratorFixedKeyValueSize(t *testing.T) {
	b := NewBuilder()
	values := make([]int, 1e5)
	for i := 0; i < 1e5; i++ {
		k := encode(i)
		values[i] = i
		b.Add(k, k)
	}
	m := b.Build(nil)

	n := 0
	for it := m.Iterator(); it != nil; it = it.Next() {
		k := decode(it.Key())
		v := decode(it.Value())

		if !assert.Equal(t, values[k], v) {
			break
		}
		n++
	}

	assert.EqualValues(t, 1e5, n)
}

func TestIteratorRandKeyValueSize(t *testing.T) {
	b := NewBuilder()

	kv := map[string][]byte{}
	for i := 0; i < 1e5; i++ {
		k := randBytes(rand.Intn(10) + 1)
		v := randBytes(rand.Intn(10))

		kv[string(k)] = v
		b.Add(k, v)
	}
	m := b.Build(nil)

	n := 0
	for it := m.Iterator(); it != nil; it = it.Next() {
		k := it.Key()
		v := it.Value()

		if !assert.Equal(t, kv[string(k)], v) {
			break
		}
		n++
	}

	assert.Equal(t, len(kv), n)
}

// BenchmarkMapGet measures the average running
// time of Map.Get operations using an IntArray.
func BenchmarkMapGetIntArray(b *testing.B) {
	builder := NewBuilder()
	keys := make([][]byte, 1e5)

	for i := range keys {
		k := randBytes(rand.Intn(8) + 1)

		keys[i] = k
		builder.Add(k, nil)
	}
	m := builder.Build(nil)

	kidx := make([]int, b.N)
	for i := range kidx {
		kidx[i] = rand.Intn(len(keys))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Get(keys[kidx[i]])
	}
}

// BenchmarkMapGet measures the average running
// time of Map.Get operations using a FibArray.
func BenchmarkMapGetFibArray(b *testing.B) {
	builder := NewBuilder()
	keys := make([][]byte, 1e5)

	for i := range keys {
		k := randBytes(rand.Intn(8) + 1)

		keys[i] = k
		builder.Add(k, nil)
	}
	m := builder.Build(NewFibArray())

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
