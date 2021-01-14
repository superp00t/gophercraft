package dbc

type Ent_Map_5875 struct {
	ID              uint32
	Directory       string
	InstanceType    uint32
	MapType         uint32
	Name            string `dbc:"(loc)"`
	MinLevel        uint32
	MaxLevel        uint32
	MaxPlayers      uint32
	EntranceMap     int32
	EntranceX       float32
	EntranceY       float32
	ParentMapID     uint32
	Description0    string `dbc:"(loc)"`
	Description1    string `dbc:"(loc)"`
	LoadingScreenID int32
	RaidOffset      uint32
	ContinentName   uint32
	Scale           float32
}

type Ent_Map_8606 struct {
	ID                   uint32
	Directory            string
	InstanceType         uint32
	MapType              uint32
	Name                 string   `dbc:"(loc)"`
	Field05              []uint32 `dbc:"(len:6)"`
	AreaTableID          uint32
	Description0         string `dbc:"(loc)"`
	Description1         string `dbc:"(loc)"`
	LoadingScreenID      uint32
	Field10              uint32
	Field11              uint32
	MinimapIconScale     float32
	RequirementText_Lang string `dbc:"(loc)"`
	HeroicText_Lang      string `dbc:"(loc)"`
	EmptyText2_Lang      string `dbc:"(loc)"`
	CorpseMapID          uint32
	CorpseX              float32
	CorpseY              float32
	ResetTimeRaid        uint32
	ResetTimeHeroic      uint32
	Field21              uint32
	TimeOfDayOverride    uint32
	ExpansionID          uint32
}
