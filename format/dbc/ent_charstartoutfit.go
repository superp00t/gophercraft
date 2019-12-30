package dbc

import (
	"github.com/superp00t/gophercraft/packet"
)

const (
	Empty = 0xffffffff

	// InventoryTypes
	IT_Unequippable = 0
	IT_Head         = 1
	IT_Neck         = 2
	IT_Shoulder     = 3
	IT_Shirt        = 4
	IT_Chest        = 5
	IT_Waist        = 6
	IT_Legs         = 7
	IT_Feet         = 8
	IT_Wrists       = 9
	IT_Hands        = 10
	IT_Finger       = 11
	IT_Trinket      = 12
	IT_Weapon       = 13
	IT_Shield       = 14
	IT_Ranged       = 15
	IT_Back         = 16
	IT_TwoHand      = 17
	IT_Bag          = 18
	IT_Tabard       = 19
	IT_Robe         = 20
	IT_MainHand     = 21
	IT_OffHand      = 22
	IT_Holdable     = 23
	IT_Ammo         = 24
	IT_Thrown       = 25
	IT_Gun          = 26
	IT_Quiver       = 27
	IT_Relic        = 28
)

type Ent_CharStartOutfit struct {
	ID             uint32       `xorm:"'id' pk"`
	Race           packet.Race  `xorm:"'race'"`
	Class          packet.Class `xorm:"'class'"`
	Gender         uint8        `xorm:"'gender'"`
	OutfitID       uint8        `xorm:"'outfit_id'"`
	ItemIDs        []uint32     `xorm:"'item_ids'" dbc:"5875(len:12),12340(len:24)"`
	DisplayInfoIDs []uint32     `xorm:"'display_info_ids'" dbc:"5875(len:12),12340(len:24)"`
	InventoryTypes []uint32     `xorm:"'inventory_types'" dbc:"5875(len:12),12340(len:24)"`
}
