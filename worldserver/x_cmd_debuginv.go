package worldserver

import (
	"github.com/superp00t/gophercraft/guid"
)

func x_DebugInv(c *C) {
	c.Session.Warnf("Player:")

	for i := 0; i < 39; i++ {
		gid := c.Session.GetGUID("InventorySlots", i)
		if gid != guid.Nil {
			c.Session.Warnf(" %d: %s", i, gid)
		}
	}

	for i := 19; i < 23; i++ {
		g := c.Session.GetGUID("InventorySlots", i)
		if g != guid.Nil {
			c.Session.Warnf("Bag %d:", i)

			gArray := c.Session.GetBagItem(uint8(i)).Get("Slots")

			for idx := 0; idx < gArray.Len(); idx++ {
				it := gArray.Index(idx).Interface().(guid.GUID)
				if it != guid.Nil {
					c.Session.Warnf(" %d: %s", idx, it)
				}
			}
		}
	}
}

// func getObjectArgument(c *C) (guid.GUID, error) {

// }

// func x_ForceUpdate(c *C) {

// }
