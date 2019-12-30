package mpq

type Header struct {
	ArchiveOffset int64

	Type        uint32
	HeaderSize  uint32
	ArchiveSize uint32

	FormatVersion uint16
	BlockSize     uint16

	HashTableOffset  uint32
	BlockTableOffset uint32

	HashTableSize  uint32
	BlockTableSize uint32

	//-- MPQ HEADER v 2 -------------------------------------------
	HiBlockTableOffset uint64
	HashTableOffsetHi  uint16
	BlockTableOffsetHi uint16

	//-- MPQ HEADER v 3 -------------------------------------------
	ArchiveSize64    uint64
	BETTableOffset64 uint64
	HETTableOffset64 uint64

	//-- MPQ HEADER v 4 -------------------------------------------
	HashTableSize64    uint64
	BlockTableSize64   uint64
	HiBlockTableSize64 uint64
	HETTableSize64     uint64
	BETTableSize64     uint64
	RawChunkSize       uint32

	MD5 [][]byte
}

type UserData struct {
	Size         uint32
	HeaderOffset uint32
	HeaderSize   uint32
}

func (m *MPQ) ReadUserData() {
	m.UserData = new(UserData)
	m.UserData.Size = lu32(m.File)
	m.UserData.HeaderOffset = lu32(m.File)
	m.UserData.HeaderSize = lu32(m.File)

	m.Header.ArchiveOffset = int64(m.UserData.HeaderOffset)
}

func (m *MPQ) ReadHeaderData() error {
	m.Header.HeaderSize = lu32(m.File)

	// Deprecated in version 2
	m.Header.ArchiveSize = lu32(m.File)

	m.Header.FormatVersion = lu16(m.File)

	m.Header.BlockSize = lu16(m.File)

	m.Header.HashTableOffset = lu32(m.File)
	m.Header.BlockTableOffset = lu32(m.File)

	m.Header.HashTableSize = lu32(m.File)
	m.Header.BlockTableSize = lu32(m.File)

	v := m.Version()

	if v > 1 { // Version 2
		m.Header.HiBlockTableOffset = lu64(m.File)
		m.Header.HashTableOffsetHi = lu16(m.File)
		m.Header.BlockTableOffsetHi = lu16(m.File)
	}

	if v > 2 { // Version 3
		m.Header.ArchiveSize64 = lu64(m.File)
		m.Header.BETTableOffset64 = lu64(m.File)
		m.Header.HETTableOffset64 = lu64(m.File)
	}

	if v > 3 { // Version 4
		m.Header.HashTableSize64 = lu64(m.File)
		m.Header.BlockTableSize64 = lu64(m.File)
		m.Header.HiBlockTableSize64 = lu64(m.File)
		m.Header.HETTableSize64 = lu64(m.File)
		m.Header.BETTableSize64 = lu64(m.File)
		m.Header.RawChunkSize = lu32(m.File)

		m.Header.MD5 = make([][]byte, MD5_ListSize)
		for i := 0; i < MD5_ListSize; i++ {
			m.Header.MD5[i] = lb(m.File, 20)
		}
	}

	return nil
}
