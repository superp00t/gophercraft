package dbc

type Ent_BarberShopStyle struct {
	ID           uint32
	Type         uint32
	Name         string `dbc:"(loc)"`
	Description  string `dbc:"(loc)"`
	CostModifier float32
	RaceID       uint32
	Gender       uint32
	HairID       uint32
}
