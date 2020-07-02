//Package srp implements a backward-compatible version of SRP-6.
//Warning: this package is ONLY suitable for Gophercraft.
//Do not use it in any other case: it provides little in the way of security.

package srp

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"strings"
)

var (
	Generator  = BigNumFromInt(7)
	Multiplier = BigNumFromInt(3)
	Prime      = NewBigNumFromHex("894B645E89E1535BBDAD5B8B290650530801B18EBFBF5E8FAB3C82872A3E9BB7")
)

func Credentials(username, password string) []byte {
	I := strings.ToUpper(username)
	P := strings.ToUpper(password)
	return []byte(I + ":" + P)
}

func hash(input ...[]byte) []byte {
	bt := sha1.Sum(bytes.Join(input, nil))
	return bt[:]
}

// Ngh = XOR(H(N), H(g))
func HashPrimeAndGenerator(N, g *BigNum) []byte {
	Nh := hash(N.ToArray())
	gh := hash(g.ToArray())

	Ngh := make([]byte, 20)
	for i := 0; i < 20; i++ {
		Ngh[i] = Nh[i] ^ gh[i]
	}

	return Ngh
}

// Compute auth := H('username' + ':' + 'pass')
// g := 7
// ....
// x := H(salt, auth)
// v := (g^x) % N
//
func CalculateVerifier(auth []byte, g, N, salt *BigNum) (x *BigNum, v *BigNum) {
	fmt.Println("calculating g=", g.ToHex(), "N=", N.ToHex())
	x = BigNumFromArray(hash(salt.ToArray(), auth))
	v = g.ModExp(x, N)
	fmt.Println("done calculating")
	return x, v
}

func ServerGenerateEphemeralValues(g, N, v *BigNum) (b *BigNum, B *BigNum) {
	b = BigNumFromRand(19)
	gMod := g.ModExp(b, N)
	B = ((v.Multiply(Multiplier.Copy())).Add(gMod)).Mod(N)
	return
}

func SRPCalculate(username, password string, _B, n, salt []byte) (*BigNum, []byte, []byte, []byte) {
	auth := HashCredentials(username, password)
	return HashCalculate(username, auth, _B, n, salt)
}

func HashCalculate(username string, auth, _B, _N, salt []byte) (*BigNum, []byte, []byte, []byte) {
	g := Generator.Copy()
	k := Multiplier.Copy()

	N := BigNumFromArray(_N)
	s := BigNumFromArray(salt)

	fmt.Println("Calculating verifier")

	x, v := CalculateVerifier(auth, g, N, s)
	fmt.Println("Calculatiedverifier")

	a := BigNumFromRand(19)
	A := g.ModExp(a, N)

	B := BigNumFromArray(_B)

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

	Ngh := HashPrimeAndGenerator(N, g)

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

func ServerLogonProof(username string, A, M1, b, B, s, N, v *BigNum) ([]byte, bool, []byte) {
	g := Generator.Copy()

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

	M2 := hash(A.ToArray(), final, K.ToArray())
	return vK, bytes.Equal(final, M1.ToArray()), M2
}
