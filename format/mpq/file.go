package mpq

import (
	"encoding/binary"
	"fmt"
	"io"
	"strings"

	"github.com/superp00t/etc/yo"
)

type File struct {
	Name      string
	Hash      *HashEntry
	Block     *BlockEntry
	Reader    io.ReadSeeker
	Multipart bool
	Volume    *MPQ
}

func (f *File) CompressedSize() int64 {
	d := int64(f.Block.CompressedSize)
	return d
}

func (f *File) GetBlockSize() int {
	if f.Multipart {
		return f.Volume.BlockSize()
	}

	return int(f.CompressedSize())
}

func (f *File) IsCompressed() bool {
	return f.Block.CompressedSize < f.Block.Size
}

func (f *File) ReadBlock() ([]byte, error) {
	// Read compression type.
	_, err := f.Reader.Seek(f.GetFileOffset(), 0)
	if err != nil {
		return nil, err
	}

	bs := int(f.Block.CompressedSize)
	if bs == 0 {
		return nil, fmt.Errorf("mpq: file has a compressed size of 0 (size: %d)", f.Block.Size)
	}

	buf := make([]byte, bs)
	_, err = f.Reader.Read(buf)
	if err != nil {
		panic(err)
	}

	key := hashString(f.Name, MPQ_HASH_FILE_KEY)

	if f.Block.Match(BlockEncrypted) {
		if f.Block.Match(BlockFixKey) {
			key = (key + uint32(f.Block.FileOffset)) ^ uint32(f.Block.Size)
		}
	}

	// Allocate memory for decompressed file.
	decompressed := make([]byte, f.Block.Size)

	crc := f.Block.Match(BlockSectorCRC)

	if f.Block.Match(BlockPKZIP) {
		panic("pkzip unsupported")
	}

	sectorSize := uint32(512 << f.Volume.Header.BlockSize)

	if f.Multipart && f.Block.Match(BlockCompress) && f.IsCompressed() {
		// Actual block data is preceded by a uint32_t array whose length == sectors
		sectors := int((f.Block.Size+sectorSize-1)/sectorSize) + 1
		if crc {
			sectors++
		}

		sectorBuffer := buf[:sectors*4]

		if f.Block.Match(BlockEncrypted) {
			// Awful.
			decrypt(key-1, sectorBuffer)
		}

		var packedOffsets []uint32

		for i := 0; i < sectors; i++ {
			offset := i * 4
			packedOffsets = append(packedOffsets, binary.LittleEndian.Uint32(sectorBuffer[offset:offset+4]))
		}

		bytesLeft := f.Block.Size
		bytesPos := 0
		readableSectors := sectors
		if crc {
			readableSectors--
		}

		dwBytesInThisSector := f.Block.Size

		for i := 0; i < readableSectors-1; i++ {
			sectorStart, sectorEnd := packedOffsets[i], packedOffsets[i+1]

			if int(sectorEnd) > len(buf) {
				yo.Ok(f.Name)
				panic(fmt.Errorf("sector pointed to data offset %d outside of the range (0-%d)", sectorEnd, len(buf)))
			}

			sector := buf[sectorStart:sectorEnd]

			if dwBytesInThisSector > bytesLeft {
				dwBytesInThisSector = bytesLeft
			}

			if f.Block.Match(BlockPKZIP) {
				panic("mpq: pkzip unsupported")
			}

			if f.Block.Match(BlockEncrypted) {
				// Grotesquely stupid scheme.
				decrypt(key+(uint32(i)), sector)
			}

			if f.Block.Match(BlockCompress) {
				sectorCompressed := true

				// Last sector is not compressed in certain cases. (What the fuck?)
				if bytesPos+len(sector) == int(f.Block.Size) {
					sectorCompressed = false
				}

				if sectorCompressed {
					compression := CompressionType(sector[0])

					rawSector, err := DecompressBlock(f.Volume.Version(), compression, sector[1:])
					if err != nil {
						// Should not occur.
						yo.Println(f.Volume.Path, f.Volume.Version(), compression, f.Block.Flags, i, "/", readableSectors-1)
						yo.Spew(sector)
						return nil, FileCorruptionError{f.Name, err}
					}
					sector = rawSector
				}
			}

			bytesLeft -= uint32(len(sector))

			// Copy decoded sector to pre-allocated buffer
			copy(decompressed[bytesPos:bytesPos+len(sector)], sector)
			bytesPos += len(sector)
		}

		if bytesPos != int(f.Block.Size) {
			return nil, fmt.Errorf("decoded data length %d is not the same size as specified in block entry %d", bytesPos, f.Block.Size)
		}
	} else {
		data := buf

		if f.Block.Match(BlockPatchFile) {
			panic("cannot handle patch files")
		}

		if f.Block.Match(BlockCompress) {
			if int(f.Block.Size) == len(data) {
				buf = data
			} else {
				buf, err = DecompressBlock(f.Volume.Version(), CompressionType(data[0]), data[1:])
				if err != nil {
					panic(err)
				}
			}
		}

		copy(decompressed[:], buf[:f.Block.Size])
	}

	return decompressed, nil
}

func (f *File) Close() error {
	f.Volume.GuardFile.Unlock()
	return nil
}

func (f *File) GetFileOffset() int64 {
	i := int64(f.Volume.Header.ArchiveOffset) + int64(f.Block.FileOffset)
	return i
}

func (m *MPQ) OpenFile(name string) (*File, error) {
	m.GuardFile.Lock()
	e, err := m.Query(name)
	if err != nil {
		return nil, err
	}

	// Instantiate File object
	f := new(File)
	f.Name = name
	f.Volume = m
	f.Hash = e
	f.Block = m.BlockTable[int(f.Hash.BlockIndex)]
	f.Reader = m.File

	if f.Block.Match(BlockDeleteMarker) {
		return nil, FileWasDeletedError(name + " in " + m.Path)
	}

	if !f.Block.Match(BlockExists) {
		return nil, fmt.Errorf("file %s doesn't even exist, apparently", name)
	}

	f.Multipart = !f.Block.Match(BlockSingleUnit)

	// Read raw data
	f.Reader = m.File

	return f, nil
}

func (m *MPQ) ListFiles() []string {
	f, err := m.OpenFile("(listfile)")
	if err != nil {
		fmt.Println("no listfile", err)
		return nil
	}
	defer f.Close()

	buf, err := f.ReadBlock()
	if err != nil {
		panic(err)
	}

	dat := string(buf)

	fin := []string{}
	fils := strings.Split(dat, "\r\n")
	for _, v := range fils {
		if v != "" {
			fin = append(fin, v)
		}
	}
	fils = nil
	return fin
}
