package worldserver

import (
	"github.com/superp00t/etc/yo"
	"github.com/superp00t/gophercraft/packet/update"
)

func x_Morph(c *C) {
	displayID := c.Uint32(0)

	yo.Ok("Morphing to ", displayID)

	c.
		Session.
		Map().
		ModifyObject(c.Session.GUID(), map[update.Global]interface{}{
			update.UnitDisplayID: displayID,
		})
}
