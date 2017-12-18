package chd

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildEmpty(t *testing.T) {
	defer func() {
		assert.Nil(t, recover())
	}()

	builder := NewBuilder(nil)
	m, err := builder.Build()
	assert.NotNil(t, m)
	assert.Nil(t, err)
}

func TestBuild_Incr2048(t *testing.T) {
	check := func(n int) (err error) {
		builder := NewBuilder(nil)
		for i := 0; i < n+1; i++ {
			builder.Add([]byte(strconv.Itoa(i)), []byte(strconv.Itoa(i+10)))
		}
		c, err := builder.Build()
		if err != nil {
			return
		}
		for i := 0; i < n+1; i++ {
			idx := c.Get([]byte(strconv.Itoa(i)))
			if strconv.Itoa(i+10) != string(idx) {
				err = fmt.Errorf("i=%d: %s != %s", i, strconv.Itoa(i+10), string(idx))
				return
			}
		}
		return
	}
	for i := 0; i < 2048; i++ {
		assert.NoError(t, check(i))
	}
}
