package arc4

// CipherVanilla implements Vanilla-style encryption (Not ARC4, what even is this?)
// some kind of basic xor counter cipher
type CipherVanilla struct {
	key            []byte
	recv_i, recv_j uint8
	send_i, send_j uint8
}

func (v *CipherVanilla) Init(server bool, sessionKey []byte) error {
	v.key = sessionKey
	return nil
}

func (v *CipherVanilla) Encrypt(data, tag []byte) error {
	for t := 0; t < len(data); t++ {
		v.send_i %= uint8(len(v.key))
		x := (data[t] ^ v.key[v.send_i]) + v.send_j
		v.send_i++
		v.send_j = x
		data[t] = v.send_j
	}
	return nil
}

func (v *CipherVanilla) Decrypt(data, tag []byte) error {
	for t := 0; t < len(data); t++ {
		v.recv_i %= uint8(len(v.key))
		x := (data[t] - v.recv_j) ^ v.key[v.recv_i]
		v.recv_i++
		v.recv_j = data[t]
		data[t] = x
	}
	return nil
}
