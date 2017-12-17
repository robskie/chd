package chd

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"strconv"
	"fmt"
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

func TestBuild_Incr1024(t *testing.T) {
	check := func(n int) (err error) {
		builder := NewBuilder(nil)
		for i:=0; i<n+1; i++ {
			builder.Add([]byte(strconv.Itoa(i)), []byte(strconv.Itoa(i+10)))
		}
		c, err := builder.Build()
		if err != nil {
			return
		}
		for i:=0; i<n+1; i++ {
			idx := c.Get([]byte(strconv.Itoa(i)))
			if strconv.Itoa(i+10) != string(idx) {
				err = fmt.Errorf("i=%d: %s != %s", i, strconv.Itoa(i+10), string(idx))
				return
			}
		}
		return
	}
	for i:=10;i<11;i++ {
		assert.NoError(t, check(i))
	}
}

//func benchmarkBuild(b *testing.B, nkeys int) {
//	builder := NewBuilder(nil)
//	for i := 0; i < nkeys; i++ {
//		builder.Add(encode(i))
//	}
//
//	b.ResetTimer()
//	for i := 0; i < b.N; i++ {
//		builder.Build()
//	}
//}
//
//func BenchmarkBuild10KKeys(b *testing.B) {
//	benchmarkBuild(b, 1e4)
//}
//
//func BenchmarkBuild100KKeys(b *testing.B) {
//	benchmarkBuild(b, 1e5)
//}
//
//func BenchmarkBuild1MKeys(b *testing.B) {
//	benchmarkBuild(b, 1e6)
//}
