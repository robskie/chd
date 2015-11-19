package chd

import "testing"

func benchmarkBuild(b *testing.B, nkeys int) {
	builder := NewBuilder()
	for i := 0; i < nkeys; i++ {
		builder.Add(encode(i))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		builder.Build(nil)
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
