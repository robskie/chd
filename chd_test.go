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
