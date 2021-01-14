package crypto

import "hash"

type SessionKeyGenerator struct {
	HashFunc func() hash.Hash
	Seeds    [3][]byte
}

func sHash(constructor func() hash.Hash, data ...[]byte) []byte {
	hash := constructor()
	for _, dat := range data {
		hash.Write(dat)
	}
	return hash.Sum(nil)
}

func NewSessionKeyGenerator(hashFunc func() hash.Hash, data []byte) *SessionKeyGenerator {
	size := len(data)
	halfSize := size / 2
	skg := &SessionKeyGenerator{
		HashFunc: hashFunc,
	}

	hash1 := sHash(skg.HashFunc, data[0:halfSize])
	hash0 := make([]byte, len(hash1))
	hash2 := sHash(skg.HashFunc, data[halfSize:])
	hash0 = sHash(skg.HashFunc, hash1, hash0, hash2)

	skg.Seeds[0] = hash0
	skg.Seeds[1] = hash1
	skg.Seeds[2] = hash2

	return skg
}

func (skg *SessionKeyGenerator) Read(data []byte) (int, error) {
	seedIndex := 0

	for i := range data {
		if seedIndex == len(skg.Seeds[0]) {
			skg.Seeds[0] = sHash(skg.HashFunc, skg.Seeds[1], skg.Seeds[0], skg.Seeds[2])
			seedIndex = 0
		}

		data[i] = skg.Seeds[0][seedIndex]

		seedIndex++
	}

	return len(data), nil
}
