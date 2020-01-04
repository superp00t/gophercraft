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

type ItemEntries struct {
	ID      string `xorm:"'id' pk"`
	EntryID uint32 `xorm:"'entry_id'"`
}

type ItemTemplate struct {
	ID                        string    `xorm:"'id' pk"`
	Class                     uint32    `xorm:"'class'"`
	Subclass                  uint32    `xorm:"'subclass'"`
	Name                      string    `xorm:"'name'"`
	DisplayID                 uint32    `xorm:"'display_id'"`
	Quality                   uint8     `xorm:"'quality'"`
	Flags                     uint64    `xorm:"'flags'"`
	BuyCount                  uint8     `xorm:"'buy_count'"`
	BuyPrice                  uint32    `xorm:"'buy_price'"`
	SellPrice                 uint32    `xorm:"'sell_price'"`
	InventoryType             uint8     `xorm:"'inv_type'"`
	AllowableClass            int32     `xorm:"'allowable_class'"`
	AllowableRace             int32     `xorm:"'allowable_race'"`
	ItemLevel                 uint32    `xorm:"'item_level'"`
	RequiredLevel             uint8     `xorm:"'required_level'"`
	RequiredSkill             uint32    `xorm:"'required_skill'"`
	RequiredSkillRank         uint32    `xorm:"'required_skill_rank'"`
	RequiredSpell             uint32    `xorm:"'required_spell'"`
	RequiredHonorRank         uint32    `xorm:"'required_honor_rank'"`
	RequiredCityRank          uint32    `xorm:"'required_city_rank'"`
	RequiredReputationFaction uint32    `xorm:"'required_eputation_faction'"`
	RequiredReputationRank    uint32    `xorm:"'required_eputation_rank'"`
	MaxCount                  uint32    `xorm:"'max_count'"`
	Stackable                 uint32    `xorm:"'stackable'"`
	ContainerSlots            uint8     `xorm:"'container_slots'"`
	StatTypes                 []uint8   `xorm:"'stat_types'"`
	StatValues                []uint32  `xorm:"'stat_values'"`
	DamageTypes               []uint8   `xorm:"'dmg_types"`
	DamageMin                 []float32 `xorm:"'dmg_min'"`
	DamageMax                 []float32 `xorm:"'dmg_max'"`
	Armor                     uint32    `xorm:"'armor'"`
	HolyRes                   uint32    `xorm:"'holy_res'"`
	FireRes                   uint32    `xorm:"'fire_res'"`
	NatureRes                 uint32    `xorm:"'nature_res'"`
	FrostRes                  uint32    `xorm:"'frost_res'"`
	ShadowRes                 uint32    `xorm:"'shadow_res'"`
	ArcaneRes                 uint32    `xorm:"'arcane_res'"`
	Delay                     uint32    `xorm:"'delay'"`
	AmmoType                  uint32    `xorm:"'ammo_type'"`
	RangedModRange            float32   `xorm:"'ranged_mod_range'"`
	SpellIDs                  []uint32  `xorm:"'spell_ids'"`
	SpellTriggers             []uint32  `xorm:"'spell_triggers'"`
	SpellCharges              []int     `xorm:"'spell_charges'"`
	SpellPPMRates             []float32 `xorm:"'spell_ppm_rates'"`
	SpellCooldowns            []int32   `xorm:"'spell_cooldowns'"`
	SpellCategories           []uint32  `xorm:"'spell_categories'"`
	SpellCategoryCooldowns    []int32   `xorm:"'spell_category_cooldowns'"`
	Bonding                   uint8     `xorm:"'bonding'"`
	Description               string    `xorm:"'description' longtext"`
	PageText                  uint32    `xorm:"'page_text"`
	LanguageID                uint32    `xorm:"'language_id'"`
	PageMaterial              uint32    `xorm:"'page_material'"`
	Startquest                uint32    `xorm:"'startquest'"`
	Lockid                    uint32    `xorm:"'lockid'"`
	Material                  int32     `xorm:"'material'"`
	Sheath                    uint32    `xorm:"'sheath'"`
	RandomProperty            uint32    `xorm:"'random_property'"`
	RandomSuffix              uint32    `xorm:"'random_suffix'"`
	Block                     uint32    `xorm:"'block'"`
	Itemset                   uint32    `xorm:"'itemset'"`
	MaxDurability             uint32    `xorm:"'max_durability'"`
	Area                      uint32    `xorm:"'area'"`
	Map                       int32     `xorm:"'map'"`
	BagFamily                 int32     `xorm:"'bag_family'"`
	TotemCategory             int32     `xorm:"'totem_category'"`
	SocketColors              []int32   `xorm:"'socket_colors'"`
	SocketContent             []int32   `xorm:"'socket_content'"`
	GemProperties             int32     `xorm:"'gem_properties'"`
	RequiredDisenchantSkill   int32     `xorm:"'required_disenchant_skill'"`
	ArmorDamageModifier       float32   `xorm:"'armor_damage_modifier'"`
	ItemLimitCategory         int32     `xorm:"'item_limit_category'"`
	ScriptName                string    `xorm:"'script_name'"`
	DisenchantID              uint32    `xorm:"'disenchant_id'"`
	FoodType                  uint8     `xorm:"'food_type'"`
	MinMoneyLoot              uint32    `xorm:"'min_money_loot'"`
	MaxMoneyLoot              uint32    `xorm:"'max_money_loot'"`
	Duration                  int32     `xorm:"'duration'"`
	ExtraFlags                uint8     `xorm:"'extra_flags'"`
}
