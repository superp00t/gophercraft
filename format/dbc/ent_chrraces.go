package dbc

type Ent_ChrRaces struct {
	ID                      uint32 `xorm:"'id' pk"`
	Flags                   uint32
	FactionID               uint32 `xorm:"'faction_id'"`
	ExplorationSound        uint32 `dbc:"5875-(only)"`
	MaleDisplayID           uint32 `xorm:"'male_display_id'"`
	FemaleDisplayID         uint32 `xorm:"'female_display_id'"`
	ClientPrefix            string
	Speed                   float32 `dbc:"12340-(disabled)"`
	BaseLanguage            uint32
	CreatureType            uint32
	LoginEffectSpellID      uint32 `dbc:"-5875(only)"`
	UnalteredVisualRaceID   uint32 `dbc:"-5875(only)"`
	ResSicknessSpellID      uint32
	SplashSoundID           uint32
	StartingTaxiNodes       uint32 `dbc:"-5875(only)"`
	ClientFileString        string
	CinematicSequenceID     uint32
	Alliance                uint32   `dbc:"12340-(only)"`
	Name                    string   `dbc:"(loc)"`
	NameFemale              string   `dbc:"8606-(only,loc)"`
	NameMale                string   `dbc:"8606-(only,loc)"`
	FacialHairCustomization []string `dbc:"5875-(only,len:2)"`
	HairCustomization       string   `dbc:"5875-(only)"`
	Expansion               uint32   `dbc:"8606-(only)"`
}
