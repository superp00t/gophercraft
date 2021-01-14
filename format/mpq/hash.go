package mpq

import (
	"bytes"
	"fmt"
)

const (
	MPQ_HASH_TABLE_INDEX = 0x000
	MPQ_HASH_NAME_A      = 0x100
	MPQ_HASH_NAME_B      = 0x200
	MPQ_HASH_FILE_KEY    = 0x300
	MPQ_HASH_KEY2_MIX    = 0x400

	MPQNeutral    = 0
	MPQChinese    = 0x404
	MPQCzech      = 0x405
	MPQGerman     = 0x407
	MPQEnglish    = 0x409
	MPQSpanish    = 0x40a
	MPQFrench     = 0x40c
	MPQItalian    = 0x410
	MPQJapanese   = 0x411
	MPQKorean     = 0x412
	MPQDutch      = 0x413
	MPQPolish     = 0x415
	MPQPortuguese = 0x416
	MPQRusssian   = 0x419
	MPQEnglishUK  = 0x809
)

func (m *MPQ) GetHashTableOffset() int64 {
	return m.Header.ArchiveOffset + int64(m.Header.HashTableOffset)
}

func (m *MPQ) GetHashTableSize() int {
	return int(m.Header.HashTableSize)
}

type HashEntry struct {
	ID_A, ID_B uint32

	Locale   uint16
	Platform uint16

	BlockIndex uint32
}

func (m *MPQ) Query(n string) (*HashEntry, error) {
	// nI := hashString(n, MPQ_HASH_TABLE_INDEX)
	nA := hashString(n, MPQ_HASH_NAME_A)
	nB := hashString(n, MPQ_HASH_NAME_B)

	// Starting index
	// sI := int(nI & uint32(m.GetHashTableSize()))
	for i := 0; i < m.GetHashTableSize(); i++ {
		e := m.HashTable[i]
		if e.ID_A == nA {
			if e.ID_B == nB {
				return e, nil
			}
		}
	}

	return nil, fmt.Errorf("file not found: %s", n)
}

func (m *MPQ) ReadHashTable() {
	_, err := m.File.Seek(m.GetHashTableOffset(), 0)
	if err != nil {
		panic(err)
	}

	sz := m.GetHashTableSize()

	buf := make([]byte, sz*16)
	_, err = m.File.Read(buf)
	if err != nil {
		panic(err)
	}

	err = decrypt(hashString("(hash table)", MPQ_HASH_FILE_KEY), buf)
	if err != nil {
		panic(err)
	}
	m.HashTable = make([]*HashEntry, sz)
	mf := new(bytes.Buffer)
	mf.Write(buf)

	for i := 0; i < sz; i++ {
		h := &HashEntry{}
		h.ID_A = lu32(mf)
		h.ID_B = lu32(mf)
		h.Locale = lu16(mf)
		h.Platform = lu16(mf)
		h.BlockIndex = lu32(mf)
		m.HashTable[i] = h
	}
}
