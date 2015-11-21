package chd

import (
	"reflect"
	"unsafe"
)

// CompactArray represents a
// compressed integer array.
type CompactArray interface {
	Add(value int)
	Get(index int) int

	Len() int
	Size() int
}

// array is the CompactArray used
// to store hash indices. If this
// is nil, then an integer array is
// used.
var arrayType reflect.Type

// SetCompactArray sets the compressed
// array used to store hash indices. It
// is important that the compact array
// used when building and reading a map
// is exactly the same.
func SetCompactArray(a CompactArray) {
	arrayType = indirect(reflect.TypeOf(a))
}

// newCompactArray returns a new instance
// of CompactArray with type arrayType.
func newCompactArray() CompactArray {
	if arrayType == nil {
		return newIntArray(0)
	}

	va := reflect.New(arrayType)
	return va.Interface().(CompactArray)
}

func indirect(t reflect.Type) reflect.Type {
	if t.Kind() != reflect.Ptr {
		return t
	}

	return t.Elem()
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
