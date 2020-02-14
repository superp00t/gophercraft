package update

import (
	"fmt"

	"github.com/superp00t/gophercraft/guid"

	"github.com/superp00t/etc"
	"github.com/superp00t/etc/yo"
)

// CreateBlock is sent when the server notifies the client of a new game object. (Player, Unit, Item, etc.)
type CreateBlock struct {
	BlockType     BlockType
	ObjectType    guid.TypeID
	MovementBlock *MovementBlock
	ValuesBlock   *ValuesBlock
}

// Data describes the entire content of an SMSG_UPDATE_OBJECT packet body.
type Data struct {
	HasTransport           bool
	Map                    uint16
	SmoothDeleteStartIndex uint16
	Blocks                 []Block
}

// Unmarshal selects a mode for decoding different versions of packets.
func Unmarshal(version uint32, data []byte) (*Data, error) {
	switch version {
	case 5875, 12340:
		return unmarshalClassic(version, data)
	// case version >= 31478:
	// 	return unmarshalModern(version, data)
	default:
		return nil, fmt.Errorf("update: unsupported version id %d", version)
	}
}

func unmarshalClassic(version uint32, input []byte) (*Data, error) {
	data := new(Data)

	i := etc.FromBytes(input)
	blockCount := i.ReadUint32()

	if version == 5875 {
		by := i.ReadByte()
		if by > 1 {
			return nil, fmt.Errorf("update: non-boolean byte %X (%d)encountered at HasTransport position", by, by)
		}

		data.HasTransport = by == 0
	}

	if blockCount > 2048 {
		return nil, fmt.Errorf("update: block count too large")
	}

	if blockCount == 0 {
		yo.Warn(i.Rpos())
		return nil, fmt.Errorf("update: no blocks")
	}

	for blockIndex := uint32(0); blockIndex < blockCount; blockIndex++ {
		if i.Available() < 1 {
			return nil, fmt.Errorf("update: unexpected end of stream")
		}

		updateType := BlockType(i.ReadByte())
		block := Block{}

		var err error

		switch updateType {
		case DeleteFarObjects, DeleteNearObjects:
			block.Data, err = DecodeDeleteObjectsBlock(version, updateType, i)
			if err != nil {
				return nil, err
			}
		case Values:
			block.GUID, err = guid.DecodePacked(version, i)
			if err != nil {
				return nil, err
			}
			block.Data, err = DecodeValuesClassic(block.GUID, version, i)
			if err != nil {
				return nil, err
			}
			yo.Spew(block.Data)
			data.Blocks = append(data.Blocks, block)
		// The SpawnObject code should be used for logging player into the world.
		case CreateObject, SpawnObject:
			var err error
			block.GUID, err = guid.DecodePacked(version, i)
			if err != nil {
				return nil, err
			}

			cBlock := &CreateBlock{
				BlockType: updateType,
			}

			cBlock.ObjectType, err = guid.DecodeTypeID(version, i)
			if err != nil {
				return nil, err
			}

			fmt.Println("Creating", block.GUID, "of type", cBlock.ObjectType)

			cBlock.MovementBlock, err = DecodeMovementBlock(version, i)
			if err != nil {
				return nil, err
			}
			cBlock.ValuesBlock, err = DecodeValuesClassic(block.GUID, version, i)
			if err != nil {
				return nil, err
			}

			block.Data = cBlock
			data.Blocks = append(data.Blocks, block)
		case Movement:
			var err error
			block.GUID, err = guid.DecodePacked(version, i)
			if err != nil {
				return nil, err
			}

			block.Data, err = DecodeMovementBlock(version, i)
			if err != nil {
				return nil, err
			}
		}
	}

	return data, nil
}

// const (
// 	modValues           = 0
// 	modCreate           = 1
// 	modSpawn            = 2
// 	modDeleteFarObjects = 3
// )

// func unmarshalModern(version uint32, data []byte) (*Data, error) {
// 	input := etc.FromBytes(data)

// 	d := new(Data)
// 	blockCount := input.ReadUint32()
// 	mapID := input.ReadUint16()

// 	farGUIDs := input.ReadBool()

// 	if farGUIDs {
// 		d.SmoothDeleteStartIndex = input.ReadUint16()
// 		farObjectsCount := input.ReadUint32()
// 		for x := 0; x < farObjectsCount; x++ {
// 			g, err := guid.DecodePacked(version, input)
// 			if err != nil {
// 				return nil, err
// 			}
// 			d.FarObjectsDelete = append(d.FarObjectsDelete, g)
// 		}
// 	}

// 	blockDataSize := input.ReadUint32()
// 	in := etc.FromBytes(input.ReadBytes(int(blockDataSize)))

// 	for x := 0; x < blockCount; x++ {
// 		ut := in.ReadByte()
// 		switch ut {
// 		case modValues:
// 			id, err := guid.DecodePacked(version, in)
// 			if err != nil {
// 				return nil, err
// 			}

// 			block, err := DecodeValues(id.HighType(), version, in)
// 			if err != nil {
// 				return nil, err
// 			}

// 			d.Blocks = append(d.Blocks, Block{
// 				id,
// 				block,
// 			})
// 		case modCreate, modSpawn:
// 			id, err := guid.DecodePacked(version, in)
// 			if err != nil {
// 				return nil, err
// 			}

// 		}
// 	}
// }
