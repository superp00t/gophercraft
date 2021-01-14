package crypto

import (
	"crypto"
	"crypto/hmac"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"testing"
)

func TestSessionKeyGenerator(t *testing.T) {
	skg := NewSessionKeyGenerator(sha256.New, []byte{0x05, 0x09, 0x4d, 0x52, 0x2b, 0x74, 0xfa, 0xc5, 0x06, 0xb0, 0x9a, 0xd6, 0x87, 0xe5, 0xfb, 0x5b, 0x8b, 0xd1, 0x76, 0x46, 0xa2, 0x30, 0x67, 0xe7, 0xa8, 0x48, 0x2b, 0x25, 0x00, 0xf2, 0x13, 0x3e})
	var data [40]byte
	skg.Read(data[:])
	str := hex.EncodeToString(data[:])
	if str != "7964853f810dc158cf9f8e42cb12865746442c9e06047776a6258dae9d5f61f6549c61d3dd76399e" {
		t.Fatal(str)
	}

	localChallenge := []byte{0x78, 0xd9, 0xed, 0xd2, 0x7f, 0xc8, 0xf8, 0xcd, 0xd6, 0x21, 0xb2, 0x55, 0x39, 0x50, 0x2c, 0x7}
	serverChallenge := []byte{0x9b, 0xb, 0x26, 0xd9, 0x0, 0xeb, 0xd6, 0x89, 0x46, 0xa2, 0x72, 0x17, 0x12, 0xb4, 0x95, 0x8}
	EncryptionKeySeed := []byte{0xE9, 0x75, 0x3C, 0x50, 0x90, 0x93, 0x61, 0xDA, 0x3B, 0x07, 0xEE, 0xFA, 0xFF, 0x9D, 0x41, 0xB8}

	hmc := hmac.New(sha256.New, data[:])
	hmc.Write(localChallenge)
	hmc.Write(serverChallenge)
	hmc.Write(EncryptionKeySeed)

	encKey := make([]byte, 16)

	copy(encKey, hmc.Sum(nil)[:16])
	if hex.EncodeToString(encKey) != "b422561c5034ac681409cc3d284485a4" {
		t.Fatal(str)
	}

	sigData, _ := hex.DecodeString("d7d2775a138070c89cde38c12aca3c24bf8ad937e90be83d605c547cfa746a7645a83a8d06c383fae910dda382be9a66e0b2ae40e0e5f5b347a89d408163def23335b1aa134495396afd367d82ae93eca27ad5be91d92b4e91d9cfd4c18ca55bcb5d60aa327ccfb271816a15b16eb61609969a64d9876ce0605fb366c1e865a36378b066263c383b54abe0931873ee5c2bde08b9acff6341849cc86ed7926f1db4ca93762b173421906b828ab9d191fac99c714cc72d9aa7e92f7a5d25ed69a2d4f8a666fa862ed5a104cc7adcc6d7172e38e7e3c9825ca228c439c87dd0b2c8632770dc25233644781438fe46ce5a59dc44ec595700ddbd2de0adf6959ac25c")
	ReverseBytes(sigData)

	enableEncryptionSeed := []byte{0x90, 0x9C, 0xD0, 0x50, 0x5A, 0x2C, 0x14, 0xDD, 0x5C, 0x2C, 0xC0, 0x64, 0x14, 0xF3, 0xFE, 0xC9}

	hmc = hmac.New(sha256.New, encKey)
	hmc.Write([]byte{1})
	hmc.Write(enableEncryptionSeed)

	rsaKey := GetConnectionRSAKey().Public().(*rsa.PublicKey)

	err := rsa.VerifyPKCS1v15(rsaKey, crypto.SHA256, hmc.Sum(nil), sigData)
	if err != nil {
		t.Fatal(err)
	}

	sig, err := rsa.SignPKCS1v15(nil, GetConnectionRSAKey(), crypto.SHA256, hmc.Sum(nil))
	if err != nil {
		t.Fatal()
	}

	for _, v := range sig {
		fmt.Printf("0x%02X, ", v)
	}
	fmt.Println("")
}
