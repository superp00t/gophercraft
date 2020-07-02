package update

import (
	"fmt"
	"io"
	"reflect"
	"sync"

	"github.com/superp00t/gophercraft/guid"
	"github.com/superp00t/gophercraft/vsn"
)

var (
	MaxBlockCount uint32 = 2048
)

type ValuesBlock struct {
	sync.Mutex

	TypeMask guid.TypeMask
	// Descriptor describes all the info about a particular version's SMSG_UPDATE_OBJECT.
	Descriptor *Descriptor
	ChangeMask *Bitmask
	// StorageDescriptor is the version and type-specific structure.
	StorageDescriptor reflect.Value
}

// Decoder decodes an SMSG_UPDATE_OBJECT input stream into various sub-structures.
type Decoder struct {
	*Descriptor
	Build                  vsn.Build
	HasTransport           bool
	Map                    uint16
	SmoothDeleteStartIndex uint16
	BlockCount             uint32
	CurrentBlockIndex      uint32
	CurrentBlockType       BlockType
	Reader                 io.Reader
}

// Decode
func NewDecoder(version vsn.Build, reader io.Reader) (*Decoder, error) {
	decoder := new(Decoder)
	decoder.Build = version
	decoder.Reader = reader
	decoder.Descriptor = Descriptors[version]
	if decoder.Descriptor == nil {
		return nil, fmt.Errorf("update: no descriptor for version %v", version)
	}

	var err error
	decoder.BlockCount, err = readUint32(reader)
	if err != nil {
		return nil, err
	}

	if decoder.DescriptorOptions&DescriptorOptionHasHasTransport != 0 {
		decoder.HasTransport, err = readBool(reader)
		if err != nil {
			return nil, err
		}
	}

	if decoder.BlockCount > MaxBlockCount {
		return nil, fmt.Errorf("update: block count (%d) > update.MaxBlockCount (%d)", decoder.BlockCount, MaxBlockCount)
	}

	return decoder, nil
}

func (decoder *Decoder) DecodeBlockType() (BlockType, error) {
	bt, err := readUint8(decoder.Reader)
	if err != nil {
		return 0, err
	}

	// todo: support new format

	return BlockType(bt), nil
}

func (decoder *Decoder) MoreBlocks() bool {
	if decoder.CurrentBlockIndex >= decoder.BlockCount {
		return false
	}

	return true
}

func (decoder *Decoder) NextBlock() (BlockType, error) {
	if !decoder.MoreBlocks() {
		return 0, io.EOF
	}

	var err error
	decoder.CurrentBlockType, err = decoder.DecodeBlockType()
	if err != nil {
		return 0, err
	}

	decoder.CurrentBlockIndex++

	return decoder.CurrentBlockType, nil
}
