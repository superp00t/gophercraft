package update

import (
	"reflect"

	"github.com/superp00t/gophercraft/guid"
	"github.com/superp00t/gophercraft/vsn"
)

type DescriptorOptions uint32

const (
	DescriptorOptionClassicGUIDs = 1 << iota
	DescriptorOptionHasHasTransport
)

// Descriptor describes all the info about a particular version's SMSG_UPDATE_OBJECT strucutre.
type Descriptor struct {
	DescriptorOptions
	ObjectDescriptors map[guid.TypeMask]reflect.Type
}

var (
	Descriptors = map[vsn.Build]*Descriptor{}
)
