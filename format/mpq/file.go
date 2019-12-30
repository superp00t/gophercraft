package mpq

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"strings"
)

type File struct {
	Name        string
	Hash        *HashEntry
	Block       *BlockEntry
	Reader      io.ReadSeeker
	Multipart   bool
	Compression uint8
	ReadSoFar   int64

	mPtr *MPQ
}

func (f *File) CompressedSize() int64 {
	d := int64(f.Block.CompressedSize)
	return d
}

func (f *File) GetBlockSize() int {
	if f.Multipart {
		return f.mPtr.BlockSize()
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

	buf := make([]byte, bs)
	_, err = f.Reader.Read(buf)
	if err != nil {
		panic(err)
	}

	key := hashString(f.Name, MPQ_HASH_FILE_KEY)

	if f.Block.Match(MPQ_FILE_ENCRYPTED) {
		dc := bytes.NewBuffer(buf)

		if f.Block.Match(MPQ_FILE_FIX_KEY) {
			key = (key + uint32(f.mPtr.Header.ArchiveOffset)) ^ uint32(f.Block.Size)
		}

		dcr := newDecryptReader(dc, key)
		nb := new(bytes.Buffer)
		io.Copy(nb, dcr)
		buf = nb.Bytes()
	}

	outBuf := bytes.NewBuffer(nil)

	crc := f.Block.Match(MPQ_FILE_SECTOR_CRC)

	if f.Block.Match(MPQ_FILE_PKZIP) {
		panic("pkzip unsupported")
	}

	sectorSize := uint32(512 << f.mPtr.Header.BlockSize)

	if f.Multipart && f.Block.Match(MPQ_FILE_COMPRESS) && f.IsCompressed() {
		sectors := (f.Block.Size / sectorSize) + 1
		if crc {
			sectors++
		}

		var p []uint32
		for i := uint32(0); i < (sectors + 1); i++ {
			o := i * 4
			p = append(p, binary.LittleEndian.Uint32(buf[o:o+4]))
		}
		bytesLeft := f.Block.Size
		readableSectors := len(p) - 1
		if crc {
			readableSectors--
			// crcData := binary.LittleEndian.Uint32(buf[p[readableSectors-1]:p[readableSectors]])
			// bytes := buf[p[0]:p[readableSectors-1]]
			// actualCrc := adler32(0, bytes)
			// yo.Fatalf("0x%x, 0x%x", p[readableSectors], actualCrc)
		}

		dwBytesInThisSector := f.Block.Size

		for i := 0; i < readableSectors; i++ {
			sector := buf[p[i]:p[i+1]]

			if dwBytesInThisSector > bytesLeft {
				dwBytesInThisSector = bytesLeft
			}

			if f.Block.Match(MPQ_FILE_PKZIP) {
				panic("mpq: pkzip unsupported")
			}

			if f.Block.Match(MPQ_FILE_COMPRESS) {
				// yo.Println(i, p[i], "->", p[i+1], p[i+1]-p[i], bytesLeft, len(sector))
				sector, err = DecompressBlock(f.mPtr.Version(), sectorSize, sector)
				if err != nil {
					panic(err)
				}
			}

			bytesLeft -= uint32(len(sector))
			outBuf.Write(sector)
		}
	} else {
		start := binary.LittleEndian.Uint32(buf[:4])
		end := binary.LittleEndian.Uint32(buf[4:8])

		offset := 8
		if crc {
			offset += 4
		}

		data := buf[start:end]

		if f.Block.Match(MPQ_FILE_PATCH_FILE) {
			panic("cannot handle patch files")
		}

		if f.Block.Match(MPQ_FILE_COMPRESS) {
			buf, err = DecompressBlock(f.mPtr.Version(), f.Block.Size, data)
			if err != nil {
				panic(err)
			}
		}
		outBuf.Write(buf)
	}

	f.Close()
	return outBuf.Bytes(), nil
}

type position struct {
	pos uint32
}

func (f *File) Close() error {
	f.mPtr.L.Unlock()
	return nil
}

func (f *File) GetFileOffset() int64 {
	i := int64(f.mPtr.Header.ArchiveOffset) + int64(f.Block.FileOffset)
	return i
}

func (m *MPQ) OpenFile(name string) (*File, error) {
	m.L.Lock()
	e, err := m.Query(name)
	if err != nil {
		return nil, err
	}

	// Instantiate File object
	f := new(File)
	f.Name = name
	f.mPtr = m
	f.Hash = e
	f.Block = m.BlockTable[int(f.Hash.BlockIndex)]
	f.Reader = m.File

	if !f.Block.Match(MPQ_FILE_EXISTS) {
		return nil, fmt.Errorf("File doesn't even exist, apparently")
	}

	f.Multipart = !f.Block.Match(MPQ_FILE_SINGLE_UNIT)

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
