package update

func init5875() {
	dc := NewDescriptorCompiler(5875)

	// class Object
	obj := dc.ObjectBase()
	obj.GUID(ObjectGUID, Public)
	obj.Uint32(ObjectType, Public)
	obj.Uint32(ObjectEntry, Public)
	obj.Float32(ObjectScaleX, Public)
	obj.Pad()

	// class Item : Object
	item := obj.Extend("Item")
	item.GUID(ItemOwner, Public)
	item.GUID(ItemContained, Public)
	item.GUID(ItemCreator, Public)
	item.GUID(ItemGiftCreator, Public)
	item.Uint32(ItemStackCount, Public)
	item.Uint32(ItemDuration, Public)
	item.Uint32Array(ItemSpellCharges, 5, Public)
	item.Uint32(ItemFlags, Public)
	item.Uint32Array(ItemEnchantment, 21, Public)
	item.Uint32(ItemPropertySeed, Public)
	item.Uint32(ItemRandomPropertiesID, Public)
	item.Uint32(ItemDurability, Public)
	item.Uint32(ItemMaxDurability, Public)
	item.Uint32(ItemCreatePlayedTime, Public)
	item.Pad()

	// class Container : Item
	container := item.Extend("Container")
	container.Uint32(ContainerNumSlots, Public)
	container.Uint32(ContainerAlignPad, Public)
	container.GUIDArray(ContainerSlots, 72, Public)

	// class Unit : Object
	unit := obj.Extend("Unit")
	unit.GUID(UnitCharm, Public)
	unit.GUID(UnitSummon, Public)
	unit.GUID(UnitCharmedBy, Public)
	unit.GUID(UnitSummonedBy, Public)
	unit.GUID(UnitCreatedBy, Public)
	unit.GUID(UnitTarget, Public)
	unit.GUID(UnitPersuaded, Public)
	unit.GUID(UnitChannelObject, Public)
	unit.Uint32(UnitHealth, Public)
	unit.Uint32Array(UnitPowers, 5, Public)
	unit.Uint32(UnitMaxHealth, Public)
	unit.Uint32Array(UnitMaxPowers, 5, Public)
	unit.Uint32(UnitLevel, Public)
	unit.Uint32(UnitFactionTemplate, Public)

	unit.Byte(UnitRace, Public)
	unit.Byte(UnitClass, Public)
	unit.Byte(UnitGender, Public)
	unit.Byte(UnitPower, Public)

	unit.Uint32Array(UnitVirtualItemSlotIDs, 3, Public)
	unit.Uint32Array(UnitVirtualItemInfos, 6, Public)
	unit.Uint32(UnitFlags, Public)
	unit.Uint32Array(UnitAuras, 48, Public)
	unit.Uint32Array(UnitAuraFlags, 6, Public)
	unit.Uint32Array(UnitAuraLevels, 12, Public)
	unit.Uint32Array(UnitAuraApplications, 12, Public)
	unit.Uint32(UnitAuraState, Public)
	unit.Uint32(UnitBaseAttackTime, Public)
	unit.Uint32(UnitOffhandAttackTime, Public)
	unit.Uint32(UnitRangedAttackTime, Public)
	unit.Float32(UnitBoundingRadius, Public)
	unit.Float32(UnitCombatReach, Public)
	unit.Uint32(UnitDisplayID, Public)
	unit.Uint32(UnitNativeDisplayID, Public)
	unit.Uint32(UnitMountDisplayID, Public)
	unit.Float32(UnitMinDamage, Public)
	unit.Float32(UnitMaxDamage, Public)
	unit.Uint32(UnitMinOffhandDamage, Public)
	unit.Uint32(UnitMaxOffhandDamage, Public)

	unit.Byte(UnitStandState, Public)
	unit.Byte(UnitLoyaltyLevel, Public)
	unit.Byte(UnitShapeshiftForm, Public)
	unit.Byte(UnitStandMiscFlags, Public)

	unit.Uint32(UnitPetNumber, Public)
	unit.Uint32(UnitPetNameTimestamp, Public)
	unit.Uint32(UnitPetExperience, Public)
	unit.Uint32(UnitPetNextLevelExp, Public)
	unit.Uint32(UnitDynamicFlags, Public)
	unit.Uint32(UnitChannelSpell, Public)
	unit.Float32(UnitModCastSpeed, Public)
	unit.Uint32(UnitCreatedBySpell, Public)
	unit.Uint32(UnitNPCFlags, Public)
	unit.Uint32(UnitNPCEmoteState, Public)
	unit.Uint32(UnitTrainingPoints, Public)
	unit.Uint32Array(UnitStats, 5, Public)
	unit.Uint32Array(UnitResistances, 7, Public)
	unit.Uint32(UnitBaseMana, Public)
	unit.Uint32(UnitBaseHealth, Public)

	unit.Byte(UnitSheathState, Public)
	unit.Byte(UnitAuraByteFlags, Public)
	unit.Byte(UnitPetRename, Public)
	unit.Byte(UnitPetShapeshiftForm, Public)

	unit.Int32(UnitAttackPower, Public)
	unit.Int32(UnitAttackPowerMods, Public)
	unit.Float32(UnitAttackPowerMultiplier, Public)
	unit.Int32(UnitRangedAttackPower, Public)
	unit.Int32(UnitRangedAttackPowerMods, Public)
	unit.Float32(UnitRangedAttackPowerMultiplier, Public)
	unit.Float32(UnitMinRangedDamage, Public)
	unit.Float32(UnitMaxRangedDamage, Public)
	unit.Uint32Array(UnitPowerCostModifier, 7, Public)
	unit.Float32Array(UnitPowerCostMultiplier, 7, Public)
	unit.Pad()

	// class Player : Unit
	plyr := unit.Extend("Player")
	plyr.GUID(PlayerDuelArbiter, Public)
	plyr.Uint32(PlayerFlags, Public)
	plyr.Uint32(PlayerGuildID, Public)
	plyr.Uint32(PlayerGuildRank, Public)

	plyr.Byte(PlayerSkin, Public)
	plyr.Byte(PlayerFace, Public)
	plyr.Byte(PlayerHairStyle, Public)
	plyr.Byte(PlayerHairColor, Public)

	plyr.Byte(PlayerFacialHair, Public)
	plyr.Byte(PlayerRestBits, Public)
	plyr.Byte(PlayerBankBagSlotCount, Public)
	plyr.Byte(PlayerRestState, Public)

	plyr.Byte(PlayerGender, Public)
	plyr.Byte(PlayerGenderUnk, Public)
	plyr.Byte(PlayerDrunkness, Public)
	plyr.Byte(PlayerPVPRank, Public)

	plyr.Uint32(PlayerDuelTeam, Public)
	plyr.Uint32(PlayerGuildTimestamp, Public)

	questLog := plyr.Array(PlayerQuestLog, 20)
	questLog.Uint32("QuestID", Public)
	questLog.Uint32("CountState", Public)
	questLog.Uint32("Time", Public)
	questLog.End()

	visItems := plyr.Array(PlayerVisibleItems, 19)
	visItems.GUID("Creator", Public)
	visItems.Uint32("Entry", Public)
	visItems.Uint32Array("Enchantments", 8, Public)
	visItems.Uint32("Properties", Public)
	visItems.Pad()
	visItems.End()

	plyr.GUIDArray(PlayerInventorySlots, 23, Private)
	plyr.GUIDArray(PlayerPackSlots, 16, Private)
	plyr.GUIDArray(PlayerBankSlots, 24, Private)
	plyr.GUIDArray(PlayerBankBagSlots, 6, Private)
	plyr.GUIDArray(PlayerVendorBuybackSlots, 12, Private)
	plyr.GUIDArray(PlayerKeyringSlots, 32, Private)
	plyr.GUID(PlayerFarSight, Public)
	plyr.GUID(PlayerFieldComboTarget, Public)
	plyr.Uint32(PlayerXP, Public)
	plyr.Uint32(PlayerNextLevelXP, Public)
	plyr.Uint32Array(PlayerSkillInfos, 384, Private)
	plyr.Uint32Array(PlayerCharacterPoints, 2, Private)
	plyr.Uint32(PlayerTrackCreatures, Private)
	plyr.Uint32(PlayerTrackResources, Private)
	plyr.Float32(PlayerBlockPercentage, Public)
	plyr.Float32(PlayerDodgePercentage, Public)
	plyr.Float32(PlayerParryPercentage, Public)
	plyr.Float32(PlayerCritPercentage, Public)
	plyr.Float32(PlayerRangedCritPercentage, Public)
	plyr.Uint32Array(PlayerExploredZones, 64, Public)
	plyr.Uint32(PlayerRestStateExperience, Public)
	plyr.Int32(PlayerCoinage, Private)
	plyr.Uint32Array(UnitPosStats, 5, Public)
	plyr.Uint32Array(UnitNegStats, 5, Public)
	plyr.Uint32Array(UnitResistanceBuffModsPositive, 7, Public)
	plyr.Uint32Array(UnitResistanceBuffModsNegative, 7, Public)
	plyr.Uint32Array(PlayerModDamageDonePositive, 7, Public)
	plyr.Uint32Array(PlayerModDamageDoneNegative, 7, Public)
	plyr.Float32Array(PlayerModDamageDonePercentage, 7, Public)

	plyr.Byte(PlayerFieldBytesFlags, Public)
	plyr.Byte(PlayerFieldBytesUnk1, Public)
	plyr.Byte(PlayerActionBarToggle, Public)
	plyr.Byte(PlayerFieldBytesUnk2, Public)

	plyr.Uint32(PlayerAmmoID, Public)
	plyr.Uint32(PlayerSelfResSpell, Public)
	plyr.Uint32(PlayerPVPMedals, Public)
	plyr.Uint32Array(PlayerBuybackPrices, 12, Private)
	plyr.Uint32Array(PlayerBuybackTimestamps, 12, Private)
	plyr.Uint32(PlayerKills, Public)
	plyr.Uint32(PlayerYesterdayKills, Public)
	plyr.Uint32(PlayerLastWeekKills, Public)
	plyr.Uint32(PlayerThisWeekKills, Public)
	plyr.Uint32(PlayerThisWeekContribution, Public)
	plyr.Uint32(PlayerLifetimeHonorableKills, Public)
	plyr.Uint32(PlayerLifetimeDishonorableKills, Public)
	plyr.Uint32(PlayerYesterdayContribution, Public)
	plyr.Uint32(PlayerLastWeekContribution, Public)
	plyr.Uint32(PlayerLastWeekRank, Public)

	plyr.Byte(PlayerHonorRankPoints, Public)
	plyr.Byte(PlayerDetectionFlags, Public)

	plyr.Int32(PlayerWatchedFactionIndex, Public)
	plyr.Uint32Array(PlayerCombatRatings, 20, Public)

	// class GameObject : Object
	gobj := obj.Extend("GameObject")

	gobj.GUID(GObjectCreatedBy, Public)
	gobj.Uint32(GObjectDisplayID, Public)
	gobj.Uint32(GObjectFlags, Public)
	gobj.Uint32(GObjectRotation, Public)
	gobj.Uint32(GObjectState, Public)
	gobj.Float32(GObjectPosX, Public)
	gobj.Float32(GObjectPosY, Public)
	gobj.Float32(GObjectPosZ, Public)
	gobj.Float32(GObjectFacing, Public)
	gobj.Uint32(GObjectDynamicFlags, Public)
	gobj.Uint32(GObjectFaction, Public)
	gobj.Uint32(GObjectTypeID, Public)
	gobj.Uint32(GObjectLevel, Public)
	gobj.Uint32(GObjectArtKit, Public)
	gobj.Uint32(GObjectAnimProgress, Public)
	gobj.Uint32(GObjectPadding, Public)

	// class DynamicObject : Object
	dobj := obj.Extend("DynamicObject")
	dobj.GUID(DynamicObjectCaster, Public)

	dobj.Byte(DynamicObjectType, Public)

	dobj.Uint32(DynamicObjectSpellID, Public)
	dobj.Float32(DynamicObjectRadius, Public)
	dobj.Float32(DynamicObjectPosX, Public)
	dobj.Float32(DynamicObjectPosY, Public)
	dobj.Float32(DynamicObjectPosZ, Public)
	dobj.Float32(DynamicObjectFacing, Public)

	// class Corpse : Object
	corp := obj.Extend("Corpse")
	corp.GUID(CorpseOwner, Public)
	corp.Float32(CorpseFacing, Public)
	corp.Float32(CorpsePosX, Public)
	corp.Float32(CorpsePosY, Public)
	corp.Float32(CorpsePosZ, Public)
	corp.Uint32(CorpseDisplayID, Public)
	corp.Uint32Array(CorpseItem, 19, Public)

	corp.Byte(CorpsePlayerUnk, Public)
	corp.Byte(CorpseRace, Public)
	corp.Byte(CorpseGender, Public)
	corp.Byte(CorpseSkin, Public)

	corp.Byte(CorpseFace, Public)
	corp.Byte(CorpseHairStyle, Public)
	corp.Byte(CorpseHairColor, Public)
	corp.Byte(CorpseFacialHair, Public)

	corp.Uint32(CorpseGuild, Public)
	corp.Uint32(CorpseFlags, Public)
	corp.Uint32(CorpseDynamicFlags, Public)
	corp.Pad()

	Descriptors[5875] = dc
}

// VDescriptor5875 = ValuesDescriptor{
// 	ObjectGUID:                      {GUID, 0, Public, 0x00},
// 	ObjectType:                      {Uint32, 0, Public, 0x02},
// 	ObjectEntry:                     {Uint32, 0, Public, 0x03},
// 	ObjectScaleX:                    {Float32, 0, Public, 0x04},
// 	ItemOwner:                       {GUID, 0, Public, 0x06 + 0x00},
// 	ItemContained:                   {GUID, 0, Public, 0x06 + 0x02},
// 	ItemCreator:                     {GUID, 0, Public, 0x06 + 0x04},
// 	ItemGiftCreator:                 {GUID, 0, Public, 0x06 + 0x06},
// 	ItemStackCount:                  {Uint32, 0, Public, 0x06 + 0x08},
// 	ItemDuration:                    {Uint32, 0, Public, 0x06 + 0x09},
// 	ItemSpellCharges:                {Uint32, 5, Public, 0x06 + 0x0A},
// 	ItemFlags:                       {Uint32, 0, Public, 0x06 + 0x0F},
// 	ItemEnchantment:                 {Uint32, 21, Public, 0x06 + 0x10},
// 	ItemPropertySeed:                {Uint32, 0, Public, 0x06 + 0x25},
// 	ItemRandomPropertiesID:          {Uint32, 0, Public, 0x06 + 0x26},
// 	ItemDurability:                  {Uint32, 0, Public, 0x06 + 0x27},
// 	ItemMaxDurability:               {Uint32, 0, Public, 0x06 + 0x28},
// 	ItemCreatePlayedTime:            {Uint32, 0, Public, 0x06 + 0x29},
// 	ContainerNumSlots:               {Uint32, 0, Public, 0x06 + 0x2A},
// 	ContainerAlignPad:               {Uint32, 0, Public, 0x06 + 0x2B},
// 	ContainerSlots:                  {GUID, 72, Public, 0x06 + 0x2C},
// 	UnitCharm:                       {GUID, 0, Public, 0x06 + 0x00},
// 	UnitSummon:                      {GUID, 0, Public, 0x06 + 0x02},
// 	UnitCharmedBy:                   {GUID, 0, Public, 0x06 + 0x04},
// 	UnitSummonedBy:                  {GUID, 0, Public, 0x06 + 0x06},
// 	UnitCreatedBy:                   {GUID, 0, Public, 0x06 + 0x08},
// 	UnitTarget:                      {GUID, 0, Public, 0x06 + 0x0A},
// 	UnitPersuaded:                   {GUID, 0, Public, 0x06 + 0x0C},
// 	UnitChannelObject:               {GUID, 0, Public, 0x06 + 0x0E},
// 	UnitHealth:                      {Uint32, 0, Public, 0x06 + 0x10},
// 	UnitPowers:                      {Uint32, 5, Public, 0x06 + 0x11},
// 	UnitMaxHealth:                   {Uint32, 0, Public, 0x06 + 0x16},
// 	UnitMaxPowers:                   {Uint32, 5, Public, 0x06 + 0x17},
// 	UnitLevel:                       {Uint32, 0, Public, 0x06 + 0x1C},
// 	UnitFactionTemplate:             {Uint32, 0, Public, 0x06 + 0x1D},
// 	UnitBytes0:                      {UBytes0, 0, Public, 0x06 + 0x1E},
// 	UnitVirtualItemSlotIDs:          {Uint32, 3, Public, 0x06 + 0x1F},
// 	UnitVirtualItemInfos:            {Uint32, 6, Public, 0x06 + 0x22},
// 	UnitFlags:                       {Uint32, 0, Public, 0x06 + 0x28},
// 	UnitAuras:                       {Uint32, 48, Public, 0x06 + 0x29},
// 	UnitAuraFlags:                   {Uint32, 6, Public, 0x06 + 0x59},
// 	UnitAuraLevels:                  {Uint32, 12, Public, 0x06 + 0x5F},
// 	UnitAuraApplications:            {Uint32, 12, Public, 0x06 + 0x5F},
// 	UnitAuraState:                   {Uint32, 0, Public, 0x06 + 0x77},
// 	UnitBaseAttackTime:              {Uint32, 0, Public, 0x06 + 0x78},
// 	UnitOffhandAttackTime:           {Uint32, 0, Public, 0x06 + 0x79},
// 	UnitRangedAttackTime:            {Uint32, 0, Public, 0x06 + 0x7A},
// 	UnitBoundingRadius:              {Float32, 0, Public, 0x06 + 0x7B},
// 	UnitCombatReach:                 {Float32, 0, Public, 0x06 + 0x7C},
// 	UnitDisplayID:                   {Uint32, 0, Public, 0x06 + 0x7D},
// 	UnitNativeDisplayID:             {Uint32, 0, Public, 0x06 + 0x7E},
// 	UnitMountDisplayID:              {Uint32, 0, Public, 0x06 + 0x7F},
// 	UnitMinDamage:                   {Uint32, 0, Public, 0x06 + 0x80},
// 	UnitMaxDamage:                   {Uint32, 0, Public, 0x06 + 0x81},
// 	UnitMinOffhandDamage:            {Uint32, 0, Public, 0x06 + 0x82},
// 	UnitMaxOffhandDamage:            {Uint32, 0, Public, 0x06 + 0x83},
// 	UnitBytes1:                      {UBytes1, 0, Public, 0x06 + 0x84},
// 	UnitPetNumber:                   {Uint32, 0, Public, 0x06 + 0x85},
// 	UnitPetNameTimestamp:            {Uint32, 0, Public, 0x06 + 0x86},
// 	UnitPetExperience:               {Uint32, 0, Public, 0x06 + 0x87},
// 	UnitPetNextLevelExp:             {Uint32, 0, Public, 0x06 + 0x88},
// 	UnitDynamicFlags:                {Uint32, 0, Public, 0x06 + 0x89},
// 	UnitChannelSpell:                {Uint32, 0, Public, 0x06 + 0x8A},
// 	UnitModCastSpeed:                {Uint32, 0, Public, 0x06 + 0x8B},
// 	UnitCreatedBySpell:              {Uint32, 0, Public, 0x06 + 0x8C},
// 	UnitNPCFlags:                    {Uint32, 0, Public, 0x06 + 0x8D},
// 	UnitNPCEmoteState:               {Uint32, 0, Public, 0x06 + 0x8E},
// 	UnitTrainingPoints:              {Uint32, 0, Public, 0x06 + 0x8F},
// 	UnitStats:                       {Uint32, 5, Public, 0x06 + 0x90},
// 	UnitResistances:                 {Uint32, 7, Public, 0x06 + 0x95},
// 	UnitBaseMana:                    {Uint32, 0, Public, 0x06 + 0x9C},
// 	UnitBaseHealth:                  {Uint32, 0, Public, 0x06 + 0x9D},
// 	UnitBytes2:                      {UBytes2, 0, Public, 0x06 + 0x9E},
// 	UnitAttackPower:                 {Uint32, 0, Public, 0x06 + 0x9F},
// 	UnitAttackPowerMods:             {Uint32, 0, Public, 0x06 + 0xA0},
// 	UnitAttackPowerMultiplier:       {Uint32, 0, Public, 0x06 + 0xA1},
// 	UnitRangedAttackPower:           {Uint32, 0, Public, 0x06 + 0xA2},
// 	UnitRangedAttackPowerMods:       {Uint32, 0, Public, 0x06 + 0xA3},
// 	UnitRangedAttackPowerMultiplier: {Uint32, 0, Public, 0x06 + 0xA4},
// 	UnitMinRangedDamage:             {Uint32, 0, Public, 0x06 + 0xA5},
// 	UnitMaxRangedDamage:             {Uint32, 0, Public, 0x06 + 0xA6},
// 	UnitPowerCostModifier:           {Uint32, 0, Public, 0x06 + 0xA7},
// 	UnitPowerCostMultiplier:         {Uint32, 0, Public, 0x06 + 0xA8},
// 	UnitMaxHealthModifier:           {Uint32, 0, Public, 0x06 + 0xA9},
// 	UnitHoverHeight:                 {Uint32, 0, Public, 0x06 + 0xAA},
// 	PlayerDuelArbiter:               {GUID, 0, Public, 0xB0 + 0x00},
// 	PlayerFlags:                     {Uint32, 0, Public, 0xB0 + 0x02},
// 	PlayerGuildID:                   {Uint32, 0, Public, 0xB0 + 0x03},
// 	PlayerGuildRank:                 {Uint32, 0, Public, 0xB0 + 0x04},
// 	PlayerBytes1:                    {PBytes1, 0, Public, 0xB0 + 0x05},
// 	PlayerBytes2:                    {PBytes2, 0, Public, 0xB0 + 0x06},
// 	PlayerBytes3:                    {PBytes3, 0, Public, 0xB0 + 0x07},
// 	PlayerDuelTeam:                  {Uint32, 0, Public, 0xB0 + 0x08},
// 	PlayerGuildTimestamp:            {Uint32, 0, Public, 0xB0 + 0x09},
// 	PlayerQuestLog:                  {QLog, 20, Private | Party, 0xB0 + 0x0A},
// 	PlayerVisibleItems:              {VItems, 19, Public, 0xB0 + 0x46},
// 	PlayerInventorySlots:            {GUID, 23, Public, 0xB0 + 0x128},
// 	PlayerPackSlots:                 {GUID, 16, Private, 0xB0 + 0x158},
// 	PlayerBankSlots:                 {GUID, 24, Private, 0xB0 + 0x178},
// 	PlayerBankBagSlots:              {GUID, 6, Private, 0xB0 + 0x1A8},
// 	PlayerVendorBuybackSlots:        {GUID, 12, Private, 0xB0 + 0x1B4},
// 	PlayerKeyringSlots:              {GUID, 32, Private, 0xB0 + 0x1CC},
// 	PlayerFarSight:                  {GUID, 0, Public, 0xB0 + 0x20C},
// 	PlayerFieldComboTarget:          {GUID, 0, Public, 0xB0 + 0x20E},
// 	PlayerXP:                        {Uint32, 0, Public, 0xB0 + 0x20C},
// 	PlayerNextLevelXP:               {Uint32, 0, Public, 0xB0 + 0x211},
// 	PlayerSkillInfos:                {Uint32, 384, Public, 0xB0 + 0x212},
// 	PlayerCharacterPoints:           {Uint32, 2, Public, 0xB0 + 0x392},
// 	PlayerTrackCreatures:            {Uint32, 0, Public, 0xB0 + 0x394},
// 	PlayerTrackResources:            {Uint32, 0, Public, 0xB0 + 0x395},
// 	PlayerBlockPercentage:           {Uint32, 0, Public, 0xB0 + 0x396},
// 	PlayerDodgePercentage:           {Uint32, 0, Public, 0xB0 + 0x397},
// 	PlayerParryPercentage:           {Uint32, 0, Public, 0xB0 + 0x398},
// 	PlayerCritPercentage:            {Uint32, 0, Public, 0xB0 + 0x399},
// 	PlayerRangedCritPercentage:      {Uint32, 0, Public, 0xB0 + 0x39A},
// 	PlayerExploredZones:             {Uint32, 64, Private, 0xB0 + 0x39B},
// 	PlayerRestStateExperience:       {Uint32, 0, Public, 0xB0 + 0x3DB},
// 	PlayerCoinage:                   {Uint32, 0, Private, 0xB0 + 0x3DC},
// 	UnitPosStats:                    {Uint32, 5, Private, 0xB0 + 0x3DD},
// 	UnitNegStats:                    {Uint32, 5, Private, 0xB0 + 0x332},
// 	UnitResistanceBuffModsPositive:  {Uint32, 7, Public, 0xB0 + 0x3E7},
// 	UnitResistanceBuffModsNegative:  {Uint32, 7, Public, 0xB0 + 0x3EE},
// 	PlayerModDamageDonePositive:     {Uint32, 7, Private, 0xB0 + 0x3F5},
// 	PlayerModDamageDoneNegative:     {Uint32, 7, Private, 0xB0 + 0x3FC},
// 	PlayerModDamageDonePercentage:   {Uint32, 7, Private, 0xB0 + 0x403},
// 	PlayerFieldBytes:                {PFBytes, 0, Public, 0xB0 + 0x40A},
// 	PlayerAmmoID:                    {Uint32, 0, Public, 0xB0 + 0x40B},
// 	PlayerSelfResSpell:              {Uint32, 0, Public, 0xB0 + 0x40C},
// 	PlayerPVPMedals:                 {Uint32, 0, Public, 0xB0 + 0x40D},
// 	PlayerBuybackPrices:             {Uint32, 12, Public, 0xB0 + 0x40E},
// 	PlayerBuybackTimestamps:         {Uint32, 12, Public, 0xB0 + 0x40B},
// 	PlayerKills:                     {Uint32, 0, Public, 0xB0 + 0x426},
// 	PlayerYesterdayKills:            {Uint32, 0, Public, 0xB0 + 0x427},
// 	PlayerLastWeekKills:             {Uint32, 0, Public, 0xB0 + 0x428},
// 	PlayerThisWeekKills:             {Uint32, 0, Public, 0xB0 + 0x429},
// 	PlayerThisWeekContribution:      {Uint32, 0, Public, 0xB0 + 0x42A},
// 	PlayerLifetimeHonorableKills:    {Uint32, 0, Public, 0xB0 + 0x42B},
// 	PlayerLifetimeDishonorableKills: {Uint32, 0, Public, 0xB0 + 0x42C},
// 	PlayerYesterdayContribution:     {Uint32, 0, Public, 0xB0 + 0x42D},
// 	PlayerLastWeekContribution:      {Uint32, 0, Public, 0xB0 + 0x42E},
// 	PlayerLastWeekRank:              {Uint32, 0, Public, 0xB0 + 0x42F},
// 	PlayerFieldBytes2:               {PFBytes2, 0, Public, 0xB0 + 0x430},
// 	PlayerWatchedFactionIndex:       {Uint32, 0, Public, 0xB0 + 0x431},
// 	PlayerCombatRatings:             {Uint32, 20, Public, 0xB0 + 0x432},

// DynamicObjectCaster
// DynamicObjectBytes
// DynamicObjectSpellID
// DynamicObjectRadius
// DynamicObjectCastTime
// DynamicObjectEnd

// CorpseOwner
// CorpseParty
// CorpseDisplayID
// CorpseItem
// CorpseBytes1
// CorpseBytes2
// CorpseGuild
// CorpseFlags
// CorpseDynamicFlags
// CorpsePad
