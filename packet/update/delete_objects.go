package update

import (
	"github.com/superp00t/etc"
	"github.com/superp00t/gophercraft/guid"
)

type DeleteObjectsBlock struct {
	BlockType BlockType
	GUIDs     []guid.GUID
}

func DecodeDeleteObjectsBlock(version uint32, t BlockType, in *etc.Buffer) (*DeleteObjectsBlock, error) {
	do := &DeleteObjectsBlock{}
	do.BlockType = t

	ln := in.ReadUint32()

	for x := 0; x < int(ln); x++ {
		if in.Available() < 1 {
			break
		}

		g, err := guid.DecodePacked(version, in)
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

func (f *DeleteObjectsBlock) WriteTo(blockGUID guid.GUID, e *Encoder) error {
	e.WriteUint32(uint32(len(f.GUIDs)))
	for _, g := range f.GUIDs {
		g.EncodePacked(e.Version, e)
	}
	return nil
}
