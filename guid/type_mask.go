package guid

import "fmt"

type TypeMask uint32

const (
	TypeMaskObject               TypeMask = 0x0001
	TypeMaskItem                 TypeMask = 0x0002
	TypeMaskContainer            TypeMask = 0x0004
	TypeMaskAzeriteEmpoweredItem TypeMask = 0x0008
	TypeMaskAzeriteItem          TypeMask = 0x0010
	TypeMaskUnit                 TypeMask = 0x0020
	TypeMaskPlayer               TypeMask = 0x0040
	TypeMaskActivePlayer         TypeMask = 0x0080
	TypeMaskGameObject           TypeMask = 0x0100
	TypeMaskDynamicObject        TypeMask = 0x0200
	TypeMaskCorpse               TypeMask = 0x0400
	TypeMaskAreaTrigger          TypeMask = 0x0800
	TypeMaskSceneObject          TypeMask = 0x1000
	TypeMaskConversation         TypeMask = 0x2000
	TypeMaskSeer                 TypeMask = TypeMaskPlayer | TypeMaskUnit | TypeMaskDynamicObject
)

type TypeMaskDescriptor map[TypeMask]uint32

var (
	TypeMaskDescriptors = map[uint32]TypeMaskDescriptor{
		5875: {
			TypeMaskObject:        0x0001,
			TypeMaskItem:          0x0002,
			TypeMaskContainer:     0x0004,
			TypeMaskUnit:          0x0008,
			TypeMaskPlayer:        0x0010,
			TypeMaskGameObject:    0x0020,
			TypeMaskDynamicObject: 0x0040,
			TypeMaskCorpse:        0x0080,
		},
	}
)

func (t TypeMask) Resolve(version uint32) (uint32, error) {
	td, ok := TypeMaskDescriptors[version]
	if !ok {
		return 0, fmt.Errorf("guid: invalid version code %d", version)
	}

	out := uint32(0)

	for kMask, mask := range td {
		if t&kMask != 0 {
			out |= mask
		}
	}

	return out, nil
}
