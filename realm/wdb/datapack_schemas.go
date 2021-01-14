package wdb

import (
	"github.com/superp00t/gophercraft/econ"
	"github.com/superp00t/gophercraft/i18n"
	"github.com/superp00t/gophercraft/packet/update"
)

// PlayerCreateInfo determines where a character is spawned at upon their first login, using their race and class.
type PlayerCreateInfo struct {
	Race  uint8
	Class uint8
	Map   uint32
	Zone  uint32
	X     float32
	Y     float32
	Z     float32
	O     float32
}

type PlayerCreateItem struct {
	Race   int8
	Class  int8
	ItemID string
	Amount uint32
}

type PlayerCreateAbility struct {
	Race   int8
	Class  int8
	Spell  uint32
	Note   string
	Active bool
}

type PlayerCreateActionButton struct {
	Race   int8
	Class  int8
	Button uint8
	Action uint32
	Type   uint8
	Misc   uint8
}

type LevelExperience map[uint32]uint32

type PortLocation struct {
	ID  string  `xorm:"'port_id' pk" csv:"name"`
	X   float32 `xorm:"'x_pos'" csv:"x"`
	Y   float32 `xorm:"'y_pos'" csv:"y"`
	Z   float32 `xorm:"'z_pos'" csv:"z"`
	O   float32 `xorm:"'orientation'" csv:"orientation	"`
	Map uint32  `xorm:"'map'" csv:"mapID"`
}

type CreatureTemplate struct {
	ID              string
	Entry           uint32    `xorm:"'entry'"`
	Name            i18n.Text `xorm:"'name'"`
	SubName         i18n.Text `xorm:"'sub_name'"`
	MinLevel        uint32    `xorm:"'min_level'"`
	MaxLevel        uint32    `xorm:"'max_level'"`
	DisplayIDs      []uint32  `xorm:"'display_ids'"`
	Faction         uint32    `xorm:"'faction'"`
	Scale           float32   `xorm:"'scale'"`
	Family          string    `xorm:"'family'"`
	CreatureType    uint32    `xorm:"'creature_type'"`
	InhabitType     uint32    `xorm:"'inhabit_type'"`
	RegenerateStats uint32    `xorm:"'regenerateStats'"`
	RacialLeader    bool      `xorm:"'racialLeader'"`
	// NpcFlags: should not tied to any particular version.
	Gossip         bool
	QuestGiver     bool
	Vendor         bool
	FlightMaster   bool
	Trainer        bool
	SpiritHealer   bool
	SpiritGuide    bool
	Innkeeper      bool
	Banker         bool
	Petitioner     bool
	TabardDesigner bool
	BattleMaster   bool
	Auctioneer     bool
	StableMaster   bool
	Repairer       bool
	OutdoorPVP     bool

	// UnitFlags: should not tied to any particular version.
	ServerControlled    bool // 0x1
	NonAttackable       bool // 0x2
	RemoveClientControl bool // 0x4
	PlayerControlled    bool // 0x8
	Rename              bool // 0x10
	PetAbandon          bool // 0x20
	OOCNotAttackable    bool // 0x100
	Passive             bool // 0x200
	PVP                 bool // 0x1000
	IsSilenced          bool // 0x2000
	IsPersuaded         bool // 0x4000
	Swimming            bool // 0x8000
	RemoveAttackIcon    bool // 0x10000
	IsPacified          bool // 0x20000
	IsStunned           bool // 0x40000
	InCombat            bool // 0x80000
	InTaxiFlight        bool // 0x100000
	Disarmed            bool // 0x200000
	Confused            bool // 0x400000
	Fleeing             bool // 0x800000
	Possessed           bool // 0x1000000
	NotSelectable       bool // 0x2000000
	Skinnable           bool // 0x4000000
	AurasVisible        bool // 0x8000000
	Sheathe             bool // 0x40000000
	NoKillReward        bool // 0x80000000

	// DynamicFlags: should not tied to any particular version.
	Lootable              bool
	TrackUnit             bool
	Tapped                bool
	TappedByPlayer        bool
	SpecialInfo           bool
	VisuallyDead          bool
	TappedByAllThreatList bool

	// Extra flags
	InstanceBind             bool // creature kill bind instance with killer and killer’s group
	NoAggroOnSight           bool // no aggro (ignore faction/reputation hostility)
	NoParry                  bool // creature can’t parry
	NoParryHasten            bool // creature can’t counter-attack at parry
	NoBlock                  bool //	creature can’t block
	NoCrush                  bool // creature can’t do crush attacks
	NoXPAtKill               bool // creature kill not provide XP
	Invisible                bool // creature is always invisible for player (mostly trigger creatures)
	NotTauntable             bool // creature is immune to taunt auras and effect attack me
	AggroZone                bool // creature sets itself in combat with zone on aggro
	Guard                    bool // creature is a guard
	NoCallAssist             bool // creature shouldn’t call for assistance on aggro
	Active                   bool //creature is active object. Grid of this creature will be loaded and creature set as active
	ForceEnableMMap          bool // creature is forced to use MMaps
	ForceDisableMMap         bool // creature is forced to NOT use MMaps
	WalkInWater              bool // creature is forced to walk in water even it can swim
	Civilian                 bool // CreatureInfo→civilian substitute (for expansions as Civilian Colum was removed)
	NoMelee                  bool // creature can’t melee
	FarView                  bool // creature with far view
	ForceAttackingCapability bool // SetForceAttackingCapability(true); for nonattackable, nontargetable creatures that should be able to attack nontheless
	IgnoreUsedPosition       bool // ignore creature when checking used positions around target
	CountSpawns              bool // count creature spawns in Map*
	HasteSpellImmunity       bool // immunity to COT or Mind Numbing Poison – very common in instances

	// CreatureTypeFlags    uint32     `xorm:"'creatureTypeFlags'"`
	Tameable                bool // Makes the mob tameable (must also be a beast and have family set)
	VisibleToGhosts         bool // Sets Creatures that can ALSO be seen when player is a ghost. Used in CanInteract function by client, can’t be attacked
	BossLevel               bool
	DontPlayWoundParryAnim  bool
	HideFactionTooltip      bool // Controls something in client tooltip related to creature faction
	SpellAttackable         bool
	DeadInteract            bool // Player can interact with the creature if its dead (not player dead)
	HerbLoot                bool // Uses Skinning Loot Field
	MiningLoot              bool // Makes Mob Corpse Mineable – Uses Skinning Loot Field
	DontLogDeath            bool // Does not combatlog death.
	MountedCombat           bool
	CanAssist               bool // Can aid any player or group in combat. Typically seen for escorting NPC’s
	PetHasActionBar         bool // checked from calls in Lua_PetHasActionBar
	MaskUID                 bool
	EngineerLoot            bool // Makes Mob Corpse Engineer Lootable – Uses Skinning Loot Field
	ExoticPet               bool // Tamable as an exotic pet. Normal tamable flag must also be set.
	UseDefaultCollisionBox  bool
	IsSiegeWeapon           bool
	ProjectileCollision     bool
	HideNameplate           bool
	DontPlayMountedAnim     bool
	IsLinkAll               bool
	InteractOnlyWithCreator bool
	ForceGossip             bool

	SpeedWalk            float32    `xorm:"'speedWalk'"`
	SpeedRun             float32    `xorm:"'speedRun'"`
	UnitClass            uint32     `xorm:"'unitClass'"`
	Rank                 uint32     `xorm:"'rank'"`
	HealthMultiplier     float32    `xorm:"'healthMultiplier'"`
	PowerMultiplier      float32    `xorm:"'powerMultiplier'"`
	DamageMultiplier     float32    `xorm:"'damageMultiplier'"`
	DamageVariance       float32    `xorm:"'damageVariance'"`
	ArmorMultiplier      float32    `xorm:"'armorMultiplier'"`
	ExperienceMultiplier float32    `xorm:"'experienceMultiplier'"`
	MinLevelHealth       uint32     `xorm:"'minLevelHealth'"`
	MaxLevelHealth       uint32     `xorm:"'maxLevelHealth'"`
	MinLevelMana         uint32     `xorm:"'minLevelMana'"`
	MaxLevelMana         uint32     `xorm:"'maxLevelMana'"`
	MinMeleeDmg          float32    `xorm:"'minMeleeDmg'"`
	MaxMeleeDmg          float32    `xorm:"'maxMeleeDmg'"`
	MinRangedDmg         float32    `xorm:"'minRangedDmg'"`
	MaxRangedDmg         float32    `xorm:"'maxRangedDmg'"`
	Armor                uint32     `xorm:"'armor'"`
	MeleeAttackPower     uint32     `xorm:"'meleeAttackPower'"`
	RangedAttackPower    uint32     `xorm:"'rangedAttackPower'"`
	MeleeBaseAttackTime  uint32     `xorm:"'meleeBaseAttackTime'"`
	RangedBaseAttackTime uint32     `xorm:"'mangedBaseAttackTime'"`
	DamageSchool         int32      `xorm:"'damageSchool'"`
	MinLootGold          econ.Money `xorm:"'minLootGold'"`
	MaxLootGold          econ.Money `xorm:"'maxLootGold'"`
	LootId               uint32     `xorm:"'lootId'"`
	PickpocketLootId     uint32     `xorm:"'pickpocketLootId'"`
	SkinningLootId       uint32     `xorm:"'skinningLootId'"`
	KillCredit1          uint32     `xorm:"'killCredit1'"`
	KillCredit2          uint32     `xorm:"'killCredit2'"`
	MechanicImmuneMask   uint32     `xorm:"'mechanicImmuneMask'"`
	SchoolImmuneMask     uint32     `xorm:"'schoolImmuneMask'"`
	ResistanceHoly       int32      `xorm:"'resistanceHoly'"`
	ResistanceFire       int32      `xorm:"'resistanceFire'"`
	ResistanceNature     int32      `xorm:"'resistanceNature'"`
	ResistanceFrost      int32      `xorm:"'resistanceFrost'"`
	ResistanceShadow     int32      `xorm:"'resistanceShadow'"`
	ResistanceArcane     int32      `xorm:"'resistanceArcane'"`
	PetSpellDataId       uint32     `xorm:"'petSpellDataId'"`
	MovementType         uint32     `xorm:"'movementType'"`
	TrainerType          int32      `xorm:"'trainerType'"`
	TrainerSpell         uint32     `xorm:"'trainerSpell'"`
	TrainerClass         uint32     `xorm:"'trainerClass'"`
	TrainerRace          uint32     `xorm:"'trainerRace'"`
	TrainerTemplateId    uint32     `xorm:"'trainerTemplateId'"`
	VendorTemplateId     uint32     `xorm:"'vendorTemplateId'"`
	GossipMenuId         string     `xorm:"'gossipMenuId'"`
	EquipmentTemplateId  uint32     `xorm:"'equipmentTemplateId'"`
	DishonourableKill    bool       `xorm:"'dishonourable_kill'"`
	AIName               string     `xorm:"'aIName'"`
	ScriptName           string     `xorm:"'script_name'"`
}

type ItemTemplate struct {
	Entry                     uint32          `json:",omitempty" xorm:"'entry' bigint pk" csv:"-"`
	ID                        string          `json:",omitempty" xorm:"'id' index"`
	Name                      i18n.Text       `json:",omitempty" xorm:"'name' index"`
	Class                     uint32          `json:",omitempty" xorm:"'class'"`
	Subclass                  uint32          `json:",omitempty" xorm:"'subclass'"`
	DisplayID                 uint32          `json:",omitempty" xorm:"'display_id'"`
	Quality                   uint8           `json:",omitempty" xorm:"'quality'"`
	Flags                     update.ItemFlag `json:",omitempty" xorm:"'flags'"`
	BuyCount                  uint8           `json:",omitempty" xorm:"'buy_count'"`
	BuyPrice                  uint32          `json:",omitempty" xorm:"'buy_price'"`
	SellPrice                 uint32          `json:",omitempty" xorm:"'sell_price'"`
	InventoryType             uint8           `json:",omitempty" xorm:"'inv_type'"`
	AllowableClass            int32           `json:",omitempty" xorm:"'allowable_class'"`
	AllowableRace             int32           `json:",omitempty" xorm:"'allowable_race'"`
	ItemLevel                 uint32          `json:",omitempty" xorm:"'item_level'"`
	RequiredLevel             uint8           `json:",omitempty" xorm:"'required_level'"`
	RequiredSkill             uint32          `json:",omitempty" xorm:"'required_skill'"`
	RequiredSkillRank         uint32          `json:",omitempty" xorm:"'required_skill_rank'"`
	RequiredSpell             uint32          `json:",omitempty" xorm:"'required_spell'"`
	RequiredHonorRank         uint32          `json:",omitempty" xorm:"'required_honor_rank'"`
	RequiredCityRank          uint32          `json:",omitempty" xorm:"'required_city_rank'"`
	RequiredReputationFaction uint32          `json:",omitempty" xorm:"'required_reputation_faction'"`
	RequiredReputationRank    uint32          `json:",omitempty" xorm:"'required_reputation_rank'"`
	MaxCount                  uint32          `json:",omitempty" xorm:"'max_count'"`
	Stackable                 uint32          `json:",omitempty" xorm:"'stackable'"`
	ContainerSlots            uint8           `json:",omitempty" xorm:"'container_slots'"`
	Stats                     []ItemStat      `json:",omitempty" xorm:"'stats'"`
	Damage                    []ItemDamage    `json:",omitempty" xorm:"'dmg'"`
	Armor                     uint32          `json:",omitempty" xorm:"'armor'"`
	HolyRes                   uint32          `json:",omitempty" xorm:"'holy_res'"`
	FireRes                   uint32          `json:",omitempty" xorm:"'fire_res'"`
	NatureRes                 uint32          `json:",omitempty" xorm:"'nature_res'"`
	FrostRes                  uint32          `json:",omitempty" xorm:"'frost_res'"`
	ShadowRes                 uint32          `json:",omitempty" xorm:"'shadow_res'"`
	ArcaneRes                 uint32          `json:",omitempty" xorm:"'arcane_res'"`
	Delay                     uint32          `json:",omitempty" xorm:"'delay'"`
	AmmoType                  uint32          `json:",omitempty" xorm:"'ammo_type'"`
	RangedModRange            float32         `json:",omitempty" xorm:"'ranged_mod_range'"`
	Spells                    []ItemSpell     `json:",omitempty" xorm:"'spells'"`
	Bonding                   uint8           `json:",omitempty" xorm:"'bonding'"`
	Description               i18n.Text       `json:",omitempty" xorm:"'description' longtext"`
	PageText                  uint32          `json:",omitempty" xorm:"'page_text"`
	LanguageID                uint32          `json:",omitempty" xorm:"'language_id'"`
	PageMaterial              uint32          `json:",omitempty" xorm:"'page_material'"`
	StartQuest                uint32          `json:",omitempty" xorm:"'start_quest'"`
	LockID                    uint32          `json:",omitempty" xorm:"'lock_id'"`
	Material                  int32           `json:",omitempty" xorm:"'material'"`
	Sheath                    uint32          `json:",omitempty" xorm:"'sheath'"`
	RandomProperty            uint32          `json:",omitempty" xorm:"'random_property'"`
	RandomSuffix              uint32          `json:",omitempty" xorm:"'random_suffix'"`
	Block                     uint32          `json:",omitempty" xorm:"'block'"`
	Itemset                   uint32          `json:",omitempty" xorm:"'itemset'"`
	MaxDurability             uint32          `json:",omitempty" xorm:"'max_durability'"`
	Area                      uint32          `json:",omitempty" xorm:"'area'"`
	Map                       int32           `json:",omitempty" xorm:"'map'"`
	BagFamily                 int32           `json:",omitempty" xorm:"'bag_family'"`
	TotemCategory             int32           `json:",omitempty" xorm:"'totem_category'"`
	Socket                    []ItemSocket    `json:",omitempty" xorm:"'sockets'"`
	SocketBonus               uint32
	GemProperties             int32      `json:",omitempty" xorm:"'gem_properties'"`
	RequiredDisenchantSkill   int32      `json:",omitempty" xorm:"'required_disenchant_skill'"`
	ArmorDamageModifier       float32    `json:",omitempty" xorm:"'armor_damage_modifier'"`
	ItemLimitCategory         int32      `json:",omitempty" xorm:"'item_limit_category'"`
	ScriptName                string     `json:",omitempty" xorm:"'script_name'"`
	DisenchantID              uint32     `json:",omitempty" xorm:"'disenchant_id'"`
	FoodType                  uint8      `json:",omitempty" xorm:"'food_type'"`
	MinMoneyLoot              econ.Money `json:",omitempty" xorm:"'min_money_loot'"`
	MaxMoneyLoot              econ.Money `json:",omitempty" xorm:"'max_money_loot'"`
	Duration                  int32      `json:",omitempty" xorm:"'duration'"`
	ExtraFlags                uint8      `json:",omitempty" xorm:"'extra_flags'"`
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
	Name           i18n.Text
	IconName       string
	CastBarCaption string
	Faction        uint32
	Flags          update.GameObjectFlags
	HasCustomAnim  bool
	Size           float32
	Data           []uint32
	MinGold        econ.Money
	MaxGold        econ.Money
}

type NPCTextOptionEmote struct {
	Delay uint32
	ID    uint32
}

type NPCTextOption struct {
	Text  i18n.Text
	Lang  uint32
	Prob  float32
	Emote []NPCTextOptionEmote
}

type NPCText struct {
	ID    string
	Entry uint32
	Opts  []NPCTextOption
}

type LocString struct {
	ID   string
	Text i18n.Text
}

type Map struct {
	ID           uint32
	Directory    string
	InstanceType uint32
	MapType      uint32
	Name         string
	MinLevel     uint32
	MaxLevel     uint32
	MaxPlayers   uint32
	Descriptions []string
}

func (m Map) GetDirectory() string {
	if m.Directory != "" {
		return m.Directory
	}

	return m.Name
}
