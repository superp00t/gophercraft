package update

import (
	"github.com/superp00t/etc/yo"
	"github.com/superp00t/gophercraft/guid"
)

// CreateBlock is sent when the server notifies the client of a new game object. (Player, Unit, Item, etc.)
// It contains with it a MovementBlock and a ValuesBlock, which are later updated as individual blocks.
type CreateBlock struct {
	BlockType     BlockType
	ObjectType    guid.TypeID
	MovementBlock *MovementBlock
	ValuesBlock   *ValuesBlock
}

func (cb *CreateBlock) WriteData(e *Encoder, mask VisibilityFlags, create bool) error {
	err := guid.EncodeTypeID(e.Build, cb.ObjectType, e)
	if err != nil {
		return err
	}
	if err = cb.MovementBlock.WriteData(e, mask, true); err != nil {
		return err
	}

	if err = cb.ValuesBlock.WriteData(e, mask, true); err != nil {
		return err
	}
	return nil
}

func (cb *CreateBlock) Type() BlockType {
	return cb.BlockType
}

func (decoder *Decoder) DecodeGUID() (guid.GUID, error) {
	return guid.DecodePacked(decoder.Build, decoder.Reader)
}

func (decoder *Decoder) DecodeCreateBlock() (*CreateBlock, error) {
	var err error

	cBlock := &CreateBlock{
		BlockType:   decoder.CurrentBlockType,
		ValuesBlock: &ValuesBlock{},
	}

	cBlock.ObjectType, err = guid.DecodeTypeID(decoder.Build, decoder.Reader)
	if err != nil {
		return nil, err
	}

	cBlock.MovementBlock, err = decoder.DecodeMovementBlock()
	if err != nil {
		return nil, err
	}

	err = decoder.DecodeValuesBlockData(cBlock.ValuesBlock)
	if err != nil {
		yo.Spew(cBlock.MovementBlock)
		return nil, err
	}

	return cBlock, nil
}
