package chd

import "unsafe"

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
