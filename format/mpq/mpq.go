//Package mpq allows the reading of compressed data from MPQ archives.
/*

   Based on http://www.zezula.net/en/mpq/mpqformat.html
   Cryptographic functions taken from https://github.com/aphistic/go.Zamara

   TODO: - support table encryption
		 - implement test files for MPQ versions 1-4
		 - be memory efficient
*/
package mpq

import (
	"fmt"
	"io"
	"math"
	"os"
	"sync"
)

const (
	MPQ_HEADER_DATA = "MPQ\x1A"
	MPQ_USER_DATA   = "MPQ\x1B"

	SectorSize = 512

	MD5_ListSize = 6

	MD5_BlockTable int = iota
	MD5_HashTable
	MD5_HiBlockTable
	MD5_BETTable
	MD5_HETTable
	MD5_MPQHeader
)

type MPQ struct {
	Path       string
	Header     *Header
	UserData   *UserData
	File       io.ReadSeeker
	GuardFile  sync.Mutex
	HashTable  []*HashEntry
	BlockTable []*BlockEntry
	// Prevent access of multiple files at the same time
}

func pow(x, y int) int {
	return int(math.Pow(float64(x), float64(y)))
}

func (m *MPQ) BlockSize() int {
	return SectorSize * pow(2, int(m.Header.BlockSize))
}

func (m *MPQ) Version() int {
	return int(m.Header.FormatVersion) + 1
}

func Open(filename string) (*MPQ, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	m, err := Decode(f)
	if err != nil {
		return nil, err
	}

	m.Path = filename
	return m, nil
}

func (m *MPQ) Close() error {
	if file, ok := m.File.(*os.File); ok {
		file.Close()
		return nil
	}
	return fmt.Errorf("mpq: cannot close a non-file input")
}

func Decode(i io.ReadSeeker) (*MPQ, error) {
	i.Seek(0, 0)

	h := new(Header)
	var s [4]byte
	i.Read(s[:])

	m := new(MPQ)
	m.Header = h
	m.File = i
	switch string(s[:]) {
	case MPQ_HEADER_DATA:
		if e := m.ReadHeaderData(); e != nil {
			return nil, e
		}
	case MPQ_USER_DATA:
		m.ReadUserData()
		m.File.Seek(m.Header.ArchiveOffset, 0)
		m.File.Read(s[:])
		t := string(s[:])
		if t != MPQ_HEADER_DATA {
			return nil, fmt.Errorf("Could not find MPQ header")
		}
		if e := m.ReadHeaderData(); e != nil {
			return nil, e
		}
	default:
		return nil, fmt.Errorf("Invalid MPQ header")
	}

	m.ReadHashTable()
	m.ReadBlockTable()

	return m, nil
}
