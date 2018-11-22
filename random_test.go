package chd

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPickRandom(t *testing.T) {
	check := func(t *testing.T, acc [][]byte) {
		t.Helper()
		expect := []byte("test")
		for i1 := 0; i1 < 10; i1++ {
			assert.Equal(t, expect, pickRandom(acc))
		}
	}

	for i := 0; i < 10; i++ {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Run("all empty", func(t *testing.T) {
				for j := 1; j < 100; j++ {
					var acc [][]byte
					for k := 0; k < j; k++ {
						acc = append(acc, []byte{})
					}
					assert.Empty(t, pickRandom(acc))
				}
			})
			t.Run("one of", func(t *testing.T) {
				for j := 1; j < 100; j++ {
					t.Run(strconv.Itoa(j), func(t *testing.T) {
						for k := 0; k < j; k++ {
							acc := make([][]byte, j)
							acc[k] = []byte("test")
							check(t, acc)
						}
					})
				}
			})
			t.Run("all except", func(t *testing.T) {
				for j := 1; j < 100; j++ {
					t.Run(strconv.Itoa(j), func(t *testing.T) {
						for k := 0; k < j; k++ {
							acc := make([][]byte, j)
							for h := 0; h < j; h++ {
								if h != k {
								}
								acc[h] = []byte("test")
							}
							check(t, acc)
						}
					})
				}
			})
		})
	}

}
