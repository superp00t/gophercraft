package mpq

import (
	"compress/bzip2"
	"compress/zlib"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/superp00t/etc/yo"
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
	MPQ_COMPRESSION_LZMA         CompressionType = 0x12 // LZMA compression. Added in Starcraft 2. This value is NOT a combination of flags.
	MPQ_COMPRESSION_NEXT_SAME    CompressionType = 0xFF // Same compression
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
	MPQ_COMPRESSION_LZMA:         {"LZMA", dcLZMA},
	MPQ_COMPRESSION_NEXT_SAME:    {"Next same", nil},
}

type Decompressor func([]byte) ([]byte, error)

func DecompressBlock(version int, compressionDisabledSize uint32, n []byte) ([]byte, error) {
	flags := n[0]
	fl := CompressionType(flags)
	content := n[1:]

	if compressionDisabledSize == uint32(len(n)) {
		return n, nil
	}

	compressors := []Decompressor{nil, nil}

	if version < 2 {
		switch fl {
		// multi compression
		case 2:
			return dcMulti(fl, content)
		default:
			return nil, fmt.Errorf("cannot decode multi-compression %s", fl)
		}
	} else {
		switch fl {
		case MPQ_COMPRESSION_ZLIB:
			compressors[0] = dcZlib
		case MPQ_COMPRESSION_PKWARE:
			compressors[0] = dcl
		case MPQ_COMPRESSION_BZIP2:
			compressors[0] = dcBzip2
		case MPQ_COMPRESSION_LZMA:
			compressors[0] = dcLZMA
		case MPQ_COMPRESSION_SPARSE:
			compressors[0] = sparse.Decompress
		case MPQ_COMPRESSION_SPARSE | MPQ_COMPRESSION_ZLIB:
			compressors[0] = dcZlib
			compressors[1] = sparse.Decompress
		case MPQ_COMPRESSION_SPARSE | MPQ_COMPRESSION_BZIP2:
			compressors[0] = dcBzip2
			compressors[1] = sparse.Decompress
		case MPQ_COMPRESSION_ADPCM_MONO | MPQ_COMPRESSION_HUFFMANN:
			compressors[0] = dcHuff
			compressors[1] = dcADPCMMono
		case MPQ_COMPRESSION_ADPCM_STEREO | MPQ_COMPRESSION_HUFFMANN:
			compressors[0] = dcHuff
			compressors[1] = dcADPCMStereo
		default:
			yo.Spew(n)
			return nil, fmt.Errorf("mpq: MPQv%d file corrupted (invalid compression byte %s (%d, 0x%X))", version, fl, fl, uint8(fl))
		}
	}

	content2, err := compressors[0](content)
	if err != nil {
		return nil, err
	}

	if compressors[1] != nil {
		content2, err = compressors[1](content2)
		if err != nil {
			return nil, err
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
	bf := etc.FromBytes(in)
	dr, err := zlib.NewReader(bf)
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
func dcMulti(flags CompressionType, in []byte) ([]byte, error) {
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

	if successMask&MPQ_COMPRESSION_NEXT_SAME != 0 {
		successMask &^= MPQ_COMPRESSION_NEXT_SAME
	}

	if successMask > 0 {
		return nil, fmt.Errorf("mpq: unable to decompress %s", successMask)
	}

	return in, nil
}
