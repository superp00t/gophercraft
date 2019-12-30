package srp

import (
	"bytes"
	"crypto/rand"
	"math/big"
)

func reverseBuffer(input []byte) []byte {
	buf := make([]byte, len(input))
	inc := 0
	for x := len(input) - 1; x > -1; x-- {
		buf[inc] = input[x]
		inc++
	}
	return buf
}

type BigNum struct {
	X *big.Int
}

func (x *BigNum) ModExp(y, m *BigNum) *BigNum {
	return &BigNum{
		X: new(big.Int).Exp(x.X, y.X, m.X),
	}
}

func (x *BigNum) Add(y *BigNum) *BigNum {
	return &BigNum{new(big.Int).Add(x.X, y.X)}
}

func (x *BigNum) Subtract(y *BigNum) *BigNum {
	return &BigNum{new(big.Int).Sub(x.X, y.X)}
}

func (x *BigNum) Multiply(y *BigNum) *BigNum {
	return &BigNum{new(big.Int).Mul(x.X, y.X)}
}

func (x *BigNum) Divide(y *BigNum) *BigNum {
	return &BigNum{new(big.Int).Div(x.X, y.X)}
}

func (x *BigNum) Equals(y *BigNum) bool {
	return bytes.Equal(x.X.Bytes(), y.X.Bytes())
}

func (x *BigNum) Mod(y *BigNum) *BigNum {
	return &BigNum{new(big.Int).Mod(x.X, y.X)}
}

func (x *BigNum) ToArray() []byte {
	return reverseBuffer(x.X.Bytes())
}

func BigNumFromArray(arr []byte) *BigNum {
	bigb := reverseBuffer(arr)
	bn := &BigNum{}
	bn.X = new(big.Int).SetBytes(bigb)
	return bn
}

func BigNumFromRand(l int) *BigNum {
	bigb := make([]byte, l)
	rand.Read(bigb)
	bn := &BigNum{}
	bn.X = new(big.Int).SetBytes(bigb)
	return bn
}

func BigNumFromInt(i int64) *BigNum {
	return &BigNum{big.NewInt(i)}
}
