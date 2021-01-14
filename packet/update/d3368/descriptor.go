// Descriptor module for version 3368 (Alpha)
package d3368

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
	Owner        guid.GUID
	Contained    guid.GUID
	Creator      guid.GUID
	StackCount   uint32
	Duration     uint32
	SpellCharges [5]int32
	Flags        uint32
	Enchantment  [21]uint32
	update.ChunkPad
}

type ContainerData struct {
	NumSlots uint32
	AlignPad uint32
	Slots    [20]guid.GUID
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
	MaxHealth          uint32
	MaxMana            uint32
	MaxRage            uint32
	MaxFocus           uint32
	MaxEnergy          uint32
	Level              uint32
	FactionTemplate    uint32
	Race               uint8
	Class              uint8
	Gender             uint8
	Power              uint8
	Stats              [5]uint32
	BaseStats          [5]uint32
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
	Resting             bool // 0x80
	OOCNotAttackable    bool // 0x100
	Passive             bool // 0x200
	Looting             bool // 0x400
	Unk11               bool // 0x800
	MountIcon           bool // 0x1000
	Mount               bool // 0x2000
	Dead                bool // 0x4000
	Sneak               bool // 0x8000
	Ghost               bool // 0x10000
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

	Coinage                    int32
	Auras                      [56]uint32
	AuraFlags                  [7 * 4]byte
	AuraState                  uint32
	ModDamageDonePositive      [6]uint32
	ModDamageDoneNegative      [6]uint32
	ModDamageDonePercentage    [8]float32
	BaseAttackTime             uint32
	OffhandAttackTime          uint32
	Resistances                [6]uint32
	BoundingRadius             float32
	CombatReach                float32
	WeaponReach                float32
	DisplayID                  uint32
	MountDisplayID             uint32
	Damage                     float32
	ResistanceBuffModsPositive [6]uint32
	ResistanceBuffModsNegative [6]uint32
	ResistanceItemMods         [6]uint32
	StandState                 uint8
	LoyaltyLevel               uint8
	ShapeshiftForm             uint8
	StandMiscFlags             uint8
	PetNumber                  uint32
	PetNameTimestamp           uint32
	PetExperience              uint32
	PetNextLevelExp            uint32
	DynamicFlags               uint32
	EmoteState                 uint32
	ChannelSpell               uint32
	ModCastSpeed               float32
	CreatedBySpell             uint32
	ComboPoints                uint8
	AuraByteFlags              uint8
	FieldBytes2Unk             update.ChunkPad
	EndPadding                 update.ChunkPad
}

type PlayerData struct {
	InventorySlots [39]guid.GUID `update:"private"`
	BankSlots      [24]guid.GUID `update:"private"`
	BankBagSlots   [6]guid.GUID  `update:"private"`
	Selection      guid.GUID
	FarSight       guid.GUID
	DuelArbiter    guid.GUID
	NumInvSlots    uint32
	GuildID        uint32
	GuildRank      uint32
	Skin           uint8
	Face           uint8
	HairStyle      uint8
	HairColor      uint8
	XP             uint32
	NextLevelXP    uint32
	SkillInfos     [64]struct {
		ID         uint16
		Step       uint16
		SkillLevel uint16
		SkillCap   uint16
		Bonus      uint32
	} `update:"private"`
	PlayerFlags      uint8
	FacialHair       uint8
	BankBagSlotCount uint8
	RestState        uint8
	QuestLog         [32]struct {
		QuestID    uint32
		CountState uint32
		Time       uint32
	}
	CharacterPoints [2]uint32 `update:"private"`
	TrackCreatures  uint32    `update:"private"`
	TrackResources  uint32    `update:"private"`
	ChatFilters     uint32
	BlockPercentage float32
	DodgePercentage float32
	ParryPercentage float32
	BaseMana        uint32
	GuildTimestamp  uint32
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
	Caster  guid.GUID
	Type    uint8
	SpellID uint32
	Radius  float32
	PosX    float32
	PosY    float32
	PosZ    float32
	Facing  float32
}

type CorpseData struct {
	Owner        guid.GUID
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
	update.Descriptors[vsn.Alpha] = &update.Descriptor{
		vsn.Alpha,
		update.DescriptorOptionClassicGUIDs | update.DescriptorOptionAlpha,
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
