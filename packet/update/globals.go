//Package update provides an interface serializing/deserializing SMSG_UPDATE_OBJECT, a packet which handles many crucial aspects to game synchronization.
//Instead of mapping codes for a single game version, this package uses multiple "descriptor modules", which describe the offsets that Globals resolve to.
package update

// Global contains unique references to object properties serialized in multiple versions of SMSG_UPDATE_OBJECT:

//go:generate gcraft_stringer -type=Global -fromString

type Global uint32

const (
	ObjectStart Global = iota
	ObjectGUID
	ObjectType
	ObjectEntry
	ObjectScaleX
	ObjectPadding
	ObjectEnd
	// Item Fields
	ItemStart
	ItemOwner
	ItemContained
	ItemCreator
	ItemGiftCreator
	ItemStackCount
	ItemDuration
	ItemSpellCharges
	ItemFlags
	ItemEnchantment
	ItemPropertySeed
	ItemRandomPropertiesID
	ItemDurability
	ItemMaxDurability
	ItemCreatePlayedTime
	ItemPadding
	ItemEnd
	// Containers
	ContainerNumSlots
	ContainerAlignPad
	ContainerSlots
	ContainerEnd
	// Unit Fields (NPCs+Players)
	UnitStart
	UnitCharm
	UnitSummon
	UnitCritter
	UnitCharmedBy
	UnitSummonedBy
	UnitCreatedBy
	UnitTarget
	UnitPersuaded // not present in 12340
	UnitChannelObject
	UnitChannelSpell
	UnitRace // UNIT_FIELD_BYTES_0
	UnitClass
	UnitGender
	UnitPower
	UnitHealth
	UnitPowers // array<7
	UnitMaxHealth
	UnitMaxPowers // array<7
	UnitPowerRegenFlatModifier
	UnitPowerRegenInterruptedFlatModifier
	UnitLevel
	UnitFactionTemplate
	UnitVirtualItemSlotIDs // array<uint32, 3>
	UnitVirtualItemInfos
	UnitFlags
	UnitFlags2
	UnitAuras
	UnitAuraFlags
	UnitAuraLevels
	UnitAuraApplications
	UnitAuraState
	UnitBaseAttackTime
	UnitOffhandAttackTime
	UnitRangedAttackTime
	UnitBoundingRadius
	UnitCombatReach
	UnitDisplayID
	UnitNativeDisplayID
	UnitMountDisplayID
	UnitMinDamage
	UnitMaxDamage
	UnitMinOffhandDamage
	UnitMaxOffhandDamage
	// UNIT_FIELD_BYTES_1 {
	UnitStandState
	UnitLoyaltyLevel
	UnitShapeshiftForm
	UnitStandMiscFlags
	// }
	UnitPetNumber
	UnitPetNameTimestamp
	UnitPetExperience
	UnitPetNextLevelExp
	UnitDynamicFlags
	UnitModCastSpeed
	UnitCreatedBySpell
	UnitNPCFlags
	UnitNPCEmoteState
	UnitTrainingPoints
	UnitStats    // array<uint32, 5>
	UnitPosStats // array<uint32, 5>
	UnitNegStats // array<uint32, 5>
	UnitResistances
	UnitResistanceBuffModsPositive
	UnitResistanceBuffModsNegative
	UnitBaseMana
	UnitBaseHealth
	// UNIT_FIELD_BYTES_2 {
	UnitSheathState
	UnitAuraByteFlags
	UnitPetRename
	UnitPetShapeshiftForm
	// }
	UnitAttackPower
	UnitAttackPowerMods
	UnitAttackPowerMultiplier
	UnitRangedAttackPower
	UnitRangedAttackPowerMods
	UnitRangedAttackPowerMultiplier
	UnitMinRangedDamage
	UnitMaxRangedDamage
	UnitPowerCostModifier
	UnitPowerCostMultiplier
	UnitMaxHealthModifier
	UnitHoverHeight
	PlayerDuelArbiter
	PlayerFlags
	PlayerGuildID
	PlayerGuildRank
	// PLAYER_BYTES_1 {
	PlayerSkin
	PlayerFace
	PlayerHairStyle
	PlayerHairColor
	// } PLAYER_BYTES_2 {
	PlayerFacialHair
	PlayerRestBits
	PlayerBankBagSlotCount
	PlayerRestState
	// } PLAYER_BYTES_3 {
	PlayerGender
	PlayerGenderUnk
	PlayerDrunkness
	PlayerPVPRank
	// }
	PlayerDuelTeam
	PlayerGuildTimestamp
	PlayerQuestLog     // array
	PlayerVisibleItems // array
	PlayerChosenTitle
	PlayerFakeInebriation
	PlayerInventorySlots // array<guid, 23>
	PlayerPackSlots      //
	PlayerBankSlots
	PlayerBankBagSlots
	PlayerVendorBuybackSlots
	PlayerKeyringSlots
	PlayerCurrencyTokenSlots
	PlayerFarSight
	PlayerFieldComboTarget
	PlayerKnownTitles
	PlayerKnownCurrencies
	PlayerXP
	PlayerNextLevelXP
	PlayerSkillInfos
	PlayerCharacterPoints
	PlayerTrackCreatures
	PlayerTrackResources
	PlayerBlockPercentage
	PlayerDodgePercentage
	PlayerParryPercentage
	PlayerExpertise
	PlayerOffhandExpertise
	PlayerCritPercentage
	PlayerRangedCritPercentage
	PlayerOffhandCritPercentage
	PlayerSpellCritPercentage
	PlayerShieldBlock
	PlayerShieldBlockCritPercentage
	PlayerExploredZones
	PlayerRestStateExperience
	PlayerCoinage
	PlayerModDamageDonePositive
	PlayerModDamageDoneNegative
	PlayerModDamageDonePercentage
	PlayerModHealingDonePositive
	PlayerModHealingPercentage
	PlayerModHealingDonePercentage
	PlayerModTargetResistance
	PlayerModTargetPhysicalResistance
	PlayerFieldBytes
	// PLAYER_FIELD_BYTES {
	PlayerFieldBytesFlags
	PlayerFieldBytesUnk1
	PlayerActionBarToggle
	PlayerFieldBytesUnk2
	// }
	PlayerAmmoID
	PlayerSelfResSpell
	PlayerPVPMedals
	PlayerBuybackPrices
	PlayerBuybackTimestamps
	PlayerKills
	PlayerYesterdayKills
	PlayerLastWeekKills
	PlayerThisWeekKills
	PlayerThisWeekContribution
	PlayerTodayContribution
	PlayerYesterdayContribution
	PlayerLastWeekContribution
	PlayerLastWeekRank
	PlayerLifetimeHonorableKills
	PlayerLifetimeDishonorableKills
	// PLAYER_FIELD_BYTES2 {
	PlayerHonorRankPoints
	PlayerDetectionFlags
	// }
	PlayerWatchedFactionIndex
	PlayerCombatRatings
	PlayerArenaTeamInfos
	PlayerHonorCurrency
	PlayerArenaCurrency
	PlayerMaxLevel
	PlayerDailyQuests
	PlayerRuneRegens
	PlayerNoReagentCosts
	PlayerGlyphSlots
	PlayerGlyphs
	PlayerGlyphsEnabled
	PlayerPetSpellPower
	PlayerEnd

	GObjectCreatedBy
	GObjectDisplayID
	GObjectFlags
	GObjectRotation
	GObjectState
	GObjectPosX
	GObjectPosY
	GObjectPosZ
	GObjectFacing
	GObjectDynamicFlags
	GObjectFaction
	GObjectTypeID
	GObjectLevel
	GObjectArtKit
	GObjectAnimProgress
	GObjectPadding

	DynamicObjectCaster

	DynamicObjectType

	DynamicObjectPosX
	DynamicObjectPosY
	DynamicObjectPosZ
	DynamicObjectFacing

	DynamicObjectSpellID
	DynamicObjectRadius
	DynamicObjectCastTime
	DynamicObjectEnd

	CorpseOwner
	CorpseFacing
	CorpsePosX
	CorpsePosY
	CorpsePosZ

	CorpseParty
	CorpseDisplayID
	CorpseItem

	// CORPSE_FIELD_BYTES_1 {
	CorpsePlayerUnk
	CorpseRace
	CorpseGender
	CorpseSkin
	// } CORPSE_FIELD_BYTES_2 {
	CorpseFace
	CorpseHairStyle
	CorpseHairColor
	CorpseFacialHair
	// }

	CorpseGuild
	CorpseFlags
	CorpseDynamicFlags
	CorpsePad
)
