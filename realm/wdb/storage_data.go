package wdb

import (
	"github.com/superp00t/gophercraft/econ"
	"github.com/superp00t/gophercraft/guid"
)

// ObjectTemplateRegistry contains the resolved IDs of a custom object template (think custom items and creatures)
// It allows you to add and remove new content and (hopefully) avoid conflicts between the client's WDB cache and the server's in-memory database.
type ObjectTemplateRegistry struct {
	ID    string      `xorm:"'id' pk"`
	Type  guid.TypeID `xorm:"'type'"`
	Entry uint32
}

// Character describes a Player/Session's character attributes.
type Character struct {
	ID          uint64 `json:"id" xorm:"'id' pk autoincr"`
	GameAccount uint64 `json:"gameAccount"`
	Name        string `json:"name"`
	Faction     uint32 `json:"faction"`
	FirstLogin  bool
	Level       uint32 `json:"level"`
	XP          uint32
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
	Zone        uint32     `json:"zone"`
	Map         uint32     `json:"map"`
	X           float32    `json:"x"`
	Y           float32    `json:"y"`
	Z           float32    `json:"z"`
	O           float32    `json:"o"`
	Leader      uint64
	Guild       uint64
	HideHelm    bool
	HideCloak   bool
	Ghost       bool
}

// Item describes a *spawned* item. For the item's constant attributes, refer to ItemTemplate.
type Item struct {
	ID           uint64 `xorm:"'id' pk autoincr"`
	Creator      uint64 `xorm:"'creator'"` // player UID
	ItemType     uint32 `xorm:"'item_type'"`
	ItemID       string `xorm:"'item_id'"`
	DisplayID    uint32 `xorm:"'display_id'"`
	StackCount   uint32 `xorm:"'stack_count'"`
	Enchantments []uint32
	Charges      []int32 `xorm:"'charges'"`
}

// Inventory describes the positions of items/item stacks in a player's inventory.
type Inventory struct {
	ItemID uint64 `xorm:"'item_id' pk"`
	Player uint64 `xorm:"'player' index"`
	Bag    uint8  `xorm:"'bag'"`
	Slot   uint8  `xorm:"'slot'"`
}

// Contact describes the friend and ignore statuses of a Player in relation to another Player.
type Contact struct {
	Player   uint64 `xorm:"'player' index"`
	Friend   uint64 `xorm:"'friend'"`
	Friended bool   `xorm:"'friended'"`
	Ignored  bool   `xorm:"'ignored'"`
	Muted    bool   `xorm:"'muted'"`
	Note     string `xorm:"'note'"`
}

// LearnedAbility lists all the abilities/spells a player has learned.
type LearnedAbility struct {
	Player uint64 `xorm:"'player' index"`
	Spell  uint32 `xorm:"'spell'"`
	Active bool   `xorm:"'active'"`
}

// ActionButton stores all the buttons a player has in their action bars.
type ActionButton struct {
	Player uint64 `xorm:"'player' index"`
	Button uint8
	Action uint32
	Type   uint8
	Misc   uint8
}

// ExploredZone lists a player, and the zones which that player has explored in their map.
type ExploredZone struct {
	Player uint64 `xorm:"'player' index"`
	ZoneID uint32 `xorm:"'zone_id'"` // The actual zone ID, not the flag.
}

type PropID [8]byte

func (p PropID) String() string {
	return string(p[:])
}

func MakePropID(id string) PropID {
	var p PropID
	copy(p[:], []byte(id))
	return p
}

// An account having this property will
type AccountProp struct {
	ID   uint64 `xorm:"'id' pk"` // Account
	Prop PropID `xorm:"'prop' char(8)"`
}
