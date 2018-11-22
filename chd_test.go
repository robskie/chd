package chd

import (
	"bytes"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMap_WriteTo(t *testing.T) {
	b := NewBuilder(nil)
	for i := 0; i < 10; i++ {
		b.Add([]byte(strconv.Itoa(i)), []byte(strconv.Itoa(i)))
	}
	c, err := b.Build()
	assert.NoError(t, err)

	assert.Equal(t, "3", string(c.Get([]byte("3"))))

	var buf bytes.Buffer
	_, err = c.WriteTo(&buf)
	assert.NoError(t, err)

	c1 := NewMap()
	c1.Read(buf.Bytes())

	assert.Equal(t, "3", string(c1.Get([]byte("3"))))
}

func TestMap_Get(t *testing.T) {
	b := NewBuilder(nil)
	for i := 0; i < 4; i++ {
		b.Add([]byte(strconv.Itoa(i)), []byte(strconv.Itoa(i)))
	}
	c, err := b.Build()
	assert.NoError(t, err)

	t.Run(`found`, func(t *testing.T) {
		assert.NotNil(t, c.Get([]byte("0")))
		assert.NotNil(t, c.Get([]byte("3")))
	})
	t.Run(`not found`, func(t *testing.T) {
		assert.Nil(t, c.Get([]byte("10")))
	})
	t.Run(`random key`, func(t *testing.T) {
		assert.NotNil(t, c.GetRandomKey())
	})
	t.Run(`random value`, func(t *testing.T) {
		assert.NotNil(t, c.GetRandomValue())
	})
}

func TestMap_GetRandomKey(t *testing.T) {
	iterations := 1000
	for it := 1; it < iterations; it++ {
		t.Run(strconv.Itoa(it), func(t *testing.T) {
			b := NewBuilder(nil)
			for i := 0; i < it; i++ {
				b.Add([]byte("k"+strconv.Itoa(i)), []byte("v"+strconv.Itoa(i)))
			}
			c, err := b.Build()
			assert.NoError(t, err)

			vals := map[string]struct{}{}
			for i := 0; i < 10000; i++ {
				val := c.GetRandomKey()
				assert.NotEmpty(t, val)
				vals[string(val)] = struct{}{}
			}
			if it == 1 {
				assert.Len(t, vals, 1)
				return
			}
			assert.True(t, len(vals) > 1)
		})
	}

}

func TestMap_GetRandomValue(t *testing.T) {
	iterations := 1000

	for it := 1; it < iterations; it++ {
		t.Run(strconv.Itoa(it), func(t *testing.T) {
			b := NewBuilder(nil)
			for i := 0; i < it; i++ {
				b.Add([]byte("k"+strconv.Itoa(i)), []byte("v"+strconv.Itoa(i)))
			}
			c, err := b.Build()
			assert.NoError(t, err)

			vals := map[string]struct{}{}
			for i := 0; i < 10000; i++ {
				val := c.GetRandomValue()
				assert.NotEmpty(t, val)
				vals[string(val)] = struct{}{}
			}
			if it == 1 {
				assert.Len(t, vals, 1)
				return
			}
			assert.True(t, len(vals) > 1)
		})
	}

}
