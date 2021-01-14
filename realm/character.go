package realm

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/superp00t/gophercraft/packet/update"

	"github.com/superp00t/etc"
	"github.com/superp00t/etc/yo"
	"github.com/superp00t/gophercraft/format/dbc"
	"github.com/superp00t/gophercraft/realm/wdb"
	"github.com/superp00t/gophercraft/vsn"

	"github.com/superp00t/gophercraft/guid"

	"github.com/superp00t/gophercraft/packet"
)

var eMap = map[uint8]uint8{
	dbc.IT_Shield:   packet.Display_OffHand,
	dbc.IT_Robe:     packet.Display_Chest,
	dbc.IT_Head:     packet.Display_Head,
	dbc.IT_Neck:     packet.Display_Neck,
	dbc.IT_Shoulder: packet.Display_Shoulder,
	dbc.IT_Shirt:    packet.Display_Shirt,
	dbc.IT_Chest:    packet.Display_Chest,
	dbc.IT_Waist:    packet.Display_Waist,
	dbc.IT_Legs:     packet.Display_Legs,
	dbc.IT_Feet:     packet.Display_Feet,
	dbc.IT_Wrists:   packet.Display_Wrist,
	dbc.IT_Hands:    packet.Display_Hands,
	dbc.IT_Finger:   packet.Display_Finger1,
	dbc.IT_Trinket:  packet.Display_Trinket1,
	dbc.IT_Back:     packet.Display_Back,
	dbc.IT_TwoHand:  packet.Display_MainHand,
	dbc.IT_MainHand: packet.Display_MainHand,
	dbc.IT_Holdable: packet.Display_OffHand,
	dbc.IT_OffHand:  packet.Display_OffHand,
	dbc.IT_Ranged:   packet.Display_Ranged,
	dbc.IT_Gun:      packet.Display_Ranged,
	dbc.IT_Tabard:   packet.Display_Tabard,
	dbc.IT_Quiver:   packet.Display_Bag1,
	dbc.IT_Bag:      packet.Display_Bag1,
	dbc.IT_Thrown:   packet.Display_Ranged,
	dbc.IT_Ammo:     packet.Display_Ammo,
}

// func displayable(slotType Scrubuint32) bool {
// 	switch slotType {
// 		case packet.Display_Back, packet.Display_
// 	}
// }

// ScrubCharacter deletes a character PERMANENTLY from a server.
func (s *Server) ScrubCharacter(chr guid.GUID) {
	s.DB.Where("id = ?", chr.Counter()).Delete(new(wdb.Character))
	s.DB.Where("owner = ?", chr.Counter()).Delete(new(wdb.Item))
}

func (s *Session) getEquipment(chr uint64) []packet.Item {
	itemList := make([]packet.Item, packet.EquipLen(s.Build()))
	var inventory []wdb.Inventory
	err := s.DB().Where("player = ?", chr).Where("bag = 255").Where("slot < 19").Find(&inventory)
	if err != nil {
		yo.Fatal(err)
	}

	idex := 0

	for _, invRef := range inventory {
		var item wdb.Item
		found, err := s.DB().Where("id = ?", invRef.ItemID).Get(&item)
		if !found {
			panic(err)
		}

		pi := packet.Item{
			Model: item.DisplayID,
			Type:  uint8(item.ItemType),
		}

		// No transmog
		if item.DisplayID == 0 {
			var itt *wdb.ItemTemplate
			s.DB().GetData(item.ItemID, &itt)
			if itt != nil {
				pi.Model = itt.DisplayID
			}
		}

		if len(item.Enchantments) > 0 {
			pi.Enchantment = item.Enchantments[0]
		}

		if int(invRef.Slot) >= len(itemList) {
			idex++
			continue
		}
		itemList[int(invRef.Slot)] = pi

		idex++
	}

	return itemList
}

func (s *Session) HandleRequestCharacterList() {
	yo.Println("Character list requested")

	var chars []wdb.Character
	var pktChars []packet.Character

	err := s.WS.DB.Where("game_account = ?", s.GameAccount).Find(&chars)
	if err != nil {
		panic(err)
	}

	for _, v := range chars {
		characterGUID := guid.RealmSpecific(guid.Player, s.WS.RealmID(), v.ID)

		var flags packet.CharacterFlags

		sess, _ := s.WS.GetSessionByGUID(characterGUID)
		if sess != nil {
			flags |= packet.CharacterLockedForTransfer
		}

		if v.HideHelm {
			flags |= packet.CharacterHideHelm
		}

		if v.HideCloak {
			flags |= packet.CharacterHideCloak
		}

		if v.Ghost {
			flags |= packet.CharacterGhost
		}

		level := v.Level
		if level > 255 {
			level = 0
		}

		pktChars = append(pktChars, packet.Character{
			GUID:       characterGUID,
			Name:       v.Name,
			Race:       packet.Race(v.Race),
			Class:      packet.Class(v.Class),
			Gender:     v.Gender,
			Skin:       v.Skin,
			Face:       v.Face,
			HairStyle:  v.HairStyle,
			HairColor:  v.HairColor,
			FacialHair: v.FacialHair,
			FirstLogin: v.FirstLogin,
			Level:      uint8(v.Level),
			Flags:      flags,
			Zone:       v.Zone, // Goldshire. Once the login test is complete,
			Map:        v.Map,  // and players can move around within Goldshire without error
			X:          v.X,    // we can replace this with database data.
			Y:          v.Y,
			Z:          v.Z,
			Equipment:  s.getEquipment(v.ID),
		})
	}

	list := &packet.CharacterList{
		Characters: pktChars,
	}

	s.SendAsync(list.Packet(s.Build()))
}

func (s *Session) SendCharacterOp(opcode packet.CharacterOp) {
	pkt := packet.NewWorldPacket(packet.SMSG_CHAR_CREATE)
	if err := packet.EncodeCharacterOp(s.Build(), pkt.Buffer, opcode); err != nil {
		panic(err)
	}
	s.SendAsync(pkt)
}

func (s *Session) DeleteCharacter(b []byte) {
	gui := guid.Classic(etc.FromBytes(b).ReadUint64())
	s.WS.ScrubCharacter(gui)
	pkt := packet.NewWorldPacket(packet.SMSG_CHAR_DELETE)
	op := packet.CHAR_DELETE_SUCCESS
	if err := packet.EncodeCharacterOp(s.Build(), pkt.Buffer, op); err != nil {
		panic(err)
	}
	s.SendAsync(pkt)
}

func (s *Session) CreateCharacter(e *etc.Buffer) {
	name := e.ReadCString()

	if name == "" {
		s.SendCharacterOp(packet.CHAR_NAME_NO_NAME)
		return
	}

	// Check if character name is in use
	var chars []wdb.Character
	s.WS.DB.Where("name = ?", name).Find(&chars)
	if len(chars) != 0 {
		s.SendCharacterOp(packet.CHAR_CREATE_NAME_IN_USE)
		return
	}

	yo.Println("Registered name: ", name)
	pch := wdb.Character{}
	pch.ID = 0 // will be overwritten by insert
	pch.GameAccount = s.GameAccount
	pch.RealmID = s.WS.RealmID()
	pch.Name = name
	pch.Race = e.ReadByte()
	pch.Class = e.ReadByte()
	pch.Gender = e.ReadByte()
	pch.Skin = e.ReadByte()
	pch.Face = e.ReadByte()
	pch.HairStyle = e.ReadByte()
	pch.HairColor = e.ReadByte()
	pch.FacialHair = e.ReadByte()
	pch.Zone = 12
	pch.Map = 0
	pch.X = -9448.55 // TODO: we can replace this with database data, or with a location specified by config.
	pch.Y = 68.236
	pch.Z = 56.3225
	pch.O = 2.1115

	var race *dbc.Ent_ChrRaces
	// validate race and class.
	s.DB().GetData(uint32(pch.Race), &race)

	var class *dbc.Ent_ChrClasses
	s.DB().GetData(uint32(pch.Class), &class)

	if race == nil {
		fmt.Println("race not found", pch.Race)
		s.SendCharacterOp(packet.CHAR_CREATE_RESTRICTED_RACECLASS)
		return
	}

	if class == nil {
		s.SendCharacterOp(packet.CHAR_CREATE_RESTRICTED_RACECLASS)
		return
	}

	defaultLevel := s.WS.Config.Uint32("Char.StartLevel")
	pch.Level = defaultLevel
	if defaultLevel == 1 {
		// TODO: check for leveling requirement if configured
		if pch.Class == uint8(packet.CLASS_DEATH_KNIGHT) {
			pch.Level = 55
		}
	}

	startPos := s.Config().GetString("Char.StartPosition")

	if startPos == "" {
		pch.FirstLogin = true

		var plci *wdb.PlayerCreateInfo

		for _, pci := range s.WS.PlayerCreateInfo {
			if pci.Class == pch.Class && pci.Race == pch.Race {
				plci = &pci
				break
			}
		}

		if plci != nil {
			pch.Zone = plci.Zone
			pch.Map = plci.Map
			pch.X = plci.X
			pch.Y = plci.Y
			pch.Z = plci.Z
			pch.O = plci.O
		}
	} else {
		strs := strings.SplitN(startPos, " ", 2)
		if len(strs) != 2 {
			panic("invalid pos string")
		}

		mapID, err := strconv.ParseUint(strs[0], 0, 32)
		if err != nil {
			panic(err)
		}

		pos, err := update.ParsePosition(strs[1])
		if err != nil {
			panic(err)
		}

		pch.Zone = 0
		pch.Map = uint32(mapID)
		pch.X = pos.X
		pch.Y = pos.Y
		pch.Z = pos.Z
		pch.O = pos.O
	}

	_, err := s.WS.DB.Insert(&pch)
	if err != nil {
		yo.Fatal(err)
	}

	var eq []wdb.Item
	var st []*dbc.Ent_CharStartOutfit

	s.DB().GetStore(st).Range(func(k, v interface{}) bool {
		cso := v.(*dbc.Ent_CharStartOutfit)
		if uint8(cso.Class) == pch.Class && uint8(cso.Race) == pch.Race {
			st = append(st, cso)
		}

		return true
	})

	if len(st) != 0 {
		for i, v := range st[0].ItemIDs {
			if v != dbc.Empty {
				sid := st[0].InventoryTypes[i]
				// TODO, add to inventory if unequippable
				if sid == dbc.Empty {
					sid = 0
				}

				did := st[0].DisplayInfoIDs[i]
				if did == dbc.Empty {
					did = 0
				}

				if sid != 0 {
					itm := wdb.Item{
						ItemType:     uint32(sid),
						DisplayID:    uint32(did),
						ItemID:       fmt.Sprintf("it:%d", v),
						Enchantments: nil,
					}

					s.DB().Insert(&itm)

					eq = append(eq, itm)
				}
			}
		}
	} else {
		yo.Warn("No starting equipment files. Please install or generate a datapack.")
	}

	var inventory []wdb.Inventory
	var activatedSlots = map[uint8]bool{}

	for _, v := range eq {
		if v.ItemType == 0 {
			panic(v.ItemID)
		}

		var slot uint8

		if v.ItemType == dbc.IT_Weapon {
			slot = packet.Display_MainHand - 1

			if activatedSlots[slot] == true {
				slot = packet.Display_OffHand - 1
			}
		} else {
			slot = uint8(eMap[uint8(v.ItemType)]) - 1
			if slot == 255 {
				panic(fmt.Errorf("unknown item type %d", v.ItemType))
			}
		}

		activatedSlots[slot] = true

		inventory = append(inventory, wdb.Inventory{
			ItemID: v.ID,
			Player: pch.ID,
			Bag:    255, // Backpack
			Slot:   slot,
		})
	}

	_, err = s.DB().Insert(&inventory)
	if err != nil {
		panic(err)
	}

	var learnedAbilities []wdb.LearnedAbility

	for _, v := range s.WS.PlayerCreateAbilities {
		if (v.Race == -1 || v.Race == int8(pch.Race)) && (v.Class == -1 || v.Class == int8(pch.Class)) {
			learnedAbilities = append(learnedAbilities, wdb.LearnedAbility{
				Player: pch.ID,
				Spell:  v.Spell,
			})
		}
	}

	var actionButtons []wdb.ActionButton

	for _, v := range s.WS.PlayerCreateActionButtons {
		if (v.Race == -1 || v.Race == int8(pch.Race)) && (v.Class == -1 || v.Class == int8(pch.Class)) {
			actionButtons = append(actionButtons, wdb.ActionButton{
				Player: pch.ID,
				Button: v.Button,
				Action: v.Action,
				Type:   v.Type,
				Misc:   v.Misc,
			})
		}
	}

	s.DB().Insert(&actionButtons)
	s.DB().Insert(&learnedAbilities)

	s.SendCharacterOp(packet.CHAR_CREATE_SUCCESS)
}

func (s *Session) Build() vsn.Build {
	return vsn.Build(s.WS.Config.Version)
}

func (s *Server) GetNative(race packet.Race, gender uint8) uint32 {
	var races *dbc.Ent_ChrRaces
	s.DB.GetData(uint32(race), &races)
	// Just a gopher
	if races == nil {
		return 2838
	}

	// gender is a fuck
	if gender == 1 {
		return races.FemaleDisplayID
	}

	return races.MaleDisplayID
}
