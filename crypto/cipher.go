package crypto

import (
	"fmt"

	"github.com/superp00t/gophercraft/crypto/arc4"
	"github.com/superp00t/gophercraft/vsn"
)

// Cipher describes some method for setting up an encryption layer over a TCP socket.
type Cipher interface {
	Init(server bool, key []byte) error
	Encrypt(data, tag []byte) error
	Decrypt(data, tag []byte) error
}

type DummyCipher struct {
}

func (d DummyCipher) Init(server bool, key []byte) error {
	return nil
}

func (d DummyCipher) Decrypt(data, tag []byte) error {
	return nil
}

func (d DummyCipher) Encrypt(data, tag []byte) error {
	return nil
}

func NewCipher(version vsn.Build, sessionKey []byte, server bool) (Cipher, error) {
	var c Cipher

	switch {
	case version < 4062:
		c = DummyCipher{}
		
	case version >= 4062 && version < vsn.V2_4_3:
		// Basic XOR encryption was added to protocol 4062.
		c = &arc4.CipherVanilla{}
	case version.AddedIn(vsn.V2_4_3) && version < 9614:
		// HMAC-SHA1 key added with pre-generated seed
		c = &arc4.CipherTBC{}
	case version == 12340:
		// Each build now includes two pre-generated HMAC-SHA1 seeds.
		// Encryption is now two ARC4 streams.
		c = &arc4.AuthCipher{
			ServerEncryptionKey: []byte{
				0xC2, 0xB3, 0x72, 0x3C, 0xC6, 0xAE, 0xD9, 0xB5,
				0x34, 0x3C, 0x53, 0xEE, 0x2F, 0x43, 0x67, 0xCE,
			},

			ServerDecryptionKey: []byte{
				0xCC, 0x98, 0xAE, 0x04, 0xE8, 0x97, 0xEA, 0xCA,
				0x12, 0xDD, 0xC0, 0x93, 0x42, 0x91, 0x53, 0x57,
			},
		}
	case version >= 33369:
		c = &AesCipher{}
	default:
		return nil, fmt.Errorf("no cipher exists for protocol %s", version)
	}

	if err := c.Init(server, sessionKey); err != nil {
		return nil, err
	}

	return c, nil
}
