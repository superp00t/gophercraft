package guid

import "github.com/superp00t/gophercraft/vsn"

//
type highTypeSupport map[HighType]uint64

var (
	htSupport = map[vsn.Build]highTypeSupport{
		OldFormat: { // Also works for 5875!
			Item:          0x4000, // blizz 4000
			Player:        0x0000, // blizz 0000
			GameObject:    0xF110, // blizz F110
			Transport:     0xF120, // blizz F120 (for GAMEOBJECT_TYPE_TRANSPORT)
			Creature:      0xF130, // blizz F130
			Pet:           0xF140, // blizz F140
			Vehicle:       0xF150, // blizz F550
			DynamicObject: 0xF100, // blizz F100
			Corpse:        0xF101, // blizz F100
			Mo_Transport:  0x1FC0, // blizz 1FC0 (for GAMEOBJECT_TYPE_MO_TRANSPORT)
			Instance:      0x1F40, // blizz 1F40
			Group:         0x1F50,
		},
	}
)
