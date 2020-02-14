package dbc

type Ent_AreaTrigger struct {
	ID     uint32 `xorm:"'id' pk"`
	MapID  uint32
	X      float32
	Y      float32
	Z      float32
	Radius float32
	BoxX   float32
	BoxY   float32
	BoxZ   float32
	BoxO   float32
}
