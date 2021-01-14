package dbc

type Ent_ChrClasses struct {
	ID                  uint32 `xorm:"'id' pk"`
	Flag                uint32
	DamageBonusStat     uint32 `dbc:"-5875(only)"`
	PowerType           uint32
	PetNameToken        uint32 `dbc:"8606-(only)"`
	PetNameTokenString  string `dbc:"-5875(only)"`
	Name                string `dbc:"(loc)"`
	NameFemale          string `dbc:"8606-(only,loc)"`
	NameMale            string `dbc:"8606-(only,loc)"`
	Filename            string
	SpellClassSet       uint32
	Flags               uint32
	CinematicSequenceID uint32 `dbc:"12340-(only)"`
	Expansion           uint32 `dbc:"12340-(only)"`
}
