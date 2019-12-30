package dbc

type Ent_ChrRaces struct {
	ID                      uint32 `xorm:"'id' pk"`
	Flags                   uint32
	FactionID               uint32 `xorm:"'faction_id'"`
	ExplorationSound        uint32
	MaleDisplayID           uint32 `xorm:"'male_display_id'"`
	FemaleDisplayID         uint32 `xorm:"'female_display_id'"`
	ClientPrefix            string
	MountScale              float32
	BaseLanguage            uint32
	CreatureType            uint32
	LoginEffectSpellID      uint32
	CombatStunSpellID       uint32
	ResSicknessSpellID      uint32
	SplashSoundID           uint32
	StartingTaxiNodes       uint32
	ClientFileString        string
	CinematicSequenceID     uint32
	NameLang                string   `dbc:"(loc)"`
	FacialHairCustomization []string `dbc:"5875(len:2)"`
	HairCustomization       string
}
