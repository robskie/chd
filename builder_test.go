package chd

import (
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

func benchmarkBuild(b *testing.B, nkeys int) {
	builder := NewBuilder(nil)
	for i := 0; i < nkeys; i++ {
		builder.Add(encode(i))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		builder.Build()
	}
}

func BenchmarkBuild10KKeys(b *testing.B) {
	benchmarkBuild(b, 1e4)
}

func BenchmarkBuild100KKeys(b *testing.B) {
	benchmarkBuild(b, 1e5)
}

func BenchmarkBuild1MKeys(b *testing.B) {
	benchmarkBuild(b, 1e6)
}
