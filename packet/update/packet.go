package update

import (
	"github.com/superp00t/etc"
	"github.com/superp00t/gophercraft/guid"
)

//go:generate gcraft_stringer -type=BlockType
type BlockType int

const (
	Values BlockType = iota
	Movement
	CreateObject
	SpawnObject
	DeleteFarObjects
	DeleteNearObjects
)

type Encoder struct {
	*etc.Buffer
	Mask    ValueMask
	Version uint32
}

type ValuesEncoder struct {
	enc *Encoder

	// key is an enabled offset, value is a 4-byte attribute chunk.
	Blocks map[int][]byte
}

type ValuesDecoder struct {
	*etc.Buffer
	ValuesBlock *ValuesBlock
	Version     uint32

	EnabledOffsets []uint32
	OffsetIndex    uint32
}

type BlockData interface {
	Type() BlockType
	WriteTo(guid.GUID, *Encoder) error
}

type Block struct {
	GUID guid.GUID
	Data BlockData
}
