package chd

import (
	rand "github.com/remerge/go-xorshift"
)

func pickRandom(slice [][]byte) (v []byte) {
	length := len(slice)
	from := rand.Intn(length)
	v = slice[from]
	if len(v) > 0 {
		return v
	}

	for l, r := from-1, from+1; l >= 0 || r < length; l, r = l-1, r+1 {
		if r < length && len(slice[r]) > 0 {
			return slice[r]
		}
		if l >= 0 && len(slice[l]) > 0 {
			return slice[l]
		}
	}
	return nil
}
