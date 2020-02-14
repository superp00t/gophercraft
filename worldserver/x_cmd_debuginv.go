package worldserver

import (
	"github.com/superp00t/gophercraft/guid"
	"github.com/superp00t/gophercraft/packet/update"
)

func x_DebugInv(c *C) {
	c.Session.Warnf("Player:")

	for i := 0; i < 39; i++ {
		gid := c.Session.GetGUIDArrayValue(update.PlayerInventorySlots, i)
		if gid != guid.Nil {
			c.Session.Warnf(" %d: %s", i, gid)
		}
	}

	for i := 19; i < 23; i++ {
		g := c.Session.GetGUIDArrayValue(update.PlayerInventorySlots, i)
		if g != guid.Nil {
			c.Session.Warnf("Bag %d:", i)

			gArray, err := c.Session.GetBagItem(uint8(i)).Get(update.ContainerSlots)
			if err != nil {
				panic(err)
			}

			for idx, it := range gArray.([]*guid.GUID) {
				gd := guid.Nil
				if it != nil {
					gd = *it
					c.Session.Warnf(" %d: %s", idx, gd)
				}
			}
		}
	}
}

// func getObjectArgument(c *C) (guid.GUID, error) {

// }

// func x_ForceUpdate(c *C) {

// }
