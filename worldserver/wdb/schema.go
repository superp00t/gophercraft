package wdb

import "github.com/superp00t/gophercraft/econ"

type Character struct {
	ID          uint64     `json:"id" xorm:"'id' pk autoincr"`
	GameAccount uint64     `json:"gameAccount"`
	Name        string     `json:"name"`
	Faction     uint32     `json:"faction"`
	Level       uint8      `json:"level"`
	RealmID     uint64     `json:"realmID" xorm:"'realm_id'"`
	Race        uint8      `json:"race"`
	Class       uint8      `json:"class"`
	Gender      uint8      `json:"gender"`
	Skin        uint8      `json:"skin"`
	Face        uint8      `json:"face"`
	HairStyle   uint8      `json:"hairStyle"`
	HairColor   uint8      `json:"hairColor"`
	FacialHair  uint8      `json:"facialHair"`
	Coinage     econ.Money `json:"coinage"`
	Map         uint32     `json:"map"`
	X           float32    `json:"x"`
	Y           float32    `json:"y"`
	Z           float32    `json:"z"`
	O           float32    `json:"o"`
}

type Item struct {
	ID          uint64 `xorm:"'id' pk autoincr"`
	Owner       uint64 `xorm:"'owner'"`
	Equipped    bool   `xorm:"'equipped'"`
	ItemType    uint32
	DisplayID   uint32 `xorm:"'display_id'"`
	ItemID      uint32 `xorm:"'item_id'"`
	Enchantment uint32
}

type PortLocation struct {
	Name string  `xorm:"'port_id' pk" csv:"name"`
	X    float32 `xorm:"'x_pos'" csv:"x"`
	Y    float32 `xorm:"'y_pos'" csv:"y"`
	Z    float32 `xorm:"'z_pos'" csv:"z"`
	O    float32 `xorm:"'orientation'" csv:"orientation"`
	Map  uint32  `xorm:"'map'" csv:"mapID"`
}
