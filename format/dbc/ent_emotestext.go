package dbc

type Ent_EmotesText struct {
	ID        uint32 `xorm:"'id' pk"`
	Name      string
	EmoteID   uint32   `xorm:"'emote_id'"`
	EmoteText []uint32 `dbc:"(len:16)"`
}
