package update

import (
	"fmt"
	"io"

	"github.com/superp00t/gophercraft/guid"
	"github.com/superp00t/gophercraft/vsn"
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

type BlockData interface {
	Type() BlockType
	WriteData(*Encoder, VisibilityFlags, bool) error
}

type Encoder struct {
	io.Writer
	Build      vsn.Build
	Descriptor *Descriptor
}

func NewEncoder(version vsn.Build, writer io.Writer, numBlocks int) (*Encoder, error) {
	e := &Encoder{
		Writer: writer,
		Build:  version,
	}

	desc, ok := Descriptors[version]
	if !ok {
		return nil, fmt.Errorf("update: no descriptor for encoder to use: %s", version)
	}

	writeUint32(writer, uint32(numBlocks))

	e.Descriptor = desc

	if e.Descriptor.DescriptorOptions&DescriptorOptionHasHasTransport != 0 {
		writeBool(writer, false)
	}

	return e, nil
}

func (enc *Encoder) EncodeGUID(id guid.GUID) error {
	if enc.Descriptor.DescriptorOptions&DescriptorOptionAlpha != 0 {
		id.EncodeUnpacked(enc.Build, enc)
		return nil
	}
	id.EncodePacked(enc.Build, enc)
	return nil
}

func (enc *Encoder) EncodeBlockType(bt BlockType) error {
	desc, ok := BlockTypeDescriptors[enc.Build]
	if !ok {
		return writeUint8(enc, uint8(bt))
	}

	value, ok := desc[bt]
	if !ok {
		panic("no type found for " + bt.String())
	}

	return writeUint8(enc, value)
}

func (enc *Encoder) AddBlock(id guid.GUID, data BlockData, viewMask VisibilityFlags) error {
	enc.EncodeBlockType(data.Type())

	if id != guid.Nil {
		enc.EncodeGUID(id)
	}

	return data.WriteData(enc, viewMask, false)
}
