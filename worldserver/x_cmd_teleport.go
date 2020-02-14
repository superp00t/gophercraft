package worldserver

import (
	"github.com/superp00t/etc/yo"
	"github.com/superp00t/gophercraft/packet/update"
	"github.com/superp00t/gophercraft/worldserver/wdb"
)

func x_Tele(c *C) {
	// port string
	if len(c.Args) < 5 && len(c.Args) != 1 {
		c.Session.Sysf(".go <x> <y> <z> <o> <map>")
		return
	}

	pos := update.Position{}

	var mapID uint32

	if len(c.Args) == 1 {
		portID := c.String(0)

		var port wdb.PortLocation
		found, _ := c.Session.WS.DB.Where("port_id = ?", portID).Get(&port)

		if !found {
			found, _ = c.Session.WS.DB.Where("port_id like ?", portID+"%").Get(&port)
			if !found {
				c.Session.Warnf("could not find port location: '%s'", portID)
				return
			}
			yo.Warn(port.Name)
			c.Session.Warnf("Could not find teleport location %s, sending you to %s.", portID, port.Name)
		}

		mapID = port.Map
		pos.X = port.X
		pos.Y = port.Y
		pos.Z = port.Z
		pos.O = port.O
	} else {
		pos.X = c.Float32(0)
		pos.Y = c.Float32(1)
		pos.Z = c.Float32(2)
		pos.O = c.Float32(3)
		mapID = c.Uint32(4)
	}

	c.Session.TeleportTo(mapID, pos)
}
