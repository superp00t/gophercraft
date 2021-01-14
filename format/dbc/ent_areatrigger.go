package dbc

type Ent_AreaTrigger struct {
	ID     uint32 `xorm:"'id' pk"`
	MapID  uint32
	X      float32
	Y      float32
	Z      float32
	Radius float32
	BoxX   float32 `dbc:"5875-(only)"`
	BoxY   float32 `dbc:"5875-(only)"`
	BoxZ   float32 `dbc:"5875-(only)"`
	BoxO   float32 `dbc:"5875-(only)"`
}
