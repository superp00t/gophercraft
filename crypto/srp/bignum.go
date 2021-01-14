package srp

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
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

func (x *BigNum) ToArray(ln ...int) []byte {
	buffer := reverseBuffer(x.X.Bytes())
	if len(ln) > 0 {
		length := ln[0]
		if len(buffer) < length {
			buffer = append(buffer, make([]byte, length-len(buffer))...)
		}

		if len(buffer) > length {
			buffer = buffer[:length]
		}
	}

	if len(ln) > 0 {
		if len(buffer) != ln[0] {
			panic("invalid size")
		}
	}

	return buffer
}

func (x *BigNum) String() string {
	return x.X.String()
}

func (x *BigNum) Copy() *BigNum {
	deref := *x.X
	y := &BigNum{&deref}
	return y
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

func NewBigNum() *BigNum {
	return &BigNum{new(big.Int)}
}

// Big-Endian.
func NewBigNumFromHex(hx string) *BigNum {
	bytes, err := hex.DecodeString(hx)
	if err != nil {
		panic(err)
	}
	X := &BigNum{X: new(big.Int).SetBytes(bytes)}
	return X
}

// Big-Endian.
func (x *BigNum) ToHex() string {
	return hex.EncodeToString(x.X.Bytes())
}

func (x *BigNum) ToHexLE(ln ...int) string {
	return hex.EncodeToString(x.ToArray(ln...))
}
