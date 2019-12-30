package packet

import (
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
