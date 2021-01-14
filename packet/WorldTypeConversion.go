package packet

import (
	"fmt"

	"github.com/superp00t/gophercraft/vsn"
)

var WorldTypeDescriptors = map[vsn.Build]map[WorldType]uint32{
	vsn.Alpha:   worldTypeDescriptor_Alpha,
	vsn.V1_12_1: worldTypeDescriptor_5875,
	vsn.V2_4_3:  worldTypeDescriptor_8606,
	vsn.V3_3_5a: worldTypeDescriptor_12340,
	vsn.V8_3_0:  worldTypeDescriptor_33369,
}

// the inverse of WorldTypeConversion
var worldTypeLookup = map[vsn.Build]map[uint32]WorldType{}

func init() {
	// Create fast lookup maps for converting a uint32 to a WorldType.
	for build, descriptor := range WorldTypeDescriptors {
		lookup := map[uint32]WorldType{}
		for wType, code := range descriptor {
			lookup[code] = wType
		}
		worldTypeLookup[build] = lookup
	}
}

func ConvertWorldTypeToUint(build vsn.Build, wt WorldType) (uint32, error) {
	desc, ok := WorldTypeDescriptors[build]
	if !ok {
		return 0, fmt.Errorf("packet: no WorldType descriptor found for %s", build)
	}

	uint32_, ok := desc[wt]
	if !ok {
		return 0, fmt.Errorf("packet: no uint found for %s in descriptor %s", wt, build)
	}

	return uint32_, nil
}

func LookupWorldType(build vsn.Build, u32 uint32) (WorldType, error) {
	desc, ok := worldTypeLookup[build]
	if !ok {
		return 0, fmt.Errorf("packet: no WorldType lookup table found for %s", build)
	}

	wt, ok := desc[u32]
	if !ok {
		return 0, fmt.Errorf("packet: no WorldType found for 0x%04X in lookup table %s", u32, build)
	}

	return wt, nil
}
