package dbc

type Ent_CreatureFamily struct {
	ID             uint32
	MinScale       float32
	MinScaleLevel  uint32
	MaxScale       float32
	MaxScaleLevel  uint32
	SkillLine      []uint32 `dbc:"3368(only,len:2)"`
	PetFoodMask    uint32   `dbc:"5875-(only)"`
	PetTalentType  uint32   `dbc:"5875-(only)"`
	CategoryEnumID uint32   `dbc:"5875-(only)"`
	Name           string   `dbc:"5875-(only,loc)"`
	IconFile       string   `dbc:"5875-(only)"`
}
