package commands

import (
	"reflect"

	"github.com/superp00t/gophercraft/realm"
)

type CommandProvider struct {
}

func (c *CommandProvider) Activated() (bool, error) {
	return true, nil
}

func (c *CommandProvider) Terminate() error {
	return nil
}

func (c *CommandProvider) Init(s *realm.Server, info *realm.PluginInfo) error {
	s.Cmd(realm.GameMaster, "help", "Show this menu", cmdHelp)

	s.Cmd(realm.GameMaster, "modify scale", "Changes the scale of your player model.\n  .modify scale 1.5", cmdScale)
	s.Cmd(realm.GameMaster, "modify speed", "Changes all movement speeds of your character. For 2X speed:\n  .modify speed 2.0", cmdSpeed)
	s.Cmd(realm.GameMaster, "modify coinage", "Add an amount of gold to your balance.\n  To remove 500 gold, 10 silver: .modify gold -500g10s", cmdMoney)
	s.Cmd(realm.GameMaster, "modify level", "Change your character's level", cmdModLevel)
	s.Cmd(realm.GameMaster, "morph", "Changes your displayID to the supplied integer.", cmdMorph)

	s.Cmd(realm.PhaseBuilder|realm.PhaseOwner, "npc add", "Spawn a Creature at your current position.", cmdAddNPC)

	s.Cmd(realm.PhaseBuilder|realm.PhaseOwner, "gobject add", "Spawn a GameObject at your current position.", cmdSpawnGameobject)

	s.Cmd(realm.PhaseOwner, "music", "Play music throughout the current map.", cmdMusic)
	s.Cmd(realm.PhaseOwner, "sound", "Play sound throughout the current map.", cmdSound)

	s.Cmd(realm.GameMaster, "tele", "Teleport to a Port ID", cmdTele)
	s.Cmd(realm.GameMaster, "gm fly", "Enable flight", cmdFly)

	s.Cmd(realm.GameMaster, "xgps", "To move 300 forward:\n  .xgps 300 f", cmdXGPS)
	s.Cmd(realm.GameMaster, "gps", "Query current position and area information.", cmdGPS)
	s.Cmd(realm.GameMaster, "summon", "Summons the named player to your location with their consent.", cmdSummon)
	s.Cmd(realm.GameMaster, "appear", "Teleport to the named player.", cmdAppear)

	s.Cmd(realm.GameMaster, "lookup teleport", "Find a teleport location.", cmdLookupTeleport)
	s.Cmd(realm.GameMaster, "lookup item", "Find a teleport location.", cmdLookupItem)
	s.Cmd(realm.GameMaster, "lookup gobject", "Find a gameobject.", cmdLookupGameObject)

	s.Cmd(realm.GameMaster, "additem", "Add an item.", cmdAddItem)

	// Admin utilities
	s.Cmd(realm.Admin, "vmod", "(Debug) Modify a value of your Player object directly", cmdVmod)
	s.Cmd(realm.Admin, "sstats", "(Debug) Query current server usage statistics", cmdStats)
	s.Cmd(realm.Admin, "funcdump", "(Debug) Dump currently running goroutine stack traces to stdout", cmdGoroutines)
	s.Cmd(realm.Admin, "showsql", "(Debug) Display XORM queries to stdout", cmdShowSQL)
	s.Cmd(realm.Admin, "debuginv", "(Debug) Show contents of inventory manager", cmdDebugInv)
	s.Cmd(realm.Admin, "trackedguids", "(Debug) List currently tracked objects", cmdTrackedGUIDs)
	s.Cmd(realm.Admin, "prop list", "(Debug)", cmdListProps)
	s.Cmd(realm.Admin, "prop add", "(Debug)", cmdAddProp)
	s.Cmd(realm.Admin, "prop list", "(Debug)", cmdRemoveProp)

	return nil
}

func init() {
	realm.RegisterPlugin("commands", &CommandProvider{})
}

func cmdHelp(c *realm.Session, prefix string) {
	for _, command := range c.WS.CommandHandlers {
		parameters := ""

		refl := reflect.ValueOf(command.Function)

		for x := 1; x < refl.Type().NumIn(); x++ {
			parameters += " <" + refl.Type().In(x).String() + ">"
		}

		c.ColorPrintf(realm.HelpColor, ".%s%s", command.Signature, parameters)
		c.ColorPrintf(realm.DemoColor, "  %s", command.Description)
	}
}
