package dbc

type Ent_ChrClasses struct {
	ID              uint32 `xorm:"'id' pk"`
	PlayerClass     uint32
	DamageBonusStat uint32
	DisplayPower    uint32
	PetNameToken    string
	Name            string `dbc:"(loc)"`
	Filename        string
	SpellClassSet   uint32
	Flags           uint32
}
