// Descriptor module for version 8606 2.4.3 (TBC)
package d8606

import (
	"reflect"

	"github.com/superp00t/gophercraft/guid"
	"github.com/superp00t/gophercraft/packet/update"
	"github.com/superp00t/gophercraft/vsn"
)

type ObjectData struct {
	GUID   guid.GUID
	Type   uint32
	Entry  uint32
	ScaleX float32
	update.ChunkPad
}

type ItemData struct {
	Owner              guid.GUID
	Contained          guid.GUID
	Creator            guid.GUID
	GiftCreator        guid.GUID
	StackCount         uint32
	Duration           uint32
	SpellCharges       [5]int32
	Flags              uint32
	Enchantment        [33]uint32
	PropertySeed       uint32
	RandomPropertiesID uint32
	TextID             uint32
	Durability         uint32
	MaxDurability      uint32
}

type ContainerData struct {
	NumSlots uint32
	AlignPad uint32
	Slots    [36]guid.GUID
}

type UnitData struct {
	Charm              guid.GUID
	Summon             guid.GUID
	CharmedBy          guid.GUID
	SummonedBy         guid.GUID
	CreatedBy          guid.GUID
	Target             guid.GUID
	Persuaded          guid.GUID
	ChannelObject      guid.GUID
	Health             uint32
	Mana               uint32
	Rage               uint32
	Focus              uint32
	Energy             uint32
	Happiness          uint32
	MaxHealth          uint32
	MaxMana            uint32
	MaxRage            uint32
	MaxFocus           uint32
	MaxEnergy          uint32
	MaxHappiness       uint32
	Level              uint32
	FactionTemplate    uint32
	Race               uint8
	Class              uint8
	Gender             uint8
	Power              uint8
	VirtualItemSlotIDs [3]uint32
	VirtualItemInfos   [6]uint32

	// Unit flags
	ServerControlled    bool // 0x1
	NonAttackable       bool // 0x2
	RemoveClientControl bool // 0x4
	PlayerControlled    bool // 0x8
	Rename              bool // 0x10
	PetAbandon          bool // 0x20
	Unk6                bool // 0x40
	Unk7                bool // 0x80
	OOCNotAttackable    bool // 0x100
	Passive             bool // 0x200
	Unk10               bool // 0x400
	Unk11               bool // 0x800
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
	Unk28               bool // 0x10000000
	Unk29               bool // 0x20000000
	Sheathe             bool // 0x40000000
	NoKillReward        bool // 0x80000000

	// Unit flags 2
	FeignDeath          bool // 0x1
	HideBodyArmor       bool // 0x2
	IgnoreReputation    bool // 0x4
	ComprehendLanguage  bool // 0x8
	Cloned              bool // 0x10
	UnitFlagUnk5        bool // 0x20
	ForceMove           bool // 0x40
	DisarmOffhand       bool // 0x80
	UnitFlagUnk8        bool // 0x100
	UnitFlagUnk9        bool // 0x200
	DisarmRanged        bool // 0x400
	RegeneratePower     bool // 0x800
	SpellClickInGroup   bool // 0x1000
	SpellClickDisabled  bool // 0x2000
	InteractAnyReaction bool // 0x4000
	PadUnitFlags2       update.ChunkPad

	Auras             [56]uint32
	AuraFlags         [56]byte
	AuraLevels        [56]byte
	AuraApplications  [56]byte
	AuraState         uint32
	BaseAttackTime    uint32
	OffhandAttackTime uint32
	RangedAttackTime  uint32
	BoundingRadius    float32
	CombatReach       float32
	DisplayID         uint32
	NativeDisplayID   uint32
	MountDisplayID    uint32
	MinDamage         float32
	MaxDamage         float32
	MinOffhandDamage  uint32
	MaxOffhandDamage  uint32
	StandState        uint8
	LoyaltyLevel      uint8
	ShapeshiftForm    uint8
	StandMiscFlags    uint8
	PetNumber         uint32
	PetNameTimestamp  uint32
	PetExperience     uint32
	PetNextLevelExp   uint32

	Lootable              bool
	TrackUnit             bool
	Tapped                bool
	TappedByPlayer        bool
	SpecialInfo           bool
	VisuallyDead          bool
	ReferAFriend          bool
	TappedByAllThreatList bool
	EndUnitDynamicFlags   update.ChunkPad

	ChannelSpell   uint32
	ModCastSpeed   float32
	CreatedBySpell uint32
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
	EndNPCFlags    update.ChunkPad

	NPCEmoteState                  uint32
	TrainingPoints                 uint32
	Stats                          [5]uint32
	UnitPosStats                   [5]uint32
	UnitNegStats                   [5]uint32
	Resistances                    [7]uint32
	UnitResistanceBuffModsPositive [7]uint32
	UnitResistanceBuffModsNegative [7]uint32
	BaseMana                       uint32
	BaseHealth                     uint32
	SheathState                    uint8
	AuraByteFlags                  uint8
	PetRename                      uint8
	PetShapeshiftForm              uint8
	AttackPower                    int32
	AttackPowerMods                int32
	AttackPowerMultiplier          float32
	RangedAttackPower              int32
	RangedAttackPowerMods          int32
	RangedAttackPowerMultiplier    float32
	MinRangedDamage                float32
	MaxRangedDamage                float32
	PowerCostModifier              [7]uint32
	PowerCostMultiplier            [7]float32
	MaxHealthModifier              float32
	update.ChunkPad
}

type PlayerData struct {
	DuelArbiter    guid.GUID
	GroupLeader    bool
	AFK            bool
	DND            bool
	GM             bool
	Ghost          bool
	Resting        bool
	VoiceChat      bool
	FFAPVP         bool
	ContestedPVP   bool
	PVPDesired     bool
	HideHelm       bool
	HideCloak      bool
	PlayedLongTime bool
	PlayedTooLong  bool
	OutOfBounds    bool
	GhostEffect    bool
	Sanctuary      bool
	TaxiBenchmark  bool
	PVPTimer       bool
	PadPlayerFlags update.ChunkPad

	GuildID          uint32
	GuildRank        uint32
	Skin             uint8
	Face             uint8
	HairStyle        uint8
	HairColor        uint8
	FacialHair       uint8
	RestBits         uint8
	BankBagSlotCount uint8
	RestState        uint8
	PlayerGender     uint8
	GenderUnk        uint8
	Drunkness        uint8
	PVPRank          uint8
	DuelTeam         uint32
	GuildTimestamp   uint32
	QuestLog         [25]struct {
		QuestID    uint32
		CountState uint32
		QuestUnk   uint32
		Time       uint32
	}

	VisibleItems [19]struct {
		Creator      guid.GUID
		Entry        uint32
		Enchantments [11]uint32
		Properties   uint32
		update.ChunkPad
	}

	ChosenTitle uint32

	StartSlotPad       update.ChunkPad
	InventorySlots     [39]guid.GUID `update:"private"`
	BankSlots          [28]guid.GUID `update:"private"`
	BankBagSlots       [7]guid.GUID  `update:"private"`
	VendorBuybackSlots [12]guid.GUID `update:"private"`
	KeyringSlots       [32]guid.GUID `update:"private"`
	VanityPetSlots     [18]guid.GUID `update:"private"`
	FarSight           guid.GUID
	KnownTitles        [2]uint32
	XP                 uint32
	NextLevelXP        uint32
	SkillInfos         [128]struct {
		ID         uint16
		Step       uint16
		SkillLevel uint16
		SkillCap   uint16
		Bonus      uint32
	} `update:"private"`
	CharacterPoints             [2]uint32 `update:"private"`
	TrackCreatures              uint32    `update:"private"`
	TrackResources              uint32    `update:"private"`
	BlockPercentage             float32
	DodgePercentage             float32
	ParryPercentage             float32
	Expertise                   uint32
	OffhandExpertise            uint32
	CritPercentage              float32
	RangedCritPercentage        float32
	OffhandCritPercentage       float32
	SpellCritPercentage         [7]float32
	ShieldBlock                 uint32
	ExploredZones               [128]uint32 // TODO: use Bitmask type with length tag to refer to this field.
	RestStateExperience         uint32
	Coinage                     int32 `update:"private"`
	ModDamageDonePositive       [7]uint32
	ModDamageDoneNegative       [7]uint32
	ModDamageDonePercentage     [7]float32
	ModHealingDonePos           uint32
	ModTargetResistance         uint32
	ModTargetPhysicalResistance uint32
	// Flags
	PlayerFieldBytes0UnkBit0      update.BitPad // (unknown value)
	TrackStealthed                bool
	DisplaySpiritAutoReleaseTimer bool
	HideSpiritReleaseWindow       bool
	RAFGrantableLevel             uint8 // parser should automatically frame this to next byte.
	ActionBarToggles              uint8
	LifetimeMaxPVPRank            uint8

	AmmoID                 uint32
	SelfResSpell           uint32
	PVPMedals              uint32
	BuybackPrices          [12]uint32 `update:"private"`
	BuybackTimestamps      [12]uint32 `update:"private"`
	Kills                  uint32
	TodayKills             uint32
	YesterdayKills         uint32
	LifetimeHonorableKills uint32
	HonorRankPoints        uint8
	DetectionFlagUnk       bool
	DetectAmore0           bool
	DetectAmore1           bool
	DetectAmore2           bool
	DetectAmore3           bool
	DetectStealth          bool
	DetectInvisibilityGlow bool

	WatchedFactionIndex int32
	CombatRatings       [24]uint32
	ArenaTeamInfo       [18]uint32
	HonorCurrency       uint32
	ArenaCurrency       uint32

	ModManaRegen          float32
	ModManaRegenInterrupt float32
	MaxLevel              uint32
	DailyQuests           [25]uint32
}

type GameObjectData struct {
	CreatedBy    guid.GUID
	DisplayID    uint32
	Flags        uint32
	Rotation     [4]float32
	State        uint32
	PosX         float32
	PosY         float32
	PosZ         float32
	Facing       float32
	DynamicFlags uint32
	Faction      uint32
	TypeID       uint32
	Level        uint32
	ArtKit       uint32
	AnimProgress uint32
	Padding      uint32
}

type DynamicObjectData struct {
	Caster   guid.GUID
	Type     uint8
	SpellID  uint32
	Radius   float32
	PosX     float32
	PosY     float32
	PosZ     float32
	Facing   float32
	CastTime uint32
}

type CorpseData struct {
	Owner        guid.GUID
	Party        guid.GUID
	Facing       float32
	PosX         float32
	PosY         float32
	PosZ         float32
	DisplayID    uint32
	Item         [19]uint32
	PlayerUnk    uint8
	Race         uint8
	Gender       uint8
	Skin         uint8
	Face         uint8
	HairStyle    uint8
	HairColor    uint8
	FacialHair   uint8
	Guild        uint32
	Flags        uint32
	DynamicFlags uint32
	update.ChunkPad
}

type ItemDescriptor struct {
	ObjectData
	ItemData
}

type ContainerDescriptor struct {
	ObjectData
	ItemData
	ContainerData
}

type UnitDescriptor struct {
	ObjectData
	UnitData
}

type PlayerDescriptor struct {
	ObjectData
	UnitData
	PlayerData
}

type GameObjectDescriptor struct {
	ObjectData
	GameObjectData
}

type DynamicObjectDescriptor struct {
	ObjectData
	DynamicObjectData
}

type CorpseDescriptor struct {
	ObjectData
	CorpseData
}

func init() {
	update.Descriptors[vsn.V2_4_3] = &update.Descriptor{
		vsn.V2_4_3,
		update.DescriptorOptionClassicGUIDs | update.DescriptorOptionHasHasTransport,
		map[guid.TypeMask]reflect.Type{
			guid.TypeMaskObject: reflect.TypeOf(ObjectData{}),

			guid.TypeMaskObject | guid.TypeMaskItem:                          reflect.TypeOf(ItemDescriptor{}),
			guid.TypeMaskObject | guid.TypeMaskItem | guid.TypeMaskContainer: reflect.TypeOf(ContainerDescriptor{}),
			guid.TypeMaskObject | guid.TypeMaskDynamicObject:                 reflect.TypeOf(DynamicObjectDescriptor{}),
			guid.TypeMaskObject | guid.TypeMaskGameObject:                    reflect.TypeOf(GameObjectDescriptor{}),
			guid.TypeMaskObject | guid.TypeMaskCorpse:                        reflect.TypeOf(CorpseDescriptor{}),
			guid.TypeMaskObject | guid.TypeMaskUnit:                          reflect.TypeOf(UnitDescriptor{}),
			guid.TypeMaskObject | guid.TypeMaskUnit | guid.TypeMaskPlayer:    reflect.TypeOf(PlayerDescriptor{}),
		},
	}
}
