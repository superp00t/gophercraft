package guid

import (
	"fmt"

	"github.com/superp00t/etc"
)

//go:generate gcraft_stringer -type=TypeID
type TypeID uint8

const (
	TypeObject TypeID = iota
	TypeItem
	TypeContainer
	TypeAzeriteEmpoweredItem
	TypeAzeriteItem
	TypeUnit
	TypePlayer
	TypeActivePlayer
	TypeGameObject
	TypeDynamicObject
	TypeCorpse
	TypeAreaTrigger
	TypeSceneObject
	TypeConversation
)

type TypeIDDescriptor map[TypeID]uint8

var (
	TypeIDDescriptors = map[uint32]TypeIDDescriptor{
		5875: {
			TypeObject:    0,
			TypeItem:      1,
			TypeContainer: 2,
			TypeUnit:      3,
			TypePlayer:    4,

			TypeGameObject:    5,
			TypeDynamicObject: 6,
			TypeCorpse:        7,
		},
	}
)

func DecodeTypeID(version uint32, in *etc.Buffer) (TypeID, error) {
	desc, ok := TypeIDDescriptors[version]
	if !ok {
		return 0, fmt.Errorf("guid: cannot decode type ID for version %d", version)
	}

	code := in.ReadByte()
	resolved := TypeID(0)
	found := false

	for k, v := range desc {
		if v == code {
			found = true
			resolved = k
			break
		}
	}

	if !found {
		return 0, fmt.Errorf("guid: could not resolve type ID for %d", code)
	}

	return resolved, nil
}

func EncodeTypeID(version uint32, id TypeID, out *etc.Buffer) error {
	desc, ok := TypeIDDescriptors[version]
	if !ok {
		return fmt.Errorf("guid: cannot encode type ID for version %d", version)
	}

	code, ok := desc[id]
	if !ok {
		return fmt.Errorf("guid: cannot resolve code for typeID: %s", id)
	}

	out.WriteByte(code)
	return nil
}
