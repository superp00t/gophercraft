package arc4

// Cipher5875 implements Vanilla-style encryption (Not ARC4, what even is this?)
type Cipher5875 struct {
	key            []byte
	recv_i, recv_j uint8
	send_i, send_j uint8
}

func (v *Cipher5875) Init(server bool, sessionKey []byte) error {
	v.key = sessionKey
	return nil
}

func (v *Cipher5875) Encrypt(data []byte) {
	for t := 0; t < len(data); t++ {
		v.send_i %= uint8(len(v.key))
		x := (data[t] ^ v.key[v.send_i]) + v.send_j
		v.send_i++
		v.send_j = x
		data[t] = v.send_j
	}
}

func (v *Cipher5875) Decrypt(data []byte) {
	for t := 0; t < len(data); t++ {
		v.recv_i %= uint8(len(v.key))
		x := (data[t] - v.recv_j) ^ v.key[v.recv_i]
		v.recv_i++
		v.recv_j = data[t]
		data[t] = x
	}
}
