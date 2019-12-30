package packet

import (
	"bytes"
	"compress/zlib"
	"crypto/rand"
	"crypto/sha1"
	"encoding/binary"
	"io"
	"math"
	"time"

	"github.com/superp00t/etc"
	"github.com/superp00t/etc/yo"
)

func GetMSTime() uint32 {
	return uint32(time.Now().UnixNano() / int64(time.Millisecond))
}

type PackedXYZ struct {
	X, Y, Z, O float32
}

// func (e *EtcBuffer) ReadPackedXYZ() *PackedXYZ {
// 	packed := e.ReadUint32()
// 	x := float32(((packed & 0x7FF) << 21 >> 21)) * 0.25
// 	z := float32((((packed>>11)&0x7FF)<<21)>>21) * 0.25
// 	y := float32((packed>>22<<22)>>22) * 0.25

// 	return &PackedXYZ{
// 		x, y, z, 0,
// 	}
// }

func ReverseBuffer(input []byte) []byte {
	buf := make([]byte, len(input))
	inc := 0
	for x := len(input) - 1; x > -1; x-- {
		buf[inc] = input[x]
		inc++
	}
	return buf
}

func packetString(input string) []byte {
	data := []byte(input)
	data = bytes.Replace(data, []byte("."), []byte{0}, -1)
	return data
}

func randomBuffer(l int) []byte {
	buf := make([]byte, l)
	rand.Read(buf)
	return buf
}

func PutU32(u uint32) []byte {
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, u)
	return buf
}

func PutF32(u float32) []byte {
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, math.Float32bits(u))
	return buf
}

func F32(u float32) uint32 {
	return math.Float32bits(u)
}

func Hash(input ...[]byte) []byte {
	bt := sha1.Sum(bytes.Join(input, nil))
	return bt[:]
}

func BuildChecksum(data []byte) uint32 {
	h := Hash(data)
	var c uint32
	for i := 0; i < 5; i++ {
		o := i * 4
		nt := binary.LittleEndian.Uint32(h[o : o+4])
		c = c ^ nt
	}
	return c
}

func Uncompress(input []byte) []byte {
	in := etc.FromBytes(input)
	rdr, err := zlib.NewReader(in)
	if err != nil {
		panic(err)
		return nil
	}
	out := etc.NewBuffer()
	_, err = io.Copy(out, rdr)
	if err != nil {
		yo.Warn(err)
	}
	return out.Bytes()
}

func Compress(input []byte) []byte {
	b := etc.NewBuffer()
	z := zlib.NewWriter(b)
	z.Write(input)
	z.Flush()
	return b.Bytes()
}
