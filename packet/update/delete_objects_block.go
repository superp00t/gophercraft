package update

import (
	"fmt"

	"github.com/superp00t/gophercraft/guid"
)

type DeleteObjectsBlock struct {
	BlockType BlockType
	GUIDs     []guid.GUID
}

func (dec *Decoder) DecodeDeleteObjectsBlock() (*DeleteObjectsBlock, error) {
	if dec.CurrentBlockType != DeleteFarObjects && dec.CurrentBlockType != DeleteNearObjects {
		return nil, fmt.Errorf("update: mismatched call to DecodeDeleteObjectBlock (current block type is %s)", dec.CurrentBlockType)
	}

	do := &DeleteObjectsBlock{}
	do.BlockType = dec.CurrentBlockType

	guidCount, err := readUint32(dec.Reader)
	if err != nil {
		return nil, err
	}

	if guidCount > 4096 {
		return nil, fmt.Errorf("update: too many objects to delete. Probably a stream parsing error.")
	}

	for x := 0; x < int(guidCount); x++ {
		g, err := dec.DecodeGUID()
		if err != nil {
			return nil, err
		}

		do.GUIDs = append(do.GUIDs, g)
	}

	return do, nil
}

func (f *DeleteObjectsBlock) Type() BlockType {
	return f.BlockType
}

func (f *DeleteObjectsBlock) WriteAll(e *Encoder) error {
	writeUint32(e, uint32(len(f.GUIDs)))
	for _, g := range f.GUIDs {
		g.EncodePacked(e.Build, e)
	}
	return nil
}
