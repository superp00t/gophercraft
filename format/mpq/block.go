package mpq

import (
	"bytes"
	"strings"
)

type BlockFlags uint32

const (
	BlockPKZIP        BlockFlags = 0x00000100 // File is compressed using PKWARE Data compression library
	BlockCompress     BlockFlags = 0x00000200 // File is compressed using combination of compression methods
	BlockEncrypted    BlockFlags = 0x00010000 // The file is encrypted
	BlockFixKey       BlockFlags = 0x00020000 // The decryption key for the file is altered according to the position of the file in the archive
	BlockPatchFile    BlockFlags = 0x00100000 // The file contains incremental patch for an existing file in base MPQ
	BlockSingleUnit   BlockFlags = 0x01000000 // Instead of being divided to BlockFlags = 0x1000-bytes blocks, the file is stored as single unit
	BlockDeleteMarker BlockFlags = 0x02000000 // File is a deletion marker, indicating that the file no longer exists. This is used to allow patch archives to delete files present in lower-priority archives in the search chain. The file usually has length of 0 or 1 byte and its name is a hash
	BlockSectorCRC    BlockFlags = 0x04000000 // File has checksums for each sector (explained in the File Data section). Ignored if file is not compressed or imploded.
	BlockExists       BlockFlags = 0x80000000 // Set if file exists, reset when the file was deleted
)

func (bf BlockFlags) String() string {
	var s []string
	if bf&BlockPKZIP != 0 {
		s = append(s, "BlockPKZIP ")
	}
	if bf&BlockCompress != 0 {
		s = append(s, "BlockCompress")
	}
	if bf&BlockEncrypted != 0 {
		s = append(s, "BlockEncrypted")
	}
	if bf&BlockFixKey != 0 {
		s = append(s, "BlockFixKey")
	}
	if bf&BlockPatchFile != 0 {
		s = append(s, "BlockPatchFile")
	}
	if bf&BlockSingleUnit != 0 {
		s = append(s, "BlockSingleUnit")
	}
	if bf&BlockDeleteMarker != 0 {
		s = append(s, "BlockDeleteMarker")
	}
	if bf&BlockSectorCRC != 0 {
		s = append(s, "BlockSectorCRC")
	}
	if bf&BlockExists != 0 {
		s = append(s, "BlockExists")
	}
	return strings.Join(s, "|")
}

type BlockEntry struct {
	FileOffset     uint64
	CompressedSize uint32
	Size           uint32
	Flags          BlockFlags
}

func (m *BlockEntry) Match(flag BlockFlags) bool {
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

	decrypt(hashString("(block table)", MPQ_HASH_FILE_KEY), buf)

	d := bytes.NewBuffer(buf)

	for i := 0; i < m.GetBlockTableSize(); i++ {
		b := &BlockEntry{}
		b.FileOffset = uint64(lu32(d))
		b.CompressedSize = lu32(d)
		b.Size = lu32(d)
		b.Flags = BlockFlags(lu32(d))
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
