package mpq

import "bytes"

const (
	MPQ_FILE_PKZIP         uint32 = 0x00000100 // File is compressed using PKWARE Data compression library
	MPQ_FILE_COMPRESS      uint32 = 0x00000200 // File is compressed using combination of compression methods
	MPQ_FILE_ENCRYPTED     uint32 = 0x00010000 // The file is encrypted
	MPQ_FILE_FIX_KEY       uint32 = 0x00020000 // The decryption key for the file is altered according to the position of the file in the archive
	MPQ_FILE_PATCH_FILE    uint32 = 0x00100000 // The file contains incremental patch for an existing file in base MPQ
	MPQ_FILE_SINGLE_UNIT   uint32 = 0x01000000 // Instead of being divided to uint32 = 0x1000-bytes blocks, the file is stored as single unit
	MPQ_FILE_DELETE_MARKER uint32 = 0x02000000 // File is a deletion marker, indicating that the file no longer exists. This is used to allow patch archives to delete files present in lower-priority archives in the search chain. The file usually has length of 0 or 1 byte and its name is a hash
	MPQ_FILE_SECTOR_CRC    uint32 = 0x04000000 // File has checksums for each sector (explained in the File Data section). Ignored if file is not compressed or imploded.
	MPQ_FILE_EXISTS        uint32 = 0x80000000 // Set if file exists, reset when the file was deleted
)

type BlockEntry struct {
	FileOffset     uint64
	CompressedSize uint32
	Size           uint32
	Flags          uint32
}

func (m *BlockEntry) Match(flag uint32) bool {
	return (m.Flags & flag) != 0
}

func (m *MPQ) GetBlockTableSize() int {
	return int(m.Header.BlockTableSize)
}

// TODO: add high 16 bits for version 2
func (m *MPQ) GetBlockTableOffset() int64 {
	return m.Header.ArchiveOffset + int64(m.Header.BlockTableOffset)
}

func (m *MPQ) ReadBlockTable() {
	m.BlockTable = make([]*BlockEntry, m.GetBlockTableSize())

	buf := make([]byte, m.GetBlockTableSize()*16)
	m.File.Seek(m.GetBlockTableOffset(), 0)
	m.File.Read(buf)

	encryptor := newBlockEncryptor("(block table)", MPQ_HASH_FILE_KEY)
	encryptor.decrypt(&buf)

	d := bytes.NewBuffer(buf)

	for i := 0; i < m.GetBlockTableSize(); i++ {
		b := &BlockEntry{}
		b.FileOffset = uint64(lu32(d))
		b.CompressedSize = lu32(d)
		b.Size = lu32(d)
		b.Flags = lu32(d)
		m.BlockTable[i] = b
	}

	// If it exists, apply high bits array to offset table
	if m.Version() == 2 && m.Header.HiBlockTableOffset != 0 {
		m.File.Seek(int64(m.Header.ArchiveOffset)+int64(m.Header.HiBlockTableOffset), 0)
		for i := 0; i < m.GetBlockTableSize(); i++ {
			b := m.BlockTable[i]
			hiOffs := lu16(m.File)
			b.FileOffset |= (uint64(swapbytes16(hiOffs)) << 32)
		}
	}
}
