package arc4

import (
	"crypto/hmac"
	"crypto/sha1"
)

// CipherTBC implements Vanilla encryption but with an HMAC key.
type CipherTBC struct {
	key            []byte
	recv_i, recv_j uint8
	send_i, send_j uint8
}

func (v *CipherTBC) Init(server bool, sessionKey []byte) error {
	tbcSeed := []byte{0x38, 0xA7, 0x83, 0x15, 0xF8, 0x92, 0x25, 0x30, 0x71, 0x98, 0x67, 0xB1, 0x8C, 0x4, 0xE2, 0xAA}

	hm := hmac.New(sha1.New, tbcSeed)
	hm.Write(sessionKey)

	v.key = hm.Sum(nil)
	return nil
}

func (v *CipherTBC) Encrypt(data, tag []byte) error {
	for t := 0; t < len(data); t++ {
		v.send_i %= uint8(len(v.key))
		x := (data[t] ^ v.key[v.send_i]) + v.send_j
		v.send_i++
		v.send_j = x
		data[t] = v.send_j
	}
	return nil
}

func (v *CipherTBC) Decrypt(data, tag []byte) error {
	for t := 0; t < len(data); t++ {
		v.recv_i %= uint8(len(v.key))
		x := (data[t] - v.recv_j) ^ v.key[v.recv_i]
		v.recv_i++
		v.recv_j = data[t]
		data[t] = x
	}
	return nil
}
