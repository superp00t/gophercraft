package dbc

type Ent_Spell struct {
	ID                            uint32
	SchoolMask                    uint32
	Category                      uint32
	UICastType                    int32
	Dispel                        uint32
	DispelType                    uint32
	Mechanic                      uint32
	Attributes                    []uint32 `dbc:"(len:5)"`
	ShapeshiftMask                uint32
	ShapeshiftExclude             uint32
	Targets                       uint32
	TargetCreatureType            uint32
	RequiresSpellFocus            uint32
	CasterAuraState               uint32
	TargetAuraState               uint32
	CastingTimeIndex              uint32
	RecoveryTime                  uint32
	CategoryRecoveryTime          uint32
	InterruptFlags                uint32
	AuraInterruptFlags            uint32
	ChannelInterruptFlags         uint32
	ProcFlags                     uint32
	ProcChance                    uint32
	ProcCharges                   uint32
	MaximumLevel                  uint32
	BaseLevel                     uint32
	SpellLevel                    uint32
	DurationIndex                 uint32
	PowerType                     uint32
	ManaCost                      uint32
	ManaCostPerLevel              uint32
	ManaPerSecond                 uint32
	ManaPerSecondPerLevel         uint32
	RangeIndex                    uint32
	Speed                         float32
	ModalNextSpell                uint32
	StackAmount                   uint32
	Totem1                        uint32
	Totem2                        uint32
	Reagents                      []int32 `dbc:"(len:8)"`
	ReagentCounts                 []int32 `dbc:"(len:8)"`
	EquippedItemSubClassMask      int32
	EquippedItemInventoryTypeMask int32
	Effects                       []uint32  `dbc:"(len:3)"`
	EffectDieSides                []int32   `dbc:"(len:3)"`
	EffectBaseDice                []int32   `dbc:"(len:3)"`
	EffectDicePerLevel            []int32   `dbc:"(len:3)"`
	EffectRealPointsPerLevel      []int32   `dbc:"(len:3)"`
	EffectBasePoints              []int32   `dbc:"(len:3)"`
	EffectMechanics               []uint32  `dbc:"(len:3)"`
	EffectImplicitTargetsA        []uint32  `dbc:"(len:3)"`
	EffectImplicitTargetsB        []uint32  `dbc:"(len:3)"`
	EffectRadiusIndices           []uint32  `dbc:"(len:3)"`
	EffectApplyAuraNames          []uint32  `dbc:"(len:3)"`
	EffectAmplitudes              []uint32  `dbc:"(len:3)"`
	EffectMultipleValues          []float32 `dbc:"(len:3)"`
	EffectChainTargets            []uint32  `dbc:"(len:3)"`
	EffectItemTypes               []uint32  `dbc:"(len:3)"`
	EffectMiscValues              []int32   `dbc:"(len:3)"`
	EffectTriggerSpells           []uint32  `dbc:"(len:3)"`
	EffectPointsPerComboPoint     []float32 `dbc:"(len:3)"`
	SpellVisuals                  []uint32  `dbc:"(len:2)"`
	SpellIconID                   uint32
	ActiveIconID                  uint32
	SpellPriority                 uint32
	SpellName                     string `dbc:"(loc)"`
	SpellRank                     string `dbc:"(loc)"`
	SpellDescription              string `dbc:"(loc)"`
	SpellToolTip                  string `dbc:"(loc)"`
	ManaCostPercentage            uint32
	StartRecoveryCategory         uint32
	StartRecoveryTime             uint32
	MaximumTargetLevel            uint32
	SpellClassSet                 uint32
	SpellClassMask                []uint32 `dbc:"(len:2)"`
	MaximumAffectedTargets        uint32
	DamageClass                   uint32
	PreventionType                uint32
	StanceBarOrder                uint32
	EffectDamageMultipliers       []float32 `dbc:"(len:3)"`
	MinimumFactionId              uint32
	MinimumReputation             uint32
	RequiredAuraVision            uint32
}
