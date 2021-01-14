package commands

import (
	"github.com/davecgh/go-spew/spew"
	"github.com/superp00t/gophercraft/packet/update"
	"github.com/superp00t/gophercraft/realm"
	"github.com/superp00t/gophercraft/realm/wdb"
)

func cmdTele(s *realm.Session, portID string) {
	// port string
	pos := update.Position{}

	var mapID uint32

	var port *wdb.PortLocation
	s.DB().GetData(portID, &port)

	if port == nil {
		var ports []*wdb.PortLocation
		if err := s.DB().SearchTemplates(portID, 1, &ports); err != nil {
			s.Warnf("%s", err)
			return
		}
		if len(ports) == 0 {
			s.Warnf("could not find port location: '%s'", portID)
			return
		}
		port = ports[0]
		s.Warnf("Could not find teleport location %s, sending you to %s.", portID, port.ID)
	}

	mapID = port.Map
	pos.X = port.X
	pos.Y = port.Y
	pos.Z = port.Z
	pos.O = port.O
	s.Warnf("%s", spew.Sdump(port))
	// }
	// } else {
	// 	pos.X = c.Float32(0)
	// 	pos.Y = c.Float32(1)
	// 	pos.Z = c.Float32(2)
	// 	pos.O = c.Float32(3)
	// 	mapID = c.Uint32(4)
	// }

	s.TeleportTo(mapID, pos)
}
