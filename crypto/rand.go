package crypto

import (
	"crypto/rand"
	"io"
	"math/big"
)

// Generate a pseudo-random integer
func RandUint32(min, max uint32) uint32 {
	if min == max {
		return min
	}

	max_, min_ := big.NewInt(int64(max)), big.NewInt(int64(min))

	range_ := new(big.Int).Sub(max_, min_)
	bi, err := rand.Int(rand.Reader, range_)
	if err != nil {
		panic(err)
	}

	return uint32(new(big.Int).Add(min_, bi).Uint64())
}

func RandBytes(ln int) []byte {
	buf := make([]byte, ln)
	if _, err := io.ReadFull(rand.Reader, buf); err != nil {
		panic(err)
	}
	return buf
}

func ReverseBytes(bytes []byte) {
	for i, j := 0, len(bytes)-1; i < j; i, j = i+1, j-1 {
		bytes[i], bytes[j] = bytes[j], bytes[i]
	}
}
