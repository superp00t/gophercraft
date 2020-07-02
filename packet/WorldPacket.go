package packet

import (
	"encoding/binary"

	"github.com/superp00t/etc"
)

type WorldPacket struct {
	Type WorldType
	*etc.Buffer
}

func NewWorldPacket(t WorldType) *WorldPacket {
	return &WorldPacket{t, etc.NewBuffer()}
}

func (wp *WorldPacket) Finish() []byte {
	return wp.Buffer.Bytes()
}

func (wp *WorldPacket) ServerMessage() []byte {
	data := wp.Bytes()
	size := len(data) + 2
	var sizeBuffer [2]byte
	binary.BigEndian.PutUint16(sizeBuffer[:], uint16(size))
	var opcodeBuffer [2]byte
	binary.LittleEndian.PutUint16(opcodeBuffer[:], uint16(wp.Type))
	data = append(sizeBuffer[:], append(opcodeBuffer[:], data...)...)
	return data
}
