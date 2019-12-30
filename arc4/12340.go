package arc4

import (
	"crypto/hmac"
	"crypto/sha1"
)

type Cipher12340 struct {
	send, recv *ARC4
}

func (c *Cipher12340) Init(server bool, sessionKey []byte) error {
	var decKey, encKey []byte

	serverEncryptionKey := []byte{
		0xC2, 0xB3, 0x72, 0x3C, 0xC6, 0xAE, 0xD9, 0xB5,
		0x34, 0x3C, 0x53, 0xEE, 0x2F, 0x43, 0x67, 0xCE,
	}

	serverDecryptionKey := []byte{
		0xCC, 0x98, 0xAE, 0x04, 0xE8, 0x97, 0xEA, 0xCA,
		0x12, 0xDD, 0xC0, 0x93, 0x42, 0x91, 0x53, 0x57,
	}

	decryptClient := hmac.New(sha1.New, serverDecryptionKey)
	encryptServer := hmac.New(sha1.New, serverEncryptionKey)
	encryptServer.Write(sessionKey)
	decryptClient.Write(sessionKey)
	encKey = encryptServer.Sum(nil)
	decKey = decryptClient.Sum(nil)

	if server {
		c.send = New(decKey)
		c.recv = New(encKey)
	} else {
		c.recv = New(decKey)
		c.send = New(encKey)
	}

	// Drop-1024 ARC4
	for i := 0; i < 1024; i++ {
		c.recv.Next()
		c.send.Next()
	}

	return nil
}

func (c *Cipher12340) Encrypt(data []byte) {
	c.send.Encrypt(data)
}

func (c *Cipher12340) Decrypt(data []byte) {
	c.recv.Decrypt(data)
}
