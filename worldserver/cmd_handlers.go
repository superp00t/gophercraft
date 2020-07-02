package worldserver

import (
	"math"
	"strconv"
	"strings"

	"github.com/superp00t/gophercraft/gcore/sys"
)

// command invocation
type C struct {
	Session *Session
	Args    []string
}

func (c *C) String(index int) string {
	if len(c.Args) <= index {
		return ""
	}

	return c.Args[index]
}

func (c *C) ArgString() string {
	return strings.Join(c.Args, " ")
}

func (c *C) Uint32(index int) uint32 {
	i, _ := strconv.ParseUint(c.String(index), 0, 32)
	return uint32(i)
}

func (c *C) Int(index int) int {
	i, _ := strconv.ParseInt(c.String(index), 0, 64)
	return int(i)
}

func (c *C) Float32(index int) float32 {
	i, _ := strconv.ParseFloat(c.String(index), 32)
	if math.IsInf(float64(i), 0) {
		return 0
	}
	return float32(i)
}

type Command struct {
	Signature   string
	ArgDemos    [][]string
	Required    sys.Tier
	Description string
	Function    interface{}
}

var (
	CmdHandlers = []Command{
		{
			"morph",
			[][]string{
				{"displayID"},
			},
			sys.Tier_GameMaster,
			"changes your displayID",
			x_Morph,
		},

		{
			"modify",
			nil,
			sys.Tier_NormalPlayer,
			"modify an attribute",
			[]Command{
				{
					"scale",
					nil,
					sys.Tier_GameMaster,
					"changes your scale",
					x_Scale,
				},

				{
					"speed",
					nil,
					sys.Tier_GameMaster,
					"changes your speed",
					x_Speed,
				},
			},
		},

		{
			"go",
			[][]string{
				{"portID"},
				{"mapID", "x", "y", "z", "o"},
			},
			sys.Tier_GameMaster,
			"Teleports you to a location.",
			x_Tele,
		},

		{
			"gm",
			nil,
			sys.Tier_GameMaster,
			"Game Master commands",
			[]Command{
				{
					"fly",
					[][]string{{"on/off"}},
					sys.Tier_GameMaster,
					"Turns flying on or off.",
					x_Fly,
				},
			},
		},

		{
			"xgps",
			[][]string{
				{"yards", "direction"},
			},
			sys.Tier_GameMaster,
			"Teleports you in a certain direction.",
			x_XGPS,
		},

		{
			"gobject",
			[][]string{},
			sys.Tier_Admin,
			"",
			[]Command{
				{
					"spawn",
					[][]string{
						{"id"},
					},
					sys.Tier_Admin,
					"",
					x_Gobject_spawn,
				},
			},
		},

		{
			"appear",
			[][]string{
				{"playerName"},
			},
			sys.Tier_GameMaster,
			"Teleports you to a player",
			x_Appear,
		},

		{
			"summon",
			[][]string{
				{"playerName"},
			},
			sys.Tier_GameMaster,
			"Summons a player to your location",
			x_Summon,
		},

		{
			"sound",
			[][]string{
				{"soundID"},
			},
			sys.Tier_Admin,
			"plays a sound throughout the current Map",
			x_Sound,
		},

		{
			"vmod",
			[][]string{{"Global=valueType:Value"}},
			sys.Tier_Admin,
			"modifies a set of values",
			x_VMod,
		},

		{
			"sstats",
			nil,
			sys.Tier_Admin,
			"view server stats",
			x_Stats,
		},

		{
			"lookup",
			nil,
			0,
			"",
			[]Command{
				{
					"tele",
					[][]string{{"teleport location"}},
					sys.Tier_GameMaster,
					"find a teleport location",
					x_LookupTeleport,
				},

				{"item",
					[][]string{{"item"}},
					sys.Tier_GameMaster,
					"find an item",
					x_LookupItem,
				},

				{"object",
					[][]string{{"gobj"}},
					sys.Tier_GameMaster,
					"find a gameobject",
					x_LookupGameObject,
				},
			},
		},

		{
			"additem",
			nil,
			sys.Tier_GameMaster,
			"add an item",
			x_AddItem,
		},

		// debug functions

		{
			"_showsql",
			nil,
			sys.Tier_Admin,
			"",
			func(c *C) {
				c.Session.DB().ShowSQL(true)
			},
		},

		{
			"_debuginv",
			nil,
			sys.Tier_Admin,
			"",
			x_DebugInv,
		},
	}
)
