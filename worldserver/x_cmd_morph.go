package worldserver

import (
	"github.com/superp00t/etc/yo"
)

func x_Morph(c *C) {
	displayID := c.Uint32(0)

	yo.Ok("Morphing to ", displayID)

	c.Session.SetUint32("DisplayID", displayID)
	c.Session.Map().PropagateChanges(c.Session.GUID())
}
