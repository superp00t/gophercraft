package srp

import (
	"bytes"
	"crypto/sha1"
	"strings"
)

// Warning: this package is ONLY suitable for Gophercraft.
// Do not use it in any other case: it provides no actual security.

func Credentials(username, password string) []byte {
	I := strings.ToUpper(username)
	P := strings.ToUpper(password)
	return []byte(I + ":" + P)
}

func hash(input ...[]byte) []byte {
	bt := sha1.Sum(bytes.Join(input, nil))
	return bt[:]
}

func SRPCalculate(username, password string, _B, n, salt []byte) (*BigNum, []byte, []byte, []byte) {
	auth := HashCredentials(username, password)
	return HashCalculate(username, auth, _B, n, salt)
}

func HashCalculate(username string, auth, _B, n, salt []byte) (*BigNum, []byte, []byte, []byte) {
	g := BigNumFromInt(7)
	k := BigNumFromInt(3)

	N := BigNumFromArray(n)
	s := BigNumFromArray(salt)

	a := BigNumFromRand(19)
	A := g.ModExp(a, N)

	B := BigNumFromArray(_B)

	x := BigNumFromArray(hash(s.ToArray(), auth))

	v := g.ModExp(x, N)

	uh := hash(A.ToArray(), B.ToArray())
	u := BigNumFromArray(uh)

	kgx := k.Multiply(g.ModExp(x, N))
	aux := a.Add(u.Multiply(x))

	_S := B.Subtract(kgx).ModExp(aux, N)
	S := _S.ToArray()

	if len(S) > 32 {
		S = S[:32]
	} else {
		S = append(S, make([]byte, 32-len(S))...)
	}

	S1, S2 := make([]byte, 16), make([]byte, 16)

	for i := 0; i < 16; i++ {
		S1[i] = S[i*2]
		S2[i] = S[i*2+1]
	}

	S1h := hash(S1)
	S2h := hash(S2)

	K := make([]byte, 40)

	for i := 0; i < 20; i++ {
		K[i*2] = S1h[i]
		K[i*2+1] = S2h[i]
	}

	userh := hash([]byte(strings.ToUpper(username)))

	Nh := hash(N.ToArray())
	gh := hash(g.ToArray())

	Ngh := make([]byte, 20)
	for i := 0; i < 20; i++ {
		Ngh[i] = Nh[i] ^ gh[i]
	}

	M1 := hash(
		Ngh,
		userh,
		s.ToArray(),
		A.ToArray(),
		B.ToArray(),
		K,
	)

	return v, K, A.ToArray(), M1
}

func HashCredentials(username, password string) []byte {
	return hash(Credentials(username, password))
}

// ServerCalcVS
func ServerCalcVSX(hsh []byte, N *BigNum) (*BigNum, *BigNum, *BigNum) {
	s := BigNumFromRand(32)

	x := BigNumFromArray(hash(s.ToArray(), hsh))
	v := BigNumFromInt(7).ModExp(x, N)

	return v, s, x
}

func ServerLogonProof(username string, A, M1, b, B, s, N, v *BigNum) ([]byte, bool, []byte) {
	g := BigNumFromInt(7)

	u := BigNumFromArray(hash(A.ToArray(), B.ToArray()))
	if A.Mod(N).Equals(BigNumFromInt(0)) {
		return nil, false, nil
	}

	_S := (A.Multiply(v.ModExp(u, N))).ModExp(b, N)

	S := _S.ToArray()

	S1, S2 := make([]byte, 16), make([]byte, 16)

	for i := 0; i < 16; i++ {
		if len(S) < 32 {
			return nil, false, nil
		}
		S1[i] = S[i*2]
		S2[i] = S[i*2+1]
	}

	S1h := hash(S1)
	S2h := hash(S2)

	vK := make([]byte, 40)

	for i := 0; i < 20; i++ {
		vK[i*2] = S1h[i]
		vK[i*2+1] = S2h[i]
	}

	K := BigNumFromArray(vK)

	Nh := hash(N.ToArray())
	gh := hash(g.ToArray())

	for i := 0; i < 20; i++ {
		Nh[i] ^= gh[i]
	}

	t3 := BigNumFromArray(Nh)
	t4 := hash([]byte(strings.ToUpper(username)))

	final := hash(
		t3.ToArray(),
		t4,
		s.ToArray(),
		A.ToArray(),
		B.ToArray(),
		K.ToArray(),
	)

	M3 := hash(A.ToArray(), final, K.ToArray())
	return vK, bytes.Equal(final, M1.ToArray()), M3
}
