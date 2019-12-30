package worldserver

import (
	"strconv"
	"strings"

	"github.com/superp00t/etc/yo"

	"github.com/superp00t/gophercraft/guid"
	"github.com/superp00t/gophercraft/packet/update"
	"github.com/superp00t/gophercraft/worldserver/wdb"
)

// command invocation
type C struct {
	Session *Session
	Name    string
	Args    []string
}

func (c *C) String(index int) string {
	if len(c.Args) <= index {
		return ""
	}

	return c.Args[index]
}

func (c *C) Uint32(index int) uint32 {
	i, _ := strconv.ParseInt(c.String(index), 0, 64)
	return uint32(i)
}

func (c *C) Float32(index int) float32 {
	i, _ := strconv.ParseFloat(c.String(index), 32)
	return float32(i)
}

type Command struct {
	Signature   string
	ArgDemos    [][]string
	Required    Tier
	Description string
	Function    func(*C)
}

var (
	CmdHandlers = []Command{
		{
			"morph",
			[][]string{
				{"displayID"},
			},
			GameMaster,
			"changes your displayID",
			x_Morph,
		},

		{
			"scale",
			[][]string{
				{"scale"},
			},
			GameMaster,
			"changes your scale",
			x_Scale,
		},

		{
			"go",
			[][]string{
				{"portID"},
				{"mapID", "x", "y", "z", "o"},
			},
			GameMaster,
			"teleport to a location",
			x_Tele,
		},

		{
			"appear",
			[][]string{
				{"playerName"},
			},
			GameMaster,
			"appear to a player",
			x_Appear,
		},

		{
			"sound",
			[][]string{
				{"soundID"},
			},
			GameMaster,
			"plays a sound throughout the current Map",
			x_Sound,
		},

		{
			"vmod",
			[][]string{{"Global=valueType:Value"}},
			GameMaster,
			"modifies a set of values",
			x_VMod,
		},
	}
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

func x_Tele(c *C) {
	// port string
	if len(c.Args) < 5 && len(c.Args) != 1 {
		c.Session.Sysf(".go <x> <y> <z> <o> <map>")
		return
	}

	pos := update.Quaternion{}

	var mapID uint32

	if len(c.Args) == 1 {
		portID := c.String(0)

		var port []wdb.PortLocation
		c.Session.WS.DB.Where("port_id = ?", portID).Find(&port)

		if len(port) == 0 {
			c.Session.Warnf("could not find port location: '%s'", portID)
			return
		}

		mapID = port[0].Map
		pos.X = port[0].X
		pos.Y = port[0].Y
		pos.Z = port[0].Z
		pos.O = port[0].O
	} else {
		pos.X = c.Float32(0)
		pos.Y = c.Float32(1)
		pos.Z = c.Float32(2)
		pos.O = c.Float32(3)
		mapID = c.Uint32(4)
	}

	c.Session.TeleportTo(mapID, pos)
}

func x_Appear(c *C) {
	name := c.String(0)

	c.Session.WS.PlayersL.Lock()
	player := c.Session.WS.PlayerList[name]
	c.Session.WS.PlayersL.Unlock()

	// todo: escape user input

	if player == nil {
		c.Session.Warnf("no such player as '%s' found.", name)
		return
	}

	if player.CurrentPhase != c.Session.CurrentPhase {
		c.Session.Warnf("'%s' is currently in phase %d. You must join this phase if you want to appear at this player's location.", name, player.CurrentPhase)
		return
	}

	targetMap := player.CurrentMap

	c.Session.TeleportTo(targetMap, player.Position())
}

func x_Sound(c *C) {
	// port string
	if len(c.Args) < 1 {
		c.Session.Annf(".sound <soundID>")
		return
	}

	c.Session.Map().PlaySound(c.Uint32(0))
}

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
		plyr, err := c.Session.WS.GetSessionByGUID(tgt)
		if err != nil {
			c.Session.Warnf("%s", err.Error())
			return
		}

		if plyr.Tier != Admin && plyr.Tier == GameMaster {
			c.Session.Warnf("you cannot modify other GMs with your current permissions.")
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
