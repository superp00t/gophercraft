package mpq

import (
	"bytes"
	"compress/bzip2"
	"compress/zlib"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/superp00t/gophercraft/format/pkzip"
	"github.com/superp00t/gophercraft/format/sparse"

	"github.com/superp00t/etc"
	"github.com/ulikunitz/xz/lzma"
)

type CompressionType uint8

const (
	MPQ_COMPRESSION_HUFFMANN     CompressionType = 0x01 // Huffmann compression (used on WAVE files only)
	MPQ_COMPRESSION_ZLIB         CompressionType = 0x02 // ZLIB compression
	MPQ_COMPRESSION_PKWARE       CompressionType = 0x08 // PKWARE DCL compression
	MPQ_COMPRESSION_BZIP2        CompressionType = 0x10 // BZIP2 compression (added in Warcraft III)
	MPQ_COMPRESSION_SPARSE       CompressionType = 0x20 // Sparse compression (added in Starcraft 2)
	MPQ_COMPRESSION_ADPCM_MONO   CompressionType = 0x40 // IMA ADPCM compression (mono)
	MPQ_COMPRESSION_ADPCM_STEREO CompressionType = 0x80 // IMA ADPCM compression (stereo)
	// MPQ_COMPRESSION_LZMA         CompressionType = 0x12 // LZMA compression. Added in Starcraft 2. This value is NOT a combination of flags.
	// MPQ_COMPRESSION_NEXT_SAME    CompressionType = 0xFF // Same compression
)

func (c CompressionType) String() string {
	var flags = []string{}

	for x := 0; x <= 0xFF; x++ {
		k := CompressionType(x)
		if str := CompressionTable[k]; str != nil {
			if c&k != 0 {
				flags = append(flags, str.Name)
			}
		}
	}

	return strings.Join(flags, ",") + fmt.Sprintf(" (0x%08X)", uint32(c))
}

type dcEntry struct {
	Name string
	Func Decompressor
}

var CompressionOrder = []CompressionType{
	MPQ_COMPRESSION_HUFFMANN,
	MPQ_COMPRESSION_ZLIB,
	MPQ_COMPRESSION_PKWARE,
	MPQ_COMPRESSION_BZIP2,
	MPQ_COMPRESSION_ADPCM_MONO,
	MPQ_COMPRESSION_ADPCM_STEREO,
	MPQ_COMPRESSION_SPARSE,
}

var CompressionTable = map[CompressionType]*dcEntry{
	MPQ_COMPRESSION_HUFFMANN:     {"Huffman trees", dcHuff},
	MPQ_COMPRESSION_ZLIB:         {"zlib", dcZlib},
	MPQ_COMPRESSION_PKWARE:       {"PKWARE", dcl},
	MPQ_COMPRESSION_BZIP2:        {"bzip2", dcBzip2},
	MPQ_COMPRESSION_ADPCM_MONO:   {"wave (mono)", dcADPCMMono},
	MPQ_COMPRESSION_ADPCM_STEREO: {"wave (stereo)", dcADPCMStereo},
	MPQ_COMPRESSION_SPARSE:       {"Sparse", sparse.Decompress},
	// MPQ_COMPRESSION_LZMA:         {"LZMA", dcLZMA},
	// MPQ_COMPRESSION_NEXT_SAME:    {"Next same", nil},
}

type Decompressor func([]byte) ([]byte, error)

func DecompressBlock(version int, fl CompressionType, content []byte) ([]byte, error) {
	switch {
	case version <= 1:
		return sDecompress(version, fl, content)
	case version >= 2:
		return sDecompress2(version, fl, content)
	}

	panic("unknown version")
}

func sDecompress2(version int, fl CompressionType, input []byte) ([]byte, error) {
	var decompressors [2]Decompressor

	switch fl {
	case MPQ_COMPRESSION_ZLIB:
		decompressors[0] = dcZlib
	case MPQ_COMPRESSION_PKWARE:
		decompressors[0] = dcl
	case MPQ_COMPRESSION_BZIP2:
		decompressors[0] = dcBzip2
	// case MPQ_COMPRESSION_LZMA:
	// 	compressors[0] = dcLZMA
	case MPQ_COMPRESSION_SPARSE:
		decompressors[0] = sparse.Decompress
	case MPQ_COMPRESSION_SPARSE | MPQ_COMPRESSION_ZLIB:
		decompressors[0] = dcZlib
		decompressors[1] = sparse.Decompress
	case MPQ_COMPRESSION_SPARSE | MPQ_COMPRESSION_BZIP2:
		decompressors[0] = dcBzip2
		decompressors[1] = sparse.Decompress
	case MPQ_COMPRESSION_ADPCM_MONO | MPQ_COMPRESSION_HUFFMANN:
		decompressors[0] = dcHuff
		decompressors[1] = dcADPCMMono
	case MPQ_COMPRESSION_ADPCM_STEREO | MPQ_COMPRESSION_HUFFMANN:
		decompressors[0] = dcHuff
		decompressors[1] = dcADPCMStereo
	default:
		return nil, fmt.Errorf("mpq: MPQv%d file corrupted (invalid compression byte %s (%d, 0x%X))", version, fl, fl, uint8(fl))
	}

	content2, err := decompressors[0](input)
	if err != nil {
		return nil, fmt.Errorf("mpq: first decompressor, %s", err)
	}

	if decompressors[1] != nil {
		content2, err = decompressors[1](input)
		if err != nil {
			return nil, fmt.Errorf("mpq: second decompressor, %s", err)
		}
	}

	return content2, nil
}

func dcADPCMMono(input []byte) ([]byte, error) {
	return nil, fmt.Errorf("adpcm mono nyi")
}

func dcADPCMStereo(input []byte) ([]byte, error) {
	return nil, fmt.Errorf("adpcm stereo nyi")
}

func dcHuff(input []byte) ([]byte, error) {
	return nil, fmt.Errorf("huffman nyi")
}

func dcl(input []byte) ([]byte, error) {
	return pkzip.Decompress(input)
}

func dcBzip2(in []byte) ([]byte, error) {
	bf := etc.FromBytes(in)
	dr := bzip2.NewReader(bf)
	return ioutil.ReadAll(dr)
}

func dcZlib(in []byte) ([]byte, error) {
	dr, err := zlib.NewReader(bytes.NewReader(in))
	if err != nil {
		return nil, err
	}

	return ioutil.ReadAll(dr)
}

func dcLZMA(in []byte) ([]byte, error) {
	bf := etc.FromBytes(in)
	dr, err := lzma.NewReader(bf)
	if err != nil {
		return nil, err
	}
	return ioutil.ReadAll(dr)
}

// Version 1
func sDecompress(version int, flags CompressionType, in []byte) ([]byte, error) {
	successMask := flags

	for i, ctype := range CompressionOrder {
		if successMask&ctype != 0 {
			tbl := CompressionTable[ctype]
			if tbl == nil {
				return nil, fmt.Errorf("mpq: no known deccompressor func for %s", ctype)
			}

			dec, err := tbl.Func(in)
			if err != nil {
				return nil, fmt.Errorf("mpq: %d error decompressing %s: %s", i, ctype, err)
			}

			in = dec
			successMask &^= ctype
		}
	}

	// if successMask&MPQ_COMPRESSION_NEXT_SAME != 0 {
	// 	successMask &^= MPQ_COMPRESSION_NEXT_SAME
	// }

	if successMask > 0 {
		return nil, fmt.Errorf("mpq: unable to decompress %s", successMask)
	}

	return in, nil
}
