package update

import "github.com/superp00t/gophercraft/vsn"

var BlockTypeDescriptors = map[vsn.Build]map[BlockType]uint8{
	3368: {
		Values:            0,
		Movement:          1,
		CreateObject:      2,
		SpawnObject:       2,
		DeleteFarObjects:  3,
		DeleteNearObjects: 4,
	},

	5875: {
		Values:            0,
		Movement:          1,
		CreateObject:      2,
		SpawnObject:       3,
		DeleteFarObjects:  4,
		DeleteNearObjects: 5,
	},

	8606: {
		Values:            0,
		Movement:          1,
		CreateObject:      2,
		SpawnObject:       3,
		DeleteFarObjects:  4,
		DeleteNearObjects: 5,
	},
}
