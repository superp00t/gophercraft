package packet

import (
	"encoding/binary"

	"github.com/superp00t/etc"
	"github.com/superp00t/gophercraft/vsn"
)

type WorldPacket struct {
	Type WorldType
	*etc.Buffer
	BitOffset uint8 // inverted
	Bits      uint8
}

func NewWorldPacket(t WorldType) *WorldPacket {
	return &WorldPacket{t, etc.NewBuffer(), 0, 0}
}

func (wp *WorldPacket) Finish() []byte {
	return wp.Buffer.Bytes()
}

func (wp *WorldPacket) ServerMessage(build vsn.Build) []byte {
	data := wp.Bytes()
	size := len(data) + 2
	var sizeBuffer [2]byte
	binary.BigEndian.PutUint16(sizeBuffer[:], uint16(size))
	var opcodeBuffer [2]byte
	op, err := ConvertWorldTypeToUint(build, wp.Type)
	if err != nil {
		panic(err)
	}
	binary.LittleEndian.PutUint16(opcodeBuffer[:], uint16(op))
	data = append(sizeBuffer[:], append(opcodeBuffer[:], data...)...)
	return data
}

func (wp *WorldPacket) WriteBit(t bool) {
	flag := uint8(1) << (7 - wp.BitOffset)
	if t {
		wp.Bits |= flag
	} else {
		wp.Bits &= ^flag
	}
	wp.BitOffset++
	if wp.BitOffset == 8 {
		wp.FlushBits()
		wp.BitOffset = 0
	}
}

func (wp *WorldPacket) FlushBits() {
	if wp.BitOffset == 0 {
		return
	}
	wp.WriteByte(wp.Bits)
	wp.BitOffset = 0
	wp.Bits = 0
}
