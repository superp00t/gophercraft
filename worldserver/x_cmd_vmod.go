package worldserver

import (
	"strconv"
	"strings"

	"github.com/superp00t/gophercraft/guid"
	"github.com/superp00t/gophercraft/packet/update"
)

func x_VMod(c *C) {
	if len(c.Args) < 1 {
		c.Session.Warnf(".vmod ValueOfPiGlobal=f3.14")
		return
	}

	tgt := c.Session.GUID()

	if t := c.Session.GetTarget(); t != guid.Nil {
		tgt = t
	}

	if tgt.HighType() == guid.Player {
		_, err := c.Session.WS.GetSessionByGUID(tgt)
		if err != nil {
			c.Session.Warnf("%s", err.Error())
			return
		}
	}

	changeMask := map[update.Global]interface{}{}

	for _, v := range c.Args {
		els := strings.Split(v, "=")
		if len(els) != 2 {
			c.Session.Warnf("failed to parse input")
			return
		}

		glob, err := update.GlobalFromString(els[0])
		if err != nil {
			c.Session.Warnf("invalid global key: %s", err.Error())
			return
		}

		value := els[1]
		if len(value) <= 1 {
			c.Session.Warnf("no value type specifier")
			return
		}

		switch value[0] {
		case 'f':
			f64, err := strconv.ParseFloat(value[1:], 32)
			if err != nil {
				c.Session.Warnf("%s", err.Error())
				return
			}

			changeMask[glob] = float32(f64)
		case 'i':
			i32, err := strconv.ParseInt(value[1:], 0, 32)
			if err != nil {
				c.Session.Warnf("%s", err.Error())
				return
			}

			changeMask[glob] = int32(i32)
		case 'u':
			u32, err := strconv.ParseUint(value[1:], 0, 32)
			if err != nil {
				c.Session.Warnf("%s", err.Error())
				return
			}

			changeMask[glob] = uint32(u32)
		case 'g':
			g, err := guid.FromString(value[1:])
			if err != nil {
				c.Session.Warnf("%s", err.Error())
				return
			}

			changeMask[glob] = g
		case 'b':
			b, err := strconv.ParseUint(value[1:], 0, 8)
			if err != nil {
				c.Session.Warnf("%s", err.Error())
				return
			}

			changeMask[glob] = uint8(b)
		}
	}

	c.Session.Map().ModifyObject(tgt, changeMask)
}
