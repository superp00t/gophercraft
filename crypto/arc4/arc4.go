// Package arc4 provides multiple implementations of RC4-based socket encryption methods.
// WARNING: RC4 is widely deemed as UNSAFE by cryptography specialists.
// We only are providing this package for backward compatibility with existing implementations.
package arc4

type ARC4 struct {
	S    []byte
	i, j byte
}

func New(key []byte) *ARC4 {
	this := &ARC4{}
	this.S = make([]byte, 256)
	for i := 0; i < 256; i++ {
		this.S[i] = byte(i)
	}

	var j uint8 = 0
	var t uint8 = 0
	for i := 0; i < 256; i++ {
		j = (j + this.S[i] + key[i%len(key)]) & 255
		t = this.S[i]
		this.S[i] = this.S[j]
		this.S[j] = t
	}

	this.i = 0
	this.j = 0
	return this
}

func (this *ARC4) Next() uint8 {
	var t uint8
	this.i = (this.i + 1) & 255
	this.j = (this.j + this.S[this.i]) & 255
	t = this.S[this.i]
	this.S[this.i] = this.S[this.j]
	this.S[this.j] = t
	return this.S[(t+this.S[this.i])&255]
}

func (this *ARC4) Encrypt(data []byte) {
	for i := 0; i < len(data); i++ {
		data[i] ^= this.Next()
	}
}

func (this *ARC4) Decrypt(data []byte) {
	this.Encrypt(data)
}
