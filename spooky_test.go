package chd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSpookyHash(t *testing.T) {
	type testcase struct {
		in  string
		out []uint64
	}

	// These are generated from the original C++ implementation
	testcases := []testcase{
		{"", []uint64{0x232706fc6bf50919, 0x8b72ee65b4e851c7, 0x88d8e9628fb694ae, 0x15c99660e766a98}},
		{"0", []uint64{0x50209687d54ec67e, 0x62fe85108df1cf6d, 0xe293ccf8bc18048f, 0xdfaa1b32797c62c6}},
		{"01", []uint64{0xfbe67d8368f3fb4f, 0xb54a5a89706d5a5a, 0x97a6a7de4bc93c0d, 0x7584d4a5dc92928e}},
		{"012", []uint64{0x2882d11a5846ccfa, 0x6b21b0e870109222, 0x53b76f081be71d6d, 0x827b586f534e81f9}},
		{"0123", []uint64{0xf5e0d56325d6d000, 0xaf8703c9f9ac75e5, 0xd1065083f59621a3, 0x30ead776f0ad91fc}},
		{"01234", []uint64{0x59a0f67b7ae7a5ad, 0x84d7aeabc053b848, 0x5179bd6873944d59, 0x12552182789dea54}},
		{"012345", []uint64{0xf01562a268e42c21, 0xdfe994ab22873e7e, 0xaa5a78a7760472fb, 0xa4013e44eaa8e7c}},
		{"0123456", []uint64{0x16133104620725dd, 0xa5ca36afa7182e6a, 0x1f01a740882ab623, 0xf5934148ef50b1ce}},
		{"01234567", []uint64{0x7a9378dcdf599479, 0x30f5a569a74ecdd7, 0xf0302d35d5f34c53, 0x609e4f7e56f76415}},
		{"012345678", []uint64{0xd9f07bdc76c20a78, 0x34f0621847f7888a, 0x64f48fc304b97973, 0xbd4d97b2ee93f109}},
		{"0123456789", []uint64{0x332a4fff07df83da, 0xfa40557cc0ea6b72, 0xcac113f2ba22daa, 0x37c13a177afe4a20}},
		{"Stay hungry, stay foolish. -Steve Jobs", []uint64{0x1ec0bf715ba9f074, 0x9f03fb7e653c5588, 0x6a543ccaaefe3eb5, 0xf449c26636c6f191}},
		{"If you can dream it, you can do it. -Walt Disney", []uint64{0xceba60735f5169f5, 0x7a58d45081c0f7c6, 0x1260a8414a428a10, 0xc2255016ba31b401}},
		{"If at first you don't succeed; call it version 1.0", []uint64{0x59289d361150cea3, 0x4859eeeeb64fe321, 0x236501d233edf81b, 0x916185422abbed09}},
		{"Limits, like fear, is often an illusion. -Michael Jordan", []uint64{0x1bcf7c4e801d0b2d, 0xafe6037456412bf0, 0x773c642238c5e853, 0x582329445f4951e}},
		{"If you can't make it good, at least make it look good. -Bill Gates", []uint64{0x8e5ef4d5b1bcbc33, 0x82b902ab5a002f07, 0x9525e86058cf3453, 0x6a87d4d752ef9478}},
		{"Better than a thousand hollow words, is one word that brings peace. -Buddha", []uint64{0xc3b481cc388762ad, 0x6eec37340ce9e037, 0x49c08099c52a4b4d, 0x2cd33c84a25da205}},
		{"I'm generally a very pragmatic person: that which works, works. -Linus Torvalds", []uint64{0x34802203ac4dc77b, 0xc09d272f1f6f1618, 0xf707eb3034ce0577, 0x80db94a35397c26d}},
		{"We cannot solve our problems with the same thinking we used when we created them. -Albert Einstein", []uint64{0xd831ea0c04101d5b, 0x953807301e8d1cec, 0x68356e6fd10b7cf9, 0x68a2f54bd090a91a}},
	}

	for _, tc := range testcases {
		h1, h2, h3, h4 := spookyHash([]byte(tc.in), 0, 0)
		if !assert.Equal(t, tc.out, []uint64{h1, h2, h3, h4}) {
			break
		}
	}
}
