package guid

import (
	"fmt"
	"strings"

	"github.com/superp00t/gophercraft/vsn"
)

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
	TypeMaskDescriptors = map[vsn.Build]TypeMaskDescriptor{
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

func ResolveTypeMask(version vsn.Build, unresolvedTypeMask uint32) (TypeMask, error) {
	td, ok := TypeMaskDescriptors[version]
	if !ok {
		return 0, fmt.Errorf("guid: invalid version code %d", version)
	}

	var out TypeMask

	for kMask, mask := range td {
		if unresolvedTypeMask&mask != 0 {
			out |= kMask
		}
	}

	return out, nil
}

func (t TypeMask) Resolve(version vsn.Build) (uint32, error) {
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

func (t TypeMask) String() string {
	var s []string
	if t&TypeMaskObject != 0 {
		s = append(s, "Object")
	}
	if t&TypeMaskItem != 0 {
		s = append(s, "Item")
	}
	if t&TypeMaskContainer != 0 {
		s = append(s, "Container")
	}
	if t&TypeMaskAzeriteEmpoweredItem != 0 {
		s = append(s, "AzeriteEmpoweredItem")
	}
	if t&TypeMaskAzeriteItem != 0 {
		s = append(s, "AzeriteItem")
	}
	if t&TypeMaskUnit != 0 {
		s = append(s, "Unit")
	}
	if t&TypeMaskPlayer != 0 {
		s = append(s, "Player")
	}
	if t&TypeMaskActivePlayer != 0 {
		s = append(s, "ActivePlayer")
	}
	if t&TypeMaskGameObject != 0 {
		s = append(s, "GameObject")
	}
	if t&TypeMaskDynamicObject != 0 {
		s = append(s, "DynamicObject")
	}
	if t&TypeMaskCorpse != 0 {
		s = append(s, "Corpse")
	}
	if t&TypeMaskAreaTrigger != 0 {
		s = append(s, "AreaTrigger")
	}
	if t&TypeMaskSceneObject != 0 {
		s = append(s, "SceneObject")
	}
	if t&TypeMaskConversation != 0 {
		s = append(s, "Conversation")
	}
	if t&TypeMaskSeer != 0 {
		s = append(s, "Seer")
	}
	return strings.Join(s, "|")
}
