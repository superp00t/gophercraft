package worldserver

import (
	"fmt"

	"github.com/superp00t/etc"
	"github.com/superp00t/etc/yo"
	"github.com/superp00t/gophercraft/format/dbc"
	"github.com/superp00t/gophercraft/worldserver/wdb"

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
}

// func displayable(slotType Scrubuint32) bool {
// 	switch slotType {
// 		case packet.Display_Back, packet.Display_
// 	}
// }

// ScrubCharacter deletes a character PERMANENTLY from a server.
func (s *WorldServer) ScrubCharacter(chr guid.GUID) {
	s.DB.Where("id = ?", chr.Counter()).Delete(new(wdb.Character))
	s.DB.Where("owner = ?", chr.Counter()).Delete(new(wdb.Item))
}

func (s *Session) getEquipment(chr uint64) []packet.Item {
	pi := make([]packet.Item, packet.EquipLen(s.Version()))
	var fnd []wdb.Item
	err := s.WS.DB.Where("owner = ?", chr).Where("equipped = 1").Find(&fnd)
	if err != nil {
		yo.Fatal(err)
	}

	idex := 0

	yo.Spew(fnd)

	for _, v := range fnd {
		equ, ok := eMap[uint8(v.ItemType)]
		if !ok {
			fmt.Println("no display for", v.ItemType)
			continue
		}

		if equ == 0 {
			continue
		}

		pi[equ-1] = packet.Item{
			Model:       v.DisplayID,
			Type:        uint8(v.ItemType),
			Enchantment: v.Enchantment,
		}

		idex++
	}

	return pi
}

func (s *Session) CharacterList(b []byte) {
	yo.Println("Character list requested")

	var chars []wdb.Character
	var pktChars []packet.Character

	err := s.WS.DB.Where("game_account = ?", s.GameAccount).Where("realm_id = ?", s.WS.RealmID).Find(&chars)
	if err != nil {
		panic(err)
	}

	for _, v := range chars {
		pktChars = append(pktChars, packet.Character{
			GUID:       guid.RealmSpecific(guid.Player, s.WS.RealmID, v.ID),
			Name:       v.Name,
			Race:       packet.Race(v.Race),
			Class:      packet.Class(v.Class),
			Gender:     v.Gender,
			Skin:       v.Skin,
			Face:       v.Face,
			HairStyle:  v.HairStyle,
			HairColor:  v.HairColor,
			FacialHair: v.FacialHair,
			Level:      v.Level,
			Zone:       12,    // Goldshire. Once the login test is complete,
			Map:        v.Map, // and players can move around within Goldshire without error
			X:          v.X,   // we can replace this with database data.
			Y:          v.Y,
			Z:          v.Z,
			Equipment:  s.getEquipment(v.ID),
		})
	}

	list := &packet.CharacterList{
		Characters: pktChars,
	}

	fmt.Println("Sending a response")
	s.SendAsync(list.Packet(s.Version()))
}

func (s *Session) SendCharacterOp(opcode packet.CharacterOp) {
	pkt := packet.NewWorldPacket(packet.SMSG_CHAR_CREATE)
	if err := packet.EncodeCharacterOp(s.Version(), pkt.Buffer, opcode); err != nil {
		panic(err)
	}
	s.SendAsync(pkt)
}

func (s *Session) DeleteCharacter(b []byte) {
	gui := guid.Classic(etc.FromBytes(b).ReadUint64())
	s.WS.ScrubCharacter(gui)
	pkt := packet.NewWorldPacket(packet.SMSG_CHAR_DELETE)
	op := packet.CHAR_DELETE_SUCCESS
	if err := packet.EncodeCharacterOp(s.Version(), pkt.Buffer, op); err != nil {
		panic(err)
	}
	s.SendAsync(pkt)
}

func (s *Session) CreateCharacter(b []byte) {
	e := etc.FromBytes(b)
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
	pch.RealmID = s.WS.RealmID
	pch.Name = name
	pch.Race = e.ReadByte()
	pch.Class = e.ReadByte()
	pch.Gender = e.ReadByte()
	pch.Skin = e.ReadByte()
	pch.Face = e.ReadByte()
	pch.HairStyle = e.ReadByte()
	pch.HairColor = e.ReadByte()
	pch.FacialHair = e.ReadByte()
	pch.Map = 0
	pch.X = -9448.55 // we can replace this with database data.
	pch.Y = 68.236
	pch.Z = 56.3225
	pch.O = 2.1115

	var race dbc.Ent_ChrRaces
	// validate race and class.
	found, _ := s.DB().Where("id = ?", pch.Race).Get(&race)

	var class dbc.Ent_ChrClasses
	found2, _ := s.DB().Where("id = ?", pch.Class).Get(&class)

	if !found || !found2 {
		s.SendCharacterOp(packet.CHAR_CREATE_EXPANSION_CLASS)
		return
	}

	defaultLevel := s.WS.Config.Byte("xp.startLevel")
	pch.Level = defaultLevel
	if defaultLevel == 1 {
		// TODO: check for leveling requirement if configured
		if pch.Class == uint8(packet.CLASS_DEATH_KNIGHT) {
			pch.Level = 55
		}
	}

	_, err := s.WS.DB.Insert(&pch)
	if err != nil {
		yo.Fatal(err)
	}

	var eq []wdb.Item
	var st []dbc.Ent_CharStartOutfit
	err = s.WS.DB.Where("race = ?", pch.Race).Where("class = ?", pch.Class).Find(&st)
	if err != nil {
		yo.Fatal(err)
	}

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

				eq = append(eq, wdb.Item{
					Owner:       pch.ID,
					ItemType:    sid,
					DisplayID:   did,
					ItemID:      v,
					Enchantment: 0,
					Equipped:    true,
				})
			}
		}
	} else {
		yo.Warn("No starting equipment files. Please install or generate a datapack.")
	}

	_, err = s.WS.DB.Insert(eq)
	if err != nil {
		yo.Fatal(err)
	}

	s.SendCharacterOp(packet.CHAR_CREATE_SUCCESS)
}

func (s *Session) Version() uint32 {
	return s.WS.Config.Version
}
