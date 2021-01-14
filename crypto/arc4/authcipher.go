package arc4

import (
	"crypto/hmac"
	"crypto/sha1"
)

// AuthCipher creates two ARC4 states for sending and recieving.
// They are created from pre-generated seeds, computed along with the session key.
type AuthCipher struct {
	ServerEncryptionKey []byte
	ServerDecryptionKey []byte

	send, recv *ARC4
}

func (c *AuthCipher) Init(server bool, sessionKey []byte) error {
	var decKey, encKey []byte

	decryptClient := hmac.New(sha1.New, c.ServerDecryptionKey)
	encryptServer := hmac.New(sha1.New, c.ServerEncryptionKey)
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

func (c *AuthCipher) Encrypt(data, tag []byte) error {
	c.send.Encrypt(data)
	return nil
}

func (c *AuthCipher) Decrypt(data, tag []byte) error {
	c.recv.Decrypt(data)
	return nil
}
