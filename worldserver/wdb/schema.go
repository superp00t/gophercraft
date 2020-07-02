package wdb

import (
	"github.com/superp00t/gophercraft/econ"
	"github.com/superp00t/gophercraft/guid"
)

type ObjectTemplate interface {
	ID() string
	SetEntry(uint32)
}

type Character struct {
	ID          uint64     `json:"id" xorm:"'id' pk autoincr"`
	GameAccount uint64     `json:"gameAccount"`
	Name        string     `json:"name"`
	Faction     uint32     `json:"faction"`
	Level       uint32     `json:"level"`
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

type Inventory struct {
	ItemID uint64 `xorm:"'item_id' pk"`
	Player uint64 `xorm:"'player' index"`
	Bag    uint8  `xorm:"'bag'"`
	Slot   uint8  `xorm:"'slot'"`
}

type PortLocation struct {
	ID  string  `xorm:"'port_id' pk" csv:"name"`
	X   float32 `xorm:"'x_pos'" csv:"x"`
	Y   float32 `xorm:"'y_pos'" csv:"y"`
	Z   float32 `xorm:"'z_pos'" csv:"z"`
	O   float32 `xorm:"'orientation'" csv:"orientation	"`
	Map uint32  `xorm:"'map'" csv:"mapID"`
}

type ObjectTemplateRegistry struct {
	ID    string      `xorm:"'id' pk"`
	Type  guid.TypeID `xorm:"'type'"`
	Entry uint32
}

type ItemTemplate struct {
	Entry                     uint32       `json:",omitempty" xorm:"'entry' bigint pk" csv:"-"`
	ID                        string       `json:",omitempty" xorm:"'id' index"`
	Name                      string       `json:",omitempty" xorm:"'name' index"`
	Class                     uint32       `json:",omitempty" xorm:"'class'"`
	Subclass                  uint32       `json:",omitempty" xorm:"'subclass'"`
	DisplayID                 uint32       `json:",omitempty" xorm:"'display_id'"`
	Quality                   uint8        `json:",omitempty" xorm:"'quality'"`
	Flags                     string       `json:",omitempty" xorm:"'flags'"`
	BuyCount                  uint8        `json:",omitempty" xorm:"'buy_count'"`
	BuyPrice                  uint32       `json:",omitempty" xorm:"'buy_price'"`
	SellPrice                 uint32       `json:",omitempty" xorm:"'sell_price'"`
	InventoryType             uint8        `json:",omitempty" xorm:"'inv_type'"`
	AllowableClass            int32        `json:",omitempty" xorm:"'allowable_class'"`
	AllowableRace             int32        `json:",omitempty" xorm:"'allowable_race'"`
	ItemLevel                 uint32       `json:",omitempty" xorm:"'item_level'"`
	RequiredLevel             uint8        `json:",omitempty" xorm:"'required_level'"`
	RequiredSkill             uint32       `json:",omitempty" xorm:"'required_skill'"`
	RequiredSkillRank         uint32       `json:",omitempty" xorm:"'required_skill_rank'"`
	RequiredSpell             uint32       `json:",omitempty" xorm:"'required_spell'"`
	RequiredHonorRank         uint32       `json:",omitempty" xorm:"'required_honor_rank'"`
	RequiredCityRank          uint32       `json:",omitempty" xorm:"'required_city_rank'"`
	RequiredReputationFaction uint32       `json:",omitempty" xorm:"'required_reputation_faction'"`
	RequiredReputationRank    uint32       `json:",omitempty" xorm:"'required_reputation_rank'"`
	MaxCount                  uint32       `json:",omitempty" xorm:"'max_count'"`
	Stackable                 uint32       `json:",omitempty" xorm:"'stackable'"`
	ContainerSlots            uint8        `json:",omitempty" xorm:"'container_slots'"`
	Stats                     []ItemStat   `json:",omitempty" xorm:"'stats'"`
	Damage                    []ItemDamage `json:",omitempty" xorm:"'dmg'"`
	Armor                     uint32       `json:",omitempty" xorm:"'armor'"`
	HolyRes                   uint32       `json:",omitempty" xorm:"'holy_res'"`
	FireRes                   uint32       `json:",omitempty" xorm:"'fire_res'"`
	NatureRes                 uint32       `json:",omitempty" xorm:"'nature_res'"`
	FrostRes                  uint32       `json:",omitempty" xorm:"'frost_res'"`
	ShadowRes                 uint32       `json:",omitempty" xorm:"'shadow_res'"`
	ArcaneRes                 uint32       `json:",omitempty" xorm:"'arcane_res'"`
	Delay                     uint32       `json:",omitempty" xorm:"'delay'"`
	AmmoType                  uint32       `json:",omitempty" xorm:"'ammo_type'"`
	RangedModRange            float32      `json:",omitempty" xorm:"'ranged_mod_range'"`
	Spells                    []ItemSpell  `json:",omitempty" xorm:"'spells'"`
	Bonding                   uint8        `json:",omitempty" xorm:"'bonding'"`
	Description               string       `json:",omitempty" xorm:"'description' longtext"`
	PageText                  uint32       `json:",omitempty" xorm:"'page_text"`
	LanguageID                uint32       `json:",omitempty" xorm:"'language_id'"`
	PageMaterial              uint32       `json:",omitempty" xorm:"'page_material'"`
	StartQuest                uint32       `json:",omitempty" xorm:"'start_quest'"`
	LockID                    uint32       `json:",omitempty" xorm:"'lock_id'"`
	Material                  int32        `json:",omitempty" xorm:"'material'"`
	Sheath                    uint32       `json:",omitempty" xorm:"'sheath'"`
	RandomProperty            uint32       `json:",omitempty" xorm:"'random_property'"`
	RandomSuffix              uint32       `json:",omitempty" xorm:"'random_suffix'"`
	Block                     uint32       `json:",omitempty" xorm:"'block'"`
	Itemset                   uint32       `json:",omitempty" xorm:"'itemset'"`
	MaxDurability             uint32       `json:",omitempty" xorm:"'max_durability'"`
	Area                      uint32       `json:",omitempty" xorm:"'area'"`
	Map                       int32        `json:",omitempty" xorm:"'map'"`
	BagFamily                 int32        `json:",omitempty" xorm:"'bag_family'"`
	TotemCategory             int32        `json:",omitempty" xorm:"'totem_category'"`
	Socket                    []ItemSocket `json:",omitempty" xorm:"'sockets'"`
	GemProperties             int32        `json:",omitempty" xorm:"'gem_properties'"`
	RequiredDisenchantSkill   int32        `json:",omitempty" xorm:"'required_disenchant_skill'"`
	ArmorDamageModifier       float32      `json:",omitempty" xorm:"'armor_damage_modifier'"`
	ItemLimitCategory         int32        `json:",omitempty" xorm:"'item_limit_category'"`
	ScriptName                string       `json:",omitempty" xorm:"'script_name'"`
	DisenchantID              uint32       `json:",omitempty" xorm:"'disenchant_id'"`
	FoodType                  uint8        `json:",omitempty" xorm:"'food_type'"`
	MinMoneyLoot              uint32       `json:",omitempty" xorm:"'min_money_loot'"`
	MaxMoneyLoot              uint32       `json:",omitempty" xorm:"'max_money_loot'"`
	Duration                  int32        `json:",omitempty" xorm:"'duration'"`
	ExtraFlags                uint8        `json:",omitempty" xorm:"'extra_flags'"`
}

type ItemStat struct {
	Type  uint8
	Value int32
}

type ItemDamage struct {
	Type uint8
	Min  float32
	Max  float32
}

type ItemSpell struct {
	ID               uint32
	Trigger          uint32
	Charges          int32
	PPMRate          float32
	Cooldown         int64
	Category         uint32
	CategoryCooldown int64
}

type ItemSocket struct {
	Color   int32
	Content int32
}

type GameObjectTemplate struct {
	ID             string `xorm:"'id' index"`
	Entry          uint32 `csv:"-" xorm:"'entry' bigint pk"`
	Type           uint32
	DisplayID      uint32 `xorm:"'display_id'"`
	Name           string
	IconName       string
	CastBarCaption string
	Faction        uint32
	Flags          string
	HasCustomAnim  bool
	Size           float32
	Data           []uint32
	MinGold        econ.Money
	MaxGold        econ.Money
}

type Contact struct {
	Player   uint64 `xorm:"'player' index"`
	Friend   uint64 `xorm:"'friend'"`
	Friended bool   `xorm:"'friended'"`
	Ignored  bool   `xorm:"'ignored'"`
	Muted    bool   `xorm:"'muted'"`
	Note     string `xorm:"'note'"`
}

func (got GameObjectTemplate) GetID() string {
	return got.ID
}

func (got GameObjectTemplate) SetEntry(entry uint32) uint32 {
	return got.Entry
}
