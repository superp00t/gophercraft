package worldserver

import (
	"github.com/superp00t/gophercraft/vsn"
)

func x_Fly(c *C) {
	on := false
	if c.String(0) == "on" {
		on = true
	} else if c.String(0) == "off" {
		on = false
	} else {
		c.Session.Warnf("on/off")
		return
	}

	if c.Session.Build() < vsn.V2_4_3 {
		c.Session.Warnf("Flying is unstable in version %s.", c.Session.Build())
		c.Session.Warnf("Only lateral movement is allowed: You can use .xgps <distance> <up/down> to move vertically.")
	}

	c.Session.Warnf("Flight activated: %v. To turn off: .gm fly off", on)

	c.Session.SetFly(on)
}
