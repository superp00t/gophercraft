package main

import (
	"fmt"
	"os"
	"reflect"

	"github.com/superp00t/etc"
	"github.com/superp00t/gophercraft/datapack/csv"
	"github.com/superp00t/gophercraft/packet/update"
	"github.com/superp00t/gophercraft/worldserver/wdb"

	_ "github.com/go-sql-driver/mysql"
	"xorm.io/xorm"
)

type ItemTemplate struct {
	Entry                     int    `xorm:"'entry'"`
	Class                     uint32 `xorm:"'class'"`
	Subclass                  uint32 `xorm:"'subclass'"`
	Name                      string `xorm:"'name'"`
	DisplayID                 uint32 `xorm:"'displayid'"`
	Quality                   uint8  `xorm:"'Quality'"`
	Flags                     uint32 `xorm:"'Flags'"`
	BuyCount                  uint8  `xorm:"'BuyCount'"`
	BuyPrice                  uint32 `xorm:"'BuyPrice'"`
	SellPrice                 uint32 `xorm:"'SellPrice'"`
	InventoryType             uint8  `xorm:"'InventoryType'"`
	AllowableClass            int32  `xorm:"'AllowableClass'"`
	AllowableRace             int32  `xorm:"'AllowableRace'"`
	ItemLevel                 uint32 `xorm:"'ItemLevel'"`
	RequiredLevel             uint8  `xorm:"'RequiredLevel'"`
	RequiredSkill             uint32 `xorm:"'RequiredSkill'"`
	RequiredSkillRank         uint32 `xorm:"'RequiredSkillRank'"`
	Requiredspell             uint32 `xorm:"'requiredspell'"`
	Requiredhonorrank         uint32 `xorm:"'requiredhonorrank'"`
	RequiredCityRank          uint32 `xorm:"'RequiredCityRank'"`
	RequiredReputationFaction uint32 `xorm:"'RequiredReputationFaction'"`
	RequiredReputationRank    uint32 `xorm:"'RequiredReputationRank'"`
	Maxcount                  uint32 `xorm:"'maxcount'"`
	Stackable                 uint32 `xorm:"'stackable'"`
	ContainerSlots            uint8  `xorm:"'ContainerSlots'"`
	// StatsCount                uint8   `xorm:"'StatsCount'"`
	Stat_type1   uint8 `xorm:"'stat_type1'"`
	Stat_value1  int32 `xorm:"'stat_value1'"`
	Stat_type2   uint8 `xorm:"'stat_type2'"`
	Stat_value2  int32 `xorm:"'stat_value2'"`
	Stat_type3   uint8 `xorm:"'stat_type3'"`
	Stat_value3  int32 `xorm:"'stat_value3'"`
	Stat_type4   uint8 `xorm:"'stat_type4'"`
	Stat_value4  int32 `xorm:"'stat_value4'"`
	Stat_type5   uint8 `xorm:"'stat_type5'"`
	Stat_value5  int32 `xorm:"'stat_value5'"`
	Stat_type6   uint8 `xorm:"'stat_type6'"`
	Stat_value6  int32 `xorm:"'stat_value6'"`
	Stat_type7   uint8 `xorm:"'stat_type7'"`
	Stat_value7  int32 `xorm:"'stat_value7'"`
	Stat_type8   uint8 `xorm:"'stat_type8'"`
	Stat_value8  int32 `xorm:"'stat_value8'"`
	Stat_type9   uint8 `xorm:"'stat_type9'"`
	Stat_value9  int32 `xorm:"'stat_value9'"`
	Stat_type10  uint8 `xorm:"'stat_type10'"`
	Stat_value10 int32 `xorm:"'stat_value10'"`
	// ScalingStatDistribution int32   `xorm:"'ScalingStatDistribution'"`
	// ScalingStatValue        int32   `xorm:"'ScalingStatValue'"`
	DMG_min1                float32 `xorm:"'dmg_min1'"`
	DMG_max1                float32 `xorm:"'dmg_max1'"`
	DMG_type1               uint8   `xorm:"'dmg_type1'"`
	DMG_min2                float32 `xorm:"'dmg_min2'"`
	DMG_max2                float32 `xorm:"'dmg_max2'"`
	DMG_type2               uint8   `xorm:"'dmg_type2'"`
	DMG_min3                float32 `xorm:"'dmg_min3'"`
	DMG_max3                float32 `xorm:"'dmg_max3'"`
	DMG_type3               uint8   `xorm:"'dmg_type3'"`
	DMG_min4                float32 `xorm:"'dmg_min4'"`
	DMG_max4                float32 `xorm:"'dmg_max4'"`
	DMG_type4               uint8   `xorm:"'dmg_type4'"`
	DMG_min5                float32 `xorm:"'dmg_min5'"`
	DMG_max5                float32 `xorm:"'dmg_max5'"`
	DMG_type5               uint8   `xorm:"'dmg_type5'"`
	Armor                   uint32  `xorm:"'armor'"`
	Holy_res                uint32  `xorm:"'holy_res'"`
	Fire_res                uint32  `xorm:"'fire_res'"`
	Nature_res              uint32  `xorm:"'nature_res'"`
	Frost_res               uint32  `xorm:"'frost_res'"`
	Shadow_res              uint32  `xorm:"'shadow_res'"`
	Arcane_res              uint32  `xorm:"'arcane_res'"`
	Delay                   uint32  `xorm:"'delay'"`
	Ammo_type               uint32  `xorm:"'ammo_type'"`
	RangedModRange          float32 `xorm:"'RangedModRange'"`
	Spellid_1               uint32  `xorm:"'spellid_1'"`
	Spelltrigger_1          uint32  `xorm:"'spelltrigger_1'"`
	Spellcharges_1          int     `xorm:"'spellcharges_1'"`
	SpellppmRate_1          float32 `xorm:"'spellppmRate_1'"`
	Spellcooldown_1         int32   `xorm:"'spellcooldown_1'"`
	Spellcategory_1         uint32  `xorm:"'spellcategory_1'"`
	Spellcategorycooldown_1 int32   `xorm:"'spellcategorycooldown_1'"`
	Spellid_2               uint32  `xorm:"'spellid_2'"`
	Spelltrigger_2          uint32  `xorm:"'spelltrigger_2'"`
	Spellcharges_2          int32   `xorm:"'spellcharges_2'"`
	SpellppmRate_2          float32 `xorm:"'spellppmRate_2'"`
	Spellcooldown_2         int32   `xorm:"'spellcooldown_2'"`
	Spellcategory_2         uint32  `xorm:"'spellcategory_2'"`
	Spellcategorycooldown_2 int32   `xorm:"'spellcategorycooldown_2'"`
	Spellid_3               uint32  `xorm:"'spellid_3'"`
	Spelltrigger_3          uint32  `xorm:"'spelltrigger_3'"`
	Spellcharges_3          int32   `xorm:"'spellcharges_3'"`
	SpellppmRate_3          float32 `xorm:"'spellppmRate_3'"`
	Spellcooldown_3         int32   `xorm:"'spellcooldown_3'"`
	Spellcategory_3         uint32  `xorm:"'spellcategory_3'"`
	Spellcategorycooldown_3 int32   `xorm:"'spellcategorycooldown_3'"`
	Spellid_4               uint32  `xorm:"'spellid_4'"`
	Spelltrigger_4          uint32  `xorm:"'spelltrigger_4'"`
	Spellcharges_4          int32   `xorm:"'spellcharges_4'"`
	SpellppmRate_4          float32 `xorm:"'spellppmRate_4'"`
	Spellcooldown_4         int32   `xorm:"'spellcooldown_4'"`
	Spellcategory_4         uint32  `xorm:"'spellcategory_4'"`
	Spellcategorycooldown_4 int32   `xorm:"'spellcategorycooldown_4'"`
	Spellid_5               uint32  `xorm:"'spellid_5'"`
	Spelltrigger_5          uint32  `xorm:"'spelltrigger_5'"`
	Spellcharges_5          int32   `xorm:"'spellcharges_5'"`
	SpellppmRate_5          float32 `xorm:"'spellppmRate_5'"`
	Spellcooldown_5         int32   `xorm:"'spellcooldown_5'"`
	Spellcategory_5         uint32  `xorm:"'spellcategory_5'"`
	Spellcategorycooldown_5 int32   `xorm:"'spellcategorycooldown_5'"`
	Bonding                 uint8   `xorm:"'bonding'"`
	Description             string  `xorm:"'description'"`
	PageText                uint32  `xorm:"'PageText'"`
	LanguageID              uint32  `xorm:"'LanguageID'"`
	PageMaterial            uint32  `xorm:"'PageMaterial'"`
	Startquest              uint32  `xorm:"'startquest'"`
	Lockid                  uint32  `xorm:"'lockid'"`
	Material                int32   `xorm:"'Material'"`
	Sheath                  uint32  `xorm:"'sheath'"`
	RandomProperty          uint32  `xorm:"'RandomProperty'"`
	// RandomSuffix            uint32  `xorm:"'RandomSuffix'"`
	Block         uint32 `xorm:"'block'"`
	Itemset       uint32 `xorm:"'itemset'"`
	MaxDurability uint32 `xorm:"'MaxDurability'"`
	Area          uint32 `xorm:"'area'"`
	Map           int32  `xorm:"'Map'"`
	BagFamily     int32  `xorm:"'BagFamily'"`
	// TotemCategory           int32   `xorm:"'TotemCategory'"`
	// SocketColor_1           int32   `xorm:"'socketColor_1'"`
	// SocketContent_1         int32   `xorm:"'socketContent_1'"`
	// SocketColor_2           int32   `xorm:"'socketColor_2'"`
	// SocketContent_2         int32   `xorm:"'socketContent_2'"`
	// SocketColor_3           int32   `xorm:"'socketColor_3'"`
	// SocketContent_3         int32   `xorm:"'socketContent_3'"`
	// SocketBonus             int32   `xorm:"'socketBonus'"`
	// GemProperties           int32   `xorm:"'GemProperties'"`
	// RequiredDisenchantSkill int32   `xorm:"'RequiredDisenchantSkill'"`
	// ArmorDamageModifier float32 `xorm:"'ArmorDamageModifier'"`
	// ItemLimitCategory int32  `xorm:"'ItemLimitCategory'"`
	ScriptName   string `xorm:"'ScriptName'"`
	DisenchantID uint32 `xorm:"'DisenchantID'"`
	FoodType     uint8  `xorm:"'FoodType'"`
	MinMoneyLoot uint32 `xorm:"'minMoneyLoot'"`
	MaxMoneyLoot uint32 `xorm:"'maxMoneyLoot'"`
	Duration     int32  `xorm:"'Duration'"`
	ExtraFlags   uint8  `xorm:"'ExtraFlags'"`
}

func (it ItemTemplate) GetItemStat(idx int) wdb.ItemStat {
	st := fmt.Sprintf("Stat_type%d", idx)
	sv := fmt.Sprintf("Stat_value%d", idx)
	val := reflect.ValueOf(it)
	_, ok := val.Type().FieldByName(st)
	if !ok {
		panic("no field for " + st)
	}

	return wdb.ItemStat{
		uint8(val.FieldByName(st).Uint()),
		int32(val.FieldByName(sv).Int()),
	}
}

func (it ItemTemplate) GetItemDamage(idx int) wdb.ItemDamage {
	dt := fmt.Sprintf("DMG_type%d", idx)
	dmn := fmt.Sprintf("DMG_min%d", idx)
	dmx := fmt.Sprintf("DMG_max%d", idx)

	val := reflect.ValueOf(it)
	_, ok := val.Type().FieldByName(dt)
	if !ok {
		panic("no field for " + dt)
	}

	return wdb.ItemDamage{
		uint8(val.FieldByName(dt).Uint()),
		float32(val.FieldByName(dmn).Float()),
		float32(val.FieldByName(dmx).Float()),
	}
}

func (it ItemTemplate) GetItemSpell(idx int) wdb.ItemSpell {
	st := fmt.Sprintf("Spellid_%d", idx)
	strigger := fmt.Sprintf("Spelltrigger_%d", idx)
	scharges := fmt.Sprintf("Spellcharges_%d", idx)
	ppm := fmt.Sprintf("SpellppmRate_%d", idx)
	cool := fmt.Sprintf("Spellcooldown_%d", idx)
	scategory := fmt.Sprintf("Spellcategory_%d", idx)
	scategoryCooldown := fmt.Sprintf("Spellcategorycooldown_%d", idx)

	val := reflect.ValueOf(it)
	_, ok := val.Type().FieldByName(st)
	if !ok {
		panic("no field for " + st)
	}

	return wdb.ItemSpell{
		uint32(val.FieldByName(st).Uint()),
		uint32(val.FieldByName(strigger).Uint()),
		int32(val.FieldByName(scharges).Int()),
		float32(val.FieldByName(ppm).Float()),
		int64(val.FieldByName(cool).Int()),
		uint32(val.FieldByName(scategory).Uint()),
		int64(val.FieldByName(scategoryCooldown).Int()),
	}
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println(os.Args[0], "<xorm url>")
		return
	}

	outdir := etc.Import("github.com/superp00t/gophercraft/datapack/contrib/DB")

	c, err := xorm.NewEngine("mysql", os.Args[1])
	if err != nil {
		panic(err)
	}

	var itt []ItemTemplate
	err = c.Find(&itt)
	if err != nil {
		panic(err)
	}

	outFile, err := os.OpenFile(outdir.Concat("ItemTemplate.5875.csv").Render(), os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0700)
	if err != nil {
		panic(err)
	}

	wr := csv.NewWriter(outFile)

	for _, t := range itt {
		var stats []wdb.ItemStat

		for i := 0; i < 10; i++ {
			st := t.GetItemStat(i + 1)
			if st.Type != 0 {
				stats = append(stats, st)
			}
		}

		var dmg []wdb.ItemDamage
		for i := 0; i < 5; i++ {
			d := t.GetItemDamage(i + 1)
			if d.Type != 0 {
				dmg = append(dmg, d)
			}
		}

		var spell []wdb.ItemSpell
		for i := 0; i < 5; i++ {
			sp := t.GetItemSpell(i + 1)
			if sp.ID != 0 {
				spell = append(spell, sp)
			}
		}

		flags, err := update.DecodeItemFlagInteger(5875, uint64(t.Flags))
		if err != nil {
			panic(err)
		}

		witem := wdb.ItemTemplate{
			ID:                        fmt.Sprintf("ent:%d", t.Entry),
			Class:                     t.Class,
			Subclass:                  t.Subclass,
			Name:                      t.Name,
			DisplayID:                 t.DisplayID,
			Quality:                   t.Quality,
			Flags:                     flags.String(), //todo: convert flags to a readable form
			BuyCount:                  t.BuyCount,
			BuyPrice:                  t.BuyPrice,
			SellPrice:                 t.SellPrice,
			InventoryType:             t.InventoryType,
			AllowableClass:            t.AllowableClass,
			AllowableRace:             t.AllowableRace,
			ItemLevel:                 t.ItemLevel,
			RequiredLevel:             t.RequiredLevel,
			RequiredSkill:             t.RequiredSkill,
			RequiredSkillRank:         t.RequiredSkillRank,
			RequiredSpell:             t.Requiredspell,
			RequiredHonorRank:         t.Requiredhonorrank,
			RequiredCityRank:          t.RequiredCityRank,
			RequiredReputationFaction: t.RequiredReputationFaction,
			MaxCount:                  t.Maxcount,
			Stackable:                 t.Stackable,
			ContainerSlots:            t.ContainerSlots,
			Stats:                     stats,
			Damage:                    dmg,
			Armor:                     t.Armor,
			HolyRes:                   t.Holy_res,
			FireRes:                   t.Fire_res,
			NatureRes:                 t.Nature_res,
			FrostRes:                  t.Frost_res,
			ShadowRes:                 t.Shadow_res,
			ArcaneRes:                 t.Arcane_res,
			Delay:                     t.Delay,
			AmmoType:                  t.Ammo_type,
			RangedModRange:            t.RangedModRange,
			Spells:                    spell,
			Bonding:                   t.Bonding,
			Description:               t.Description,
			PageText:                  t.PageText,
			LanguageID:                t.LanguageID,
			PageMaterial:              t.PageMaterial,
			StartQuest:                t.Startquest,
			LockID:                    t.Lockid,
			Material:                  t.Material,
			Sheath:                    t.Sheath,
			RandomProperty:            t.RandomProperty,
			Block:                     t.Block,
			Itemset:                   t.Itemset,
			MaxDurability:             t.MaxDurability,
			Area:                      t.Area,
			Map:                       t.Map,
			BagFamily:                 t.BagFamily,
			ScriptName:                t.ScriptName,
			DisenchantID:              t.DisenchantID,
			FoodType:                  t.FoodType,
			MinMoneyLoot:              t.MinMoneyLoot,
			MaxMoneyLoot:              t.MaxMoneyLoot,
			Duration:                  t.Duration,
			ExtraFlags:                t.ExtraFlags,
		}
		if err := wr.AddRecord(witem); err != nil {
			panic(err)
		}

	}

	wr.Close()
}
