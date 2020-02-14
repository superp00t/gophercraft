package worldserver

import (
	"fmt"
	"sort"

	"github.com/superp00t/etc"
	"github.com/superp00t/etc/yo"
	"github.com/superp00t/gophercraft/format/dbc"
	"github.com/superp00t/gophercraft/guid"
	"github.com/superp00t/gophercraft/packet"
	"github.com/superp00t/gophercraft/packet/update"
	"github.com/superp00t/gophercraft/worldserver/wdb"
)

// this code collates the inventory slots based on order in the inventory
type InventoryHeap []wdb.Inventory

func (ih InventoryHeap) Swap(i, j int) {
	_i := ih[i]
	_j := ih[j]
	ih[j] = _i
	ih[i] = _j
}

func (ih InventoryHeap) Less(i, j int) bool {
	_i := ih[i]
	_j := ih[j]

	// Backpack always takes precedence to exterior bags
	if _i.Bag == 255 && _j.Bag != 255 {
		return true
	}

	if _j.Bag == 255 && _i.Bag != 255 {
		return false
	}

	if _i.Bag == 255 && _j.Bag == 255 {
		return _i.Slot < _j.Slot
	}

	if _i.Bag != _j.Bag {
		return _i.Bag < _j.Bag
	}

	if _i.Bag == _j.Bag {
		return _i.Slot < _j.Slot
	}

	panic("should never be reached")
	return false
}

func (ih InventoryHeap) Len() int {
	return len(ih)
}

type Item struct {
	ItemID string
	*update.ValuesBlock
}

func (i *Item) Entry() uint32 {
	return i.GetUint32Value(update.ObjectEntry)
}

func (i *Item) GUID() guid.GUID {
	if i == nil {
		return guid.Nil
	}

	return i.GetGUIDValue(update.ObjectGUID)
}

func (i *Item) PropertySeed() uint32 {
	return i.GetUint32Value(update.ItemPropertySeed)
}

func (i *Item) RandomPropertiesID() uint32 {
	return i.GetUint32Value(update.ItemRandomPropertiesID)
}

func (i *Item) StackCount() uint32 {
	sc := i.GetUint32Value(update.ItemStackCount)
	if sc == 0 {
		return 1
	}

	return sc
}

func (i *Item) ContainerNumSlots() uint32 {
	return i.GetUint32Value(update.ContainerNumSlots)
}

func (i *Item) BagEmpty() bool {
	gArray, err := i.Get(update.ContainerSlots)
	if err != nil {
		return true
	}

	if gArray == nil {
		return true
	}

	for _, v := range gArray.([]*guid.GUID) {
		if v != nil {
			if *v != guid.Nil {
				return false
			}
		}
	}

	return true
}

func (i *Item) IsBag() bool {
	return i.GetUint32Value(update.ContainerNumSlots) > 0
}

func (i *Item) ID() uint64 {
	return i.GUID().Counter()
}

func (i *Item) TypeID() guid.TypeID {
	if i.IsBag() {
		return guid.TypeContainer
	}

	return guid.TypeItem
}

func (i *Item) Values() *update.ValuesBlock {
	return i.ValuesBlock
}

func (s *Session) GetItemTemplate(it wdb.Item) wdb.ItemTemplate {
	var itmp wdb.ItemTemplate
	found, err := s.DB().Where("id = ?", it.ItemID).Get(&itmp)
	if !found {
		panic(err)
	}

	return itmp
}

func (s *Session) GetItemTemplateByEntry(entry uint32) wdb.ItemTemplate {
	var itmp wdb.ItemTemplate
	found, err := s.DB().Where("entry = ?", entry).Get(&itmp)
	if !found {
		panic(err)
	}

	return itmp
}

func (s *Session) NewItem(it wdb.Item) *Item {
	template := s.GetItemTemplate(it)

	i := &Item{it.ItemID, update.NewValuesBlock()}
	g := guid.RealmSpecific(guid.Item, s.WS.RealmID(), it.ID)
	i.SetGUIDValue(update.ObjectGUID, g)
	i.SetUint32Value(update.ObjectEntry, template.Entry)
	flg, err := update.ParseItemFlag(template.Flags)
	if err != nil {
		panic(err)
	}

	code, err := flg.Resolve(s.Version())
	if err != nil {
		panic(err)
	}

	mask := guid.TypeMaskObject | guid.TypeMaskItem

	if template.ContainerSlots > 0 {
		mask |= guid.TypeMaskContainer
		i.SetUint32Value(update.ContainerNumSlots, uint32(template.ContainerSlots))
		i.Set(update.ContainerSlots, make([]*guid.GUID, int(template.ContainerSlots)))
	}

	i.SetTypeMask(s.Version(), mask)
	i.SetFloat32Value(update.ObjectScaleX, 1.0)

	i.SetGUIDValue(update.ItemOwner, s.GUID())
	if it.Creator != 0 {
		i.SetGUIDValue(update.ItemCreator, guid.RealmSpecific(guid.Player, s.WS.RealmID(), it.Creator))
	}
	i.SetUint32Value(update.ItemDurability, template.MaxDurability)
	i.SetUint32Value(update.ItemMaxDurability, template.MaxDurability)

	if template.Stackable != 0 {
		i.SetUint32Value(update.ItemStackCount, it.StackCount)
	}

	i.Set(update.ItemSpellCharges, make([]*uint32, 5))

	// todo: source charges from item struct
	for x := 0; x < len(template.Spells); x++ {
		i.SetUint32ArrayValue(update.ItemSpellCharges, x, uint32(template.Spells[x].Charges))
	}

	yo.Warn(template.Flags, flg, fmt.Sprintf("0x%08X\n", uint32(code)))

	i.SetUint32Value(update.ItemFlags, uint32(code))

	return i
}

func (s *Session) PlayerID() uint64 {
	return s.GUID().Counter()
}

func (s *Session) InitInventoryManager() {
	var inv []wdb.Inventory
	s.WS.DB.Where("player = ?", s.PlayerID()).Find(&inv)

	// changes := make(map[update.Global]interface{})
	// arrayData := &update.ArrayData{
	// 	Cols: []string{"Creator", "Entry", "Enchantments", "Properties"},
	// }

	s.ValuesBlock.Set(update.PlayerInventorySlots, make([]*guid.GUID, 39))

	displaySlots := map[uint8]uint64{}

	for _, v := range inv {
		if v.Bag == 255 && v.Slot <= uint8(packet.EquipLen(s.Version())) {
			displaySlots[v.Slot] = v.ItemID
		}
	}

	// Fill out structure
	s.ValuesBlock.Set(update.PlayerVisibleItems, update.InitArrayData(s.Version(), update.PlayerVisibleItems))

	for i, itemID := range displaySlots {
		var item wdb.Item
		found, err := s.DB().Where("id = ?", itemID).Get(&item)
		if !found {
			panic(err)
		}

		var itemTemplate wdb.ItemTemplate
		found, err = s.DB().Where("id = ?", item.ItemID).Get(&itemTemplate)
		if !found {
			panic(err)
		}

		// s.ValuesBlock.SetArrayValue(update.PlayerVisibleItems, int(i), "Creator", guid.RealmSpecific(guid.Player, s.WS.RealmID(), item.Creator))
		s.ValuesBlock.SetArrayValue(update.PlayerVisibleItems, int(i), "Entry", itemTemplate.Entry)
	}

	s.Inventory = make(map[guid.GUID]*Item)

	for _, v := range inv {
		if v.Bag == 255 {
			var item wdb.Item
			found, err := s.DB().Where("id = ?", v.ItemID).Get(&item)
			if !found {
				panic(err)
			}

			itemObject := s.NewItem(item)

			s.SetGUIDArrayValue(update.PlayerInventorySlots, int(v.Slot), itemObject.GUID())

			if itemObject.IsBag() {
				for _, bagContent := range inv {
					if bagContent.Bag == v.Slot {
						var bagContentItem wdb.Item
						found, err := s.DB().Where("id = ?", bagContent.ItemID).Get(&bagContentItem)
						if !found {
							panic(err)
						}

						bagContentObject := s.NewItem(bagContentItem)
						s.Inventory[bagContentObject.GUID()] = bagContentObject

						itemObject.SetGUIDArrayValue(update.ContainerSlots, int(bagContent.Slot), bagContentObject.GUID())

						s.SendObjectCreate(bagContentObject)
					}
				}
			}

			s.SendObjectCreate(itemObject)

			s.Inventory[itemObject.GUID()] = itemObject
		}
	}
}

func (s *Session) HandleItemQuery(e *etc.Buffer) {
	item := e.ReadUint32()
	fmt.Println("player queried item...", item)

	var it wdb.ItemTemplate
	found, err := s.DB().Where("entry = ?", item).Get(&it)
	if err != nil {
		panic(err)
	}

	fmt.Println("Queried", item, found)

	resp := packet.NewWorldPacket(packet.SMSG_ITEM_QUERY_SINGLE_RESPONSE)
	if !found {
		resp.WriteUint32(item | 0x80000000)
		s.SendAsync(resp)
		return
	}

	resp.WriteUint32(it.Entry)
	resp.WriteUint32(it.Class)
	resp.WriteUint32(it.Subclass)
	resp.WriteCString(it.Name)
	resp.WriteCString("")
	resp.WriteCString("")
	resp.WriteCString("")
	resp.WriteUint32(it.DisplayID)
	resp.WriteUint32(uint32(it.Quality))

	flg, err := update.ParseItemFlag(it.Flags)
	if err != nil {
		panic(err)
	}

	err = flg.Encode(resp, s.Version())
	if err != nil {
		panic(err)
	}

	resp.WriteUint32(it.BuyPrice)
	resp.WriteUint32(it.SellPrice)
	resp.WriteUint32(uint32(it.InventoryType))
	resp.WriteInt32(it.AllowableClass)
	resp.WriteInt32(it.AllowableRace)
	resp.WriteUint32(it.ItemLevel)
	resp.WriteUint32(uint32(it.RequiredLevel))
	resp.WriteUint32(it.RequiredSkill) // id from SkillLine.dbc
	resp.WriteUint32(it.RequiredSkillRank)
	resp.WriteUint32(it.RequiredSpell) // id from Spell.dbc
	resp.WriteUint32(it.RequiredHonorRank)
	resp.WriteUint32(it.RequiredCityRank)
	resp.WriteUint32(it.RequiredReputationFaction) // id from Faction.dbc
	resp.WriteUint32(it.RequiredReputationRank)
	resp.WriteUint32(it.MaxCount)
	resp.WriteUint32(it.Stackable)
	resp.WriteUint32(uint32(it.ContainerSlots))

	for x := 0; x < 10; x++ {
		if x >= len(it.Stats) {
			resp.WriteUint32(0)
			resp.WriteInt32(0)
		} else {
			resp.WriteUint32(uint32(it.Stats[x].Type))
			resp.WriteInt32(it.Stats[x].Value)
		}
	}

	for x := 0; x < 5; x++ {
		if x >= len(it.Damage) {
			resp.WriteFloat32(0)
			resp.WriteFloat32(0)
			resp.WriteUint32(0)
		} else {
			resp.WriteFloat32(it.Damage[x].Min)
			resp.WriteFloat32(it.Damage[x].Max)
			resp.WriteUint32(uint32(it.Damage[x].Type))
		}
	}

	resp.WriteUint32(it.Armor)
	resp.WriteUint32(it.HolyRes)
	resp.WriteUint32(it.FireRes)
	resp.WriteUint32(it.NatureRes)
	resp.WriteUint32(it.FrostRes)
	resp.WriteUint32(it.ShadowRes)
	resp.WriteUint32(it.ArcaneRes)
	resp.WriteUint32(it.Delay)
	resp.WriteUint32(it.AmmoType)
	resp.WriteFloat32(it.RangedModRange)

	for x := 0; x < 5; x++ {
		if x >= len(it.Spells) {
			resp.WriteUint32(0)
			resp.WriteUint32(0)
			resp.WriteUint32(0)
			resp.WriteInt32(-1)
			resp.WriteUint32(0)
			resp.WriteInt32(-1)
		} else {
			resp.WriteUint32(it.Spells[x].ID)
			resp.WriteUint32(it.Spells[x].Trigger)
			resp.WriteInt32(it.Spells[x].Charges)
			resp.WriteInt32(int32(it.Spells[x].Cooldown))
			resp.WriteUint32(it.Spells[x].Category)
			resp.WriteInt32(int32(it.Spells[x].CategoryCooldown))
		}
	}

	resp.WriteUint32(uint32(it.Bonding))
	resp.WriteCString(it.Description)
	resp.WriteUint32(it.PageText)
	resp.WriteUint32(it.LanguageID)
	resp.WriteUint32(it.PageMaterial)
	resp.WriteUint32(it.StartQuest)
	resp.WriteUint32(it.LockID)
	resp.WriteInt32(it.Material)
	resp.WriteUint32(it.Sheath)
	resp.WriteUint32(it.RandomProperty)
	resp.WriteUint32(it.Block)
	resp.WriteUint32(it.Itemset)
	resp.WriteUint32(it.MaxDurability)
	resp.WriteUint32(it.Area)
	resp.WriteInt32(it.Map)
	resp.WriteInt32(it.BagFamily)

	s.SendAsync(resp)
}

func (s *Session) HandleSwapInventoryItem(e *etc.Buffer) {
	src := e.ReadByte()
	dst := e.ReadByte()

	s.SwapItem(255, src, 255, dst)
}

func (s *Session) HandleSwapItem(e *etc.Buffer) {
	dstBag := e.ReadByte()
	dstSlot := e.ReadByte()
	srcBag := e.ReadByte()
	srcSlot := e.ReadByte()

	s.SwapItem(srcBag, srcSlot, dstBag, dstSlot)
}

func (s *Session) SendEquipError(ir packet.InventoryResult, src, dest *Item) {
	pkt := packet.NewWorldPacket(packet.SMSG_INVENTORY_CHANGE_FAILURE)
	inv, ok := packet.InventoryResultDescriptors[s.Version()][ir]
	if !ok {
		panic(fmt.Errorf("Cannot send this inventory result %d", ir))
	}

	pkt.WriteByte(inv)

	if ir != packet.EQUIP_ERR_OK {
		if ir == packet.EQUIP_ERR_CANT_EQUIP_LEVEL_I {
			itt := s.GetItemTemplateByEntry(src.Entry())
			pkt.WriteUint32(uint32(itt.RequiredLevel))
		}

		var srcGuid, dstGuid guid.GUID

		if src != nil {
			srcGuid = src.GUID()
		}

		if dest != nil {
			dstGuid = dest.GUID()
		}

		srcGuid.EncodeUnpacked(s.Version(), pkt)
		dstGuid.EncodeUnpacked(s.Version(), pkt)
		pkt.WriteByte(0)
	}

	s.SendAsync(pkt)
}

func (s *Session) GetItemByPos(bag, slot uint8) (*wdb.Inventory, *Item) {
	var target guid.GUID

	maxSlot := uint8(24)

	if bag == 255 {
		target = s.GetGUIDArrayValue(update.PlayerInventorySlots, int(slot))
	} else {
		if bag > maxSlot {
			return nil, nil
		}

		bagGUID := s.GetGUIDArrayValue(update.PlayerInventorySlots, int(bag))
		if bagGUID == guid.Nil {
			fmt.Println("bag", bag, "doesnt exist")
			return nil, nil
		}

		bagIt, ok := s.Inventory[bagGUID]
		if !ok {
			return nil, nil
		}

		target = bagIt.GetGUIDArrayValue(update.ContainerSlots, int(slot))
	}

	if target == guid.Nil {
		fmt.Println("target does not exist in", bag, slot)
		return nil, nil
	}

	it, ok := s.Inventory[target]
	if !ok {
		panic("Item referenced in inventory but does not exist in inventory manager: " + target.String())
	}

	return &wdb.Inventory{
		ItemID: it.ID(),
		Player: s.PlayerID(),
		Bag:    bag,
		Slot:   slot,
	}, it

	// var inv wdb.Inventory
	// found, _ := s.DB().Where("player = ?", s.PlayerID()).Where("bag = ?", bag).Where("slot = ?", slot).Get(&inv)
	// if !found {
	// 	return nil, nil
	// }

	// trg := guid.RealmSpecific(guid.Item, s.WS.RealmID(), inv.ItemID)

	// return &inv, it
}

// todo: cache
func (s *Session) GetItemRecord(id uint64) (wdb.Item, error) {
	var i wdb.Item
	fnd, err := s.DB().Where("id = ?", id).Get(&i)
	if err != nil {
		return i, err
	}

	if !fnd {
		return i, fmt.Errorf("could not find item for %d")
	}

	return i, nil
}

func (s *Session) IsEquipmentPos(bag, slot uint8) bool {
	if bag != 255 {
		return false
	}

	return slot < 19
}

func (s *Session) IsValidPos(bag, slot uint8) bool {
	// main backpack
	if bag == 255 {
		return slot < 39
	}

	if bag > 23 {
		return false
	}

	// check bag slots
	bagGUID := s.ValuesBlock.GetGUIDArrayValue(update.PlayerInventorySlots, int(bag))

	if bagGUID == guid.Nil {
		return false
	}

	bagItem := s.Inventory[bagGUID]
	if bagItem == nil {
		return false
	}

	if bagItem.IsBag() == false {
		return false
	}

	return true
}

func (s *Session) HasItem(entry string) bool {
	for _, i := range s.Inventory {
		if i.ItemID == entry {
			return true
		}
	}
	return false
}

func (s *Session) EquippableIn(it *Item, slot uint8) bool {
	itt := s.GetItemTemplateByEntry(it.Entry())

	if itt.InventoryType == dbc.IT_Weapon {
		if slot == (packet.Display_MainHand-1) || slot == (packet.Display_OffHand-1) {
			return true
		}
	}

	iType, ok := eMap[itt.InventoryType]
	if !ok {
		return false
	}

	return (iType - 1) == slot
}

// UNSAFE! will cause database/game corruption or exploits if called incorrectly, or with invalid parameters.
// Just transfers an item internally, irrespective of restrictions.
// If something is in dstInv (which it should not be), it will be lost forever.
func (s *Session) transferItemUnsafe(srcInv *wdb.Inventory, deleteSrc bool, dstBag, dstSlot uint8) {
	if srcInv.Bag == 255 {
		if deleteSrc {
			s.ValuesBlock.SetGUIDArrayValue(update.PlayerInventorySlots, int(srcInv.Slot), guid.Nil)

			if srcInv.Slot <= 19 {
				// remove armor and show change
				s.ValuesBlock.SetArrayValue(update.PlayerVisibleItems, int(srcInv.Slot), "Entry", uint32(0))
			}
		}
	} else {
		bgItem := s.GetBagItem(srcInv.Bag)
		bgItem.SetGUIDArrayValue(update.ContainerSlots, int(srcInv.Slot), guid.Nil)
	}

	srcInv.Bag = dstBag
	srcInv.Slot = dstSlot

	go s.DB().Where("item_id = ?", srcInv.ItemID).Cols("bag", "slot").Update(srcInv)

	if dstBag == 255 {
		s.ValuesBlock.SetGUIDArrayValue(update.PlayerInventorySlots, int(dstSlot), guid.RealmSpecific(guid.Item, s.WS.RealmID(), srcInv.ItemID))

		if dstSlot < 19 {
			// show armor change to other players
			var it wdb.Item
			found, err := s.DB().Where("id = ?", srcInv.ItemID).Get(&it)
			if !found {
				panic(err)
			}

			tpl := s.GetItemTemplate(it)

			s.ValuesBlock.SetArrayValue(update.PlayerVisibleItems, int(dstSlot), "Entry", tpl.Entry)
		}
	} else {
		bgItem := s.GetBagItem(dstBag)
		bgItem.SetGUIDArrayValue(update.ContainerSlots, int(dstSlot), guid.RealmSpecific(guid.Item, s.WS.RealmID(), srcInv.ItemID))
	}
}

func (s *Session) SwapItem(srcBag, srcSlot, dstBag, dstSlot uint8) {
	if !s.IsAlive() {
		s.SendEquipError(packet.EQUIP_ERR_PLAYER_DEAD, nil, nil)
		return
	}

	// Start position validation

	if !s.IsValidPos(srcBag, srcSlot) {
		yo.Warn("Invalid pos src ", srcBag, srcSlot)
		s.SendEquipError(packet.EQUIP_ERR_WRONG_SLOT, nil, nil)
		return
	}

	if !s.IsValidPos(dstBag, dstSlot) {
		yo.Warn("Invalid pos dst ", dstBag, dstSlot)
		s.SendEquipError(packet.EQUIP_ERR_WRONG_SLOT, nil, nil)
		return
	}

	srcInv, src := s.GetItemByPos(srcBag, srcSlot)
	dstInv, dst := s.GetItemByPos(dstBag, dstSlot)

	if src == nil {
		s.SendEquipError(packet.EQUIP_ERR_ITEM_NOT_FOUND, src, dst)
		return
	}

	// cannot put bag in itself.
	if src.IsBag() && dstBag == srcSlot {
		s.SendEquipError(packet.EQUIP_ERR_BAG_IN_BAG, src, dst)
		return
	}

	if dst != nil {
		if src.IsBag() && dst.IsBag() {
			// todo: seamless swap
			if !src.BagEmpty() || !dst.BagEmpty() {
				s.SendEquipError(packet.EQUIP_ERR_BAG_IN_BAG, src, dst)
				return
			}
		}
	} else {
		if src.IsBag() {
			if !src.BagEmpty() {
				s.SendEquipError(packet.EQUIP_ERR_BAG_IN_BAG, src, dst)
				return
			}
		}
	}

	if dstBag == 255 {
		if srcBag == 255 {
			// What the hell? You can't put your shirt in your pants slot, come on...
			if srcSlot < 19 && dstSlot < 19 {
				s.SendEquipError(packet.EQUIP_ERR_INTERNAL_BAG_ERROR, src, dst)
				return
			}
		}

		// is the target slot an equipment slot?
		if dstSlot < 19 {
			if !s.EquippableIn(src, dstSlot) {
				yo.Warn(src.Entry(), "not equippable in", dstSlot)
				s.SendEquipError(packet.EQUIP_ERR_WRONG_SLOT, src, dst)
				return
			}
		}
	}

	// Target is not empty. We have to transfer the target back to src to complete the slot.
	if dst != nil && srcBag == 255 && srcSlot < 19 {
		if !s.EquippableIn(dst, srcSlot) {
			s.SendEquipError(packet.EQUIP_ERR_WRONG_SLOT, src, dst)
			return
		}
	}

	// TODO: implement bag checks
	// TODO: swap filled bags

	// merge stacks
	if dst != nil && src != nil {
		// same type
		if src.Entry() == dst.Entry() {
			tpl := s.GetItemTemplateByEntry(src.Entry())
			if tpl.Stackable != 0 {
				availableSpace := tpl.Stackable - dst.StackCount()
				if availableSpace != 0 {
					srcHas := src.StackCount()
					if availableSpace < srcHas {
						// destination can't hold all of src's stack
						if err := s.modifyStackCount(dst.GUID(), tpl.Stackable); err != nil {
							panic(err)
						}

						if err := s.modifyStackCount(src.GUID(), srcHas-availableSpace); err != nil {
							panic(err)
						}
					} else {
						// destination can hold src's stack
						if err := s.modifyStackCount(dst.GUID(), dst.StackCount()+srcHas); err != nil {
							panic(err)
						}

						if _, err := s.removeItemByGUID(src.GUID()); err != nil {
							panic(err)
						}
					}

					return
				}
			}
		}
	}

	s.transferItemUnsafe(srcInv, true, dstBag, dstSlot)

	if dst != nil {
		s.transferItemUnsafe(dstInv, false, srcBag, srcSlot)
	}

	s.SendBagUpdate(srcBag)

	if dstBag != srcBag {
		s.SendBagUpdate(dstBag)
	}
}

func (s *Session) getEquippableInventorySlot(ty uint8) (uint8, packet.InventoryResult) {
	// todo: check for dual wield capability
	if ty == dbc.IT_Weapon {
		return packet.Display_MainHand - 1, 0
	}

	if ty == dbc.IT_Bag {
		for x := 19; x < 23; x++ {
			if s.GetGUIDArrayValue(update.PlayerInventorySlots, x) == guid.Nil {
				return uint8(x), packet.EQUIP_ERR_OK
			}
		}
	}

	u, ok := eMap[ty]
	if !ok {
		return 0, packet.EQUIP_ERR_NOT_EQUIPPABLE
	}

	return u - 1, packet.EQUIP_ERR_OK
}

func (s *Session) HandleAutoEquipItem(e *etc.Buffer) {
	srcBag := e.ReadByte()
	srcSlot := e.ReadByte()

	srcInv, src := s.GetItemByPos(srcBag, srcSlot)

	if srcInv == nil {
		return
	}

	template := s.GetItemTemplateByEntry(src.Entry())
	dstBag := uint8(255)
	dstSlot, err := s.getEquippableInventorySlot(template.InventoryType)
	if err != packet.EQUIP_ERR_OK {
		s.SendEquipError(err, src, nil)
		return
	}

	s.SwapItem(srcBag, srcSlot, dstBag, dstSlot)
}

// only call if you have lock
func (s *Session) removeItemByGUID(g guid.GUID) (uint32, error) {
	it, ok := s.Inventory[g]
	if !ok {
		return 0, fmt.Errorf("no such item: %s", g)
	}

	var inv wdb.Inventory
	found, err := s.DB().Where("item_id = ?", it.ID()).Get(&inv)
	if err != nil {
		return 0, err
	}

	if !found {
		return 0, fmt.Errorf("could not find inventory slot for item %s", g)
	}

	if inv.Bag == 255 {
		if inv.Slot < 19 {
			s.ValuesBlock.SetArrayValue(update.PlayerVisibleItems, int(inv.Slot), "Entry", uint32(0))
		}
		s.ValuesBlock.SetGUIDArrayValue(update.PlayerInventorySlots, int(inv.Slot), guid.Nil)

		s.Map().PropagateChanges(s.GUID())
	} else {
		panic("nyi")
	}

	stackCount := it.StackCount()

	delete(s.Inventory, g)
	s.SendObjectDelete(g)

	go s.DB().Where("item_id = ?", it.ID()).Delete(new(wdb.Inventory))
	go s.DB().Where("id = ?", it.ID()).Delete(new(wdb.Item))

	return stackCount, nil
}

func (s *Session) SendItemUpdate(it *Item) {
	s.SendUpdateData(update.ValuesPrivate, &update.Data{
		Blocks: []update.Block{
			{
				it.GUID(),
				it.ValuesBlock,
			},
		},
	})
}

func (s *Session) modifyStackCount(item guid.GUID, count uint32) error {
	it, ok := s.Inventory[item]
	if !ok {
		return fmt.Errorf("no item %s", item)
	}

	var itemData wdb.Item
	found, err := s.DB().Where("id = ?", it.ID()).Get(&itemData)
	if err != nil {
		return err
	}

	if !found {
		return fmt.Errorf("could not find item in database for %s", it.GUID())
	}

	itemData.StackCount = count
	go s.DB().Where("id = ?", it.ID()).Cols("stack_count").Update(&itemData)

	it.SetUint32Value(update.ItemStackCount, count)
	s.SendItemUpdate(it)

	return nil
}

func (s *Session) GetInventoryHeap() InventoryHeap {
	var inv []wdb.Inventory
	s.DB().Where("player = ?", s.PlayerID()).Find(&inv)

	var nheap []wdb.Inventory

	// de-select for equipped items.
	for _, v := range inv {
		if v.Bag == 255 {
			if v.Slot > 23 {
				nheap = append(nheap, v)
			}
		} else {
			nheap = append(nheap, v)
		}
	}

	ih := InventoryHeap(nheap)
	sort.Sort(ih)

	return ih
}

// May fail. run s.VerifyAvailableSpaceFor(itemID) before executing
func (s *Session) AddItem(itemID string, count int, received, created bool) error {
	if count == 0 {
		count = 1
	}

	// get entry from itemID
	var template wdb.ItemTemplate
	found, err := s.DB().Where("id = ?", itemID).Get(&template)
	if err != nil {
		return err
	}

	if !found {
		return fmt.Errorf("no such item: %s", itemID)
	}

	ih := s.GetInventoryHeap()

	if count < 0 {
		// negative count means subtract items
		countRemaining := uint32(-count)
		for _, inventory := range ih {
			if countRemaining == 0 {
				return nil
			}

			itemGUID := guid.RealmSpecific(guid.Item, s.WS.RealmID(), inventory.ItemID)
			item, ok := s.Inventory[itemGUID]
			if !ok {
				return fmt.Errorf("no inventory for %s", itemGUID)
			}

			if item.Entry() == template.Entry {
				if template.Stackable == 0 {
					i, err := s.removeItemByGUID(item.GUID())
					if err != nil {
						return err
					}

					countRemaining -= i
				} else {
					stackCount := item.StackCount()

					if stackCount <= countRemaining {
						// This slot has less then the remaining count, so we can just remove it entirely.
						removed, err := s.removeItemByGUID(item.GUID())
						if err != nil {
							return err
						}

						countRemaining -= removed
					} else {
						removed := stackCount - countRemaining
						// this slot has more than the remaining count of items to be removed, so let's remove the remaining count of items to be destroyed from the item object
						if err := s.modifyStackCount(item.GUID(), removed); err != nil {
							return err
						}

						countRemaining -= removed
					}
				}
			}
		}

		if countRemaining > 0 {
			return fmt.Errorf("could not remove %d items", countRemaining)
		}

		return nil
	}

	sentItem := false

	countRemaining := uint32(count)

	if template.Stackable != 0 {
		// See if we have other items of this kind, and if we can merge.
		for _, item := range s.Inventory {
			if countRemaining == 0 {
				return nil
			}

			if item.Entry() == template.Entry {
				stackCount := item.StackCount()
				if stackCount < template.Stackable {
					// we have a mergeable item slot!
					availableStackCount := template.Stackable - stackCount

					var inv wdb.Inventory
					fnd, err := s.DB().Where("item_id = ?", item.ID()).Get(&inv)
					if !fnd {
						panic(err)
					}

					// we can add the remaining items and stop.
					if countRemaining <= availableStackCount {
						if !sentItem {
							s.SendNewItem(item, received, created, true, inv.Bag, inv.Slot, uint32(count))
							sentItem = true
						}
						s.modifyStackCount(item.GUID(), stackCount+countRemaining)
						return nil
					}

					s.SendNewItem(item, received, created, true, inv.Bag, inv.Slot, template.Stackable)
					// we can stack, but it will overflow.
					s.modifyStackCount(item.GUID(), template.Stackable)

					countRemaining -= availableStackCount
				}
			}
		}
	}

	// Transfer to empty slots.
	if countRemaining > 0 {
		type bagReceiverPos struct {
			Bag  uint8
			Slot uint8
		}

		var freeSlots []bagReceiverPos

		for x := 23; x < 39; x++ {
			gp := s.ValuesBlock.GetGUIDArrayValue(update.PlayerInventorySlots, x)

			if gp == guid.Nil {
				freeSlots = append(freeSlots, bagReceiverPos{
					Bag:  255,
					Slot: uint8(x),
				})
			}
		}

		for x := 0; x < 4; x++ {
			if s.IsValidPos(uint8(x), 0) {
				bgItem := s.GetBagItem(uint8(x))
				for bagSlot := uint32(0); bagSlot < bgItem.ContainerNumSlots(); bagSlot++ {
					gp := bgItem.GetGUIDArrayValue(update.ContainerSlots, int(bagSlot))
					if gp == guid.Nil {
						freeSlots = append(freeSlots, bagReceiverPos{
							Bag:  uint8(x),
							Slot: uint8(bagSlot),
						})
					}
				}
			}
		}

		for _, pos := range freeSlots {
			if countRemaining == 0 {
				break
			}

			newItem := wdb.Item{
				ItemType:  uint32(template.InventoryType),
				ItemID:    itemID,
				DisplayID: 0,
			}

			if template.Stackable != 0 {
				if countRemaining >= template.Stackable {
					newItem.StackCount = template.Stackable
					if newItem.StackCount == 0 {
						newItem.StackCount = 1
					}
				} else {
					newItem.StackCount = countRemaining
				}
			} else {
				newItem.StackCount = 1
			}

			if _, err := s.DB().Insert(&newItem); err != nil {
				panic(err)
			}

			invObject := wdb.Inventory{
				ItemID: newItem.ID,
				Player: s.PlayerID(),
				Bag:    pos.Bag,
				Slot:   pos.Slot,
			}

			go s.DB().Insert(&invObject)

			it := s.NewItem(newItem)
			s.Inventory[it.GUID()] = it

			s.SetBagGUIDSlot(pos.Bag, pos.Slot, it.GUID())

			s.Map().PropagateChanges(s.GUID())

			if !sentItem {
				s.SendNewItem(it, received, created, true, pos.Bag, pos.Slot, uint32(count))
				sentItem = true
			}

			s.SendObjectCreate(it)

			countRemaining -= newItem.StackCount
		}
	}

	// TODO: place in additional bags

	if countRemaining > 0 {
		return fmt.Errorf("could not add %d items", countRemaining)
	}

	return nil
}

// func (s *Session) GetItemCount(entry uint32) uint32 {
// 	tpl := s.GetItemTemplateByEntry(entry)
// 	i64, err := s.DB().Where("player = ?", s.SumInt()
// }

func (s *Session) SendNewItem(item *Item, received, created, showInChat bool, bag, slot uint8, count uint32) {
	boolu32 := func(i bool) uint32 {
		if i {
			return 1
		}
		return 0
	}

	data := packet.NewWorldPacket(packet.SMSG_ITEM_PUSH_RESULT)
	s.GUID().EncodeUnpacked(s.Version(), data)
	data.WriteUint32(boolu32(received))
	data.WriteUint32(boolu32(created))
	data.WriteUint32(boolu32(showInChat))
	data.WriteByte(bag)
	if item.StackCount() == count {
		data.WriteInt32(int32(slot))
	} else {
		data.WriteInt32(-1)
	}
	data.WriteUint32(item.Entry())
	data.WriteUint32(item.PropertySeed())
	data.WriteUint32(item.RandomPropertiesID())
	data.WriteUint32(count)
	data.WriteUint32(0) // GetItemCount

	s.SendAsync(data)

	// TODO: share with group
}

func (s *Session) HandleDestroyItem(e *etc.Buffer) {
	if !s.IsAlive() {
		s.SendEquipError(packet.EQUIP_ERR_PLAYER_DEAD, nil, nil)
		return
	}

	bag := e.ReadByte()
	slot := e.ReadByte()
	count := e.ReadByte()

	_, item := s.GetItemByPos(bag, slot)

	if item.IsBag() && item.BagEmpty() == false {
		s.SendEquipError(packet.EQUIP_ERR_DESTROY_NONEMPTY_BAG, item, nil)
		return
	}

	if count != 0 {
		s.modifyStackCount(item.GUID(), item.GetUint32Value(update.ItemStackCount)-uint32(count))
	} else {
		_, err := s.removeItemByGUID(item.GUID())
		if err != nil {
			panic(err)
		}
	}
}

func (s *Session) SetBagGUIDSlot(bag, slot uint8, g guid.GUID) {
	if bag == 255 {
		s.ValuesBlock.SetGUIDArrayValue(update.PlayerInventorySlots, int(slot), g)
		return
	}

	if bag > 4 {
		panic("invalid bag")
	}

	bagItem := s.GetBagItem(bag)
	bagItem.SetGUIDArrayValue(update.ContainerSlots, int(slot), g)
}

func (s *Session) GetBagItem(bag uint8) *Item {
	if bag > 23 {
		panic("invalid bag")
	}

	bagGUID := s.ValuesBlock.GetGUIDArrayValue(update.PlayerInventorySlots, int(bag))

	if bagGUID == guid.Nil {
		panic("failed bag check, call IsValidPos before calling this function")
	}

	bagItem := s.Inventory[bagGUID]
	if bagItem == nil {
		panic(bagGUID.String() + " refers to non-existent item")
	}

	return bagItem
}

func (s *Session) SendBagUpdate(bag uint8) {
	if bag == 255 {
		s.ValuesBlock.Lock()
		s.Map().PropagateChanges(s.GUID())
		s.ValuesBlock.ClearChangesAndUnlock()
		return
	}

	bagItem := s.GetBagItem(bag)
	s.SendItemUpdate(bagItem)
}

func (s *Session) HandleSplitItem(e *etc.Buffer) {
	srcBag := e.ReadByte()
	srcSlot := e.ReadByte()
	dstBag := e.ReadByte()
	dstSlot := e.ReadByte()
	count := uint32(e.ReadByte())

	if !s.IsValidPos(srcBag, srcSlot) {
		s.Warnf("Invalid src pos: %d %d", srcBag, srcSlot)
		return
	}

	if !s.IsValidPos(dstBag, dstSlot) {
		s.Warnf("Invalid dst pos: %d %d", dstBag, dstSlot)
		return
	}

	_, src := s.GetItemByPos(srcBag, srcSlot)
	if src == nil {
		s.Warnf("Could not find source for that item.")
		return
	}

	// cannot split to equipment
	if s.IsEquipmentPos(dstBag, dstSlot) {
		s.Warnf("Destination is equipment position.")
		return
	}

	tpl := s.GetItemTemplateByEntry(src.Entry())

	if tpl.Stackable == 0 {
		s.Warnf("Template is unstackable.")
		return
	}

	if count >= src.StackCount() {
		s.Warnf("Attempted to split %d, more than you have in source: %d", count, src.StackCount())
		return
	}

	if count >= tpl.Stackable {
		s.Warnf("Attempted to split %d, more than is stackable: %d", count, tpl.Stackable)
		return
	}

	_, dst := s.GetItemByPos(dstBag, dstSlot)
	if dst != nil {
		if dst.Entry() != src.Entry() {
			s.SendEquipError(packet.EQUIP_ERR_CANT_STACK, src, dst)
			return
		}

		availableSpace := tpl.Stackable - dst.StackCount()
		if count > availableSpace {
			s.SendEquipError(packet.EQUIP_ERR_CANT_STACK, src, dst)
			return
		}

		if err := s.modifyStackCount(dst.GUID(), dst.StackCount()+count); err != nil {
			panic(err)
		}
	} else {
		s.modifyStackCount(src.GUID(), src.StackCount()-count)

		// create new item
		var newItem wdb.Item
		found, err := s.DB().Where("id = ?", src.ID()).Get(&newItem)
		if !found {
			panic(err)
		}

		newItem.ID = 0
		newItem.StackCount = count

		if _, err := s.DB().Insert(&newItem); err != nil {
			panic(err)
		}

		newInv := wdb.Inventory{
			ItemID: newItem.ID,
			Player: s.PlayerID(),
			Bag:    dstBag,
			Slot:   dstSlot,
		}

		newItemObject := s.NewItem(newItem)

		s.DB().Insert(&newInv)

		s.Inventory[newItemObject.GUID()] = newItemObject

		s.SendObjectCreate(newItemObject)

		s.SetBagGUIDSlot(dstBag, dstSlot, newItemObject.GUID())
		s.SendBagUpdate(dstBag)
	}
}
