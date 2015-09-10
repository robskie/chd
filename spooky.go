package chd

import (
	"reflect"
	"unsafe"
)

const sc = uint64(0xdeadbeefdeadbeef)

func rot64(x uint64, k uint) uint64 {
	return (x << k) | (x >> (64 - k))
}

func shortMix(h0, h1, h2, h3 uint64) (uint64, uint64, uint64, uint64) {
	h2 = rot64(h2, 50)
	h2 += h3
	h0 ^= h2
	h3 = rot64(h3, 52)
	h3 += h0
	h1 ^= h3
	h0 = rot64(h0, 30)
	h0 += h1
	h2 ^= h0
	h1 = rot64(h1, 41)
	h1 += h2
	h3 ^= h1
	h2 = rot64(h2, 54)
	h2 += h3
	h0 ^= h2
	h3 = rot64(h3, 48)
	h3 += h0
	h1 ^= h3
	h0 = rot64(h0, 38)
	h0 += h1
	h2 ^= h0
	h1 = rot64(h1, 37)
	h1 += h2
	h3 ^= h1
	h2 = rot64(h2, 62)
	h2 += h3
	h0 ^= h2
	h3 = rot64(h3, 34)
	h3 += h0
	h1 ^= h3
	h0 = rot64(h0, 5)
	h0 += h1
	h2 ^= h0
	h1 = rot64(h1, 36)
	h1 += h2
	h3 ^= h1
	return h0, h1, h2, h3
}

func shortEnd(h0, h1, h2, h3 uint64) (uint64, uint64, uint64, uint64) {
	h3 ^= h2
	h2 = rot64(h2, 15)
	h3 += h2
	h0 ^= h3
	h3 = rot64(h3, 52)
	h0 += h3
	h1 ^= h0
	h0 = rot64(h0, 26)
	h1 += h0
	h2 ^= h1
	h1 = rot64(h1, 51)
	h2 += h1
	h3 ^= h2
	h2 = rot64(h2, 28)
	h3 += h2
	h0 ^= h3
	h3 = rot64(h3, 9)
	h0 += h3
	h1 ^= h0
	h0 = rot64(h0, 47)
	h1 += h0
	h2 ^= h1
	h1 = rot64(h1, 54)
	h2 += h1
	h3 ^= h2
	h2 = rot64(h2, 32)
	h3 += h2
	h0 ^= h3
	h3 = rot64(h3, 25)
	h0 += h3
	h1 ^= h0
	h0 = rot64(h0, 63)
	h1 += h0
	return h0, h1, h2, h3
}

// spookyHash is a port of Bon Jenkins' SpookyHash::Short but
// instead of just returning just two, it returns four uint64 values.
func spookyHash(message []byte, seed1, seed2 uint64) (uint64, uint64, uint64, uint64) {
	u8 := message

	length := len(u8)

	var u64 []uint64
	var u32 []uint32
	if length >= 8 {
		u64 = uint64SliceFromByteSlice(u8)
	}
	if length >= 4 {
		u32 = uint32SliceFromByteSlice(u8)
	}

	remainder := length & 31
	a := seed1
	b := seed2
	c := sc
	d := sc

	if length > 15 {

		// handle all complete sets of 32 bytes
		for len(u64) >= 4 {
			c += u64[0]
			d += u64[1]
			a, b, c, d = shortMix(a, b, c, d)
			a += u64[2]
			b += u64[3]
			u64 = u64[4:]
			u32 = u32[8:]
			u8 = u8[32:]
		}

		//Handle the case of 16+ remaining bytes.
		if remainder >= 16 {
			c += u64[0]
			d += u64[1]
			a, b, c, d = shortMix(a, b, c, d)
			u64 = u64[2:]
			u32 = u32[4:]
			u8 = u8[16:]
			remainder -= 16
		}
	}

	// Handle the last 0..15 bytes, and its length
	d += uint64(length) << 56
	switch remainder {
	case 15:
		d += uint64(u8[14]) << 48
		fallthrough
	case 14:
		d += uint64(u8[13]) << 40
		fallthrough
	case 13:
		d += uint64(u8[12]) << 32
		fallthrough
	case 12:
		d += uint64(u32[2])
		c += u64[0]
		break
	case 11:
		d += uint64(u8[10]) << 16
		fallthrough
	case 10:
		d += uint64(u8[9]) << 8
		fallthrough
	case 9:
		d += uint64(u8[8])
		fallthrough
	case 8:
		c += u64[0]
		break
	case 7:
		c += uint64(u8[6]) << 48
		fallthrough
	case 6:
		c += uint64(u8[5]) << 40
		fallthrough
	case 5:
		c += uint64(u8[4]) << 32
		fallthrough
	case 4:
		c += uint64(u32[0])
		break
	case 3:
		c += uint64(u8[2]) << 16
		fallthrough
	case 2:
		c += uint64(u8[1]) << 8
		fallthrough
	case 1:
		c += uint64(u8[0])
		break
	case 0:
		c += sc
		d += sc
	}

	return shortEnd(a, b, c, d)
}

func uint64SliceFromByteSlice(bytes []byte) []uint64 {
	sh := &reflect.SliceHeader{}
	sh.Cap = cap(bytes) / 8
	sh.Len = len(bytes) / 8
	sh.Data = (uintptr)(unsafe.Pointer(&bytes[0]))
	data := *(*[]uint64)(unsafe.Pointer(sh))

	return data
}

func uint32SliceFromByteSlice(bytes []byte) []uint32 {
	sh := &reflect.SliceHeader{}
	sh.Cap = cap(bytes) / 4
	sh.Len = len(bytes) / 4
	sh.Data = (uintptr)(unsafe.Pointer(&bytes[0]))
	data := *(*[]uint32)(unsafe.Pointer(sh))

	return data
}
