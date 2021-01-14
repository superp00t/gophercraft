package dbc

type Ent_AreaTable struct {
	ID                          uint32
	ContinentID                 uint32
	ParentArea                  uint32
	ExploreFlag                 uint32
	Flags                       uint32
	SoundProviderPref           uint32
	SoundProviderPrefUnderwater uint32
	AmbienceID                  uint32
	ZoneMusic                   uint32
	IntroMusic                  uint32
	ExplorationLevel            uint32
	Name                        string `dbc:"(loc)"`
	FactionGroup                uint32
	LiquidTypes                 []uint32 `dbc:"-5875(len:1),8606-(len:4)"`
	MinElevation                float32
	AmbientMultiplier           float32
	LightID                     uint32 `dbc:"8606(disabled)"`
}
