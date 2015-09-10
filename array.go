package chd

import (
	"unsafe"

	"github.com/robskie/fibvec"
)

// CompactArray represents a
// compressed integer array.
type CompactArray interface {
	Add(value int)
	Get(index int) int

	Len() int
	Size() int
}

type intArray []int

func newIntArray(size int) *intArray {
	a := make(intArray, 0, size)
	return &a
}

func (a *intArray) Add(value int) {
	*a = append(*a, value)
}

func (a intArray) Get(index int) int {
	return a[index]
}

func (a intArray) Len() int {
	return len(a)
}

func (a intArray) Size() int {
	sizeofInt := int(unsafe.Sizeof(int(0)))
	return len(a) * sizeofInt
}

// FibArray is a wrapper to a fibonacci
// vector that implements CompactArray
// interface.
type FibArray struct {
	*fibvec.Vector
}

// NewFibArray returns a new FibArray.
func NewFibArray() *FibArray {
	return &FibArray{fibvec.NewVector()}
}

// Add adds a new value.
func (f *FibArray) Add(value int) {
	f.Vector.Add(uint(value))
}

// Get returns the value at the given index.
func (f *FibArray) Get(index int) int {
	return int(f.Vector.Get(index))
}
