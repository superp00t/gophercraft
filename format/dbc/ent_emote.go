package dbc

type Ent_Emotes struct {
	ID                 uint32 `xorm:"'id' pk"`
	EmoteSlashCommand  string
	AnimID             uint32
	EmoteFlags         uint32
	EmoteSpecProc      uint32
	EmoteSpecProcParam uint32
	EventSoundID       uint32
}
