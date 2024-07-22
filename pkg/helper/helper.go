package helper

import (
	"bytes"
	"net"
)

// HumanHexConvert converts a decimal number to its *visual* equivalent in hex: i.e. 510 -> 1296  -> which is again 510 in dex print
// we're taking in a uint32 for simplicity but actually only look at 16 bit/2byte
// and we convert decimal to hex, so we actually only can handle <10000 for now, otherwise we need more than 2 byte
func HumanHexConvert(val uint32) uint32 {
	//var res, exp10, exp16 uint32 = 0, 100000, 1048576
	// exp10 = 10^7, exp16 = 16^7 -> wr got 4 byte/ 8 digits
	var res, exp10, exp16 uint32 = 0, 10000000, 268435456

	for val > 0 {
		res += val / exp10 * exp16
		val = val % exp10
		exp10 = exp10 / 10
		exp16 = exp16 / 16
	}

	return res
}

// ItoB converts a int to a byte array for byte operations.
func ItoB(val uint32) []byte {
	r := make([]byte, (32 / 8))
	l := len(r) - 1

	for i := 0; i <= l; i++ {
		r[l-i] = byte((val >> (8 * i)) & 0xff)
	}
	return r
}

// IsMask24 checks if the given IPMask is a /24 mask.
func IsMask24(mask net.IPMask) bool {
	mask24 := []byte{255, 255, 255, 0}
	return bytes.Equal(mask, mask24)
}
