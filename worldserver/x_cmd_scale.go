package worldserver

import "github.com/superp00t/gophercraft/packet/update"

func x_Scale(c *C) {
	scale := c.Float32(0)
	if scale < .1 || scale > 1000 {
		c.Session.Warnf("scale must be [0.1 - 1000.0]")
		return
	}

	if scale == 0 {
		scale = 1
	}

	c.
		Session.
		Map().
		ModifyObject(c.Session.GUID(), map[update.Global]interface{}{
			update.ObjectScaleX: scale,
		})
}
