package realm

import (
	"fmt"

	"github.com/superp00t/etc"
	"github.com/superp00t/gophercraft/guid"
	"github.com/superp00t/gophercraft/packet"
	"github.com/superp00t/gophercraft/realm/wdb"
)

func (s *Session) SendGossip(gossip *packet.Gossip) {
	s.SendAsync(gossip.Packet(s.Build()))
}

func (s *Session) GetValidGossipObject(id guid.GUID) (WorldObject, string) {
	fmt.Println("Asked to speak with ", id)

	object := s.Map().GetObject(id)
	if object == nil {
		return nil, ""
	}

	var menuID string

	switch object.TypeID() {
	case guid.TypeUnit:
		var creatureTemplate *wdb.CreatureTemplate
		s.DB().GetData(object.Values().GetUint32("Entry"), &creatureTemplate)
		if creatureTemplate != nil {
			menuID = creatureTemplate.GossipMenuId
		}
		if creatureTemplate.Gossip == false && creatureTemplate.Innkeeper == false {
			return nil, ""
		}
	default:
		fmt.Println("Client tried to speak with ", object.TypeID())
		return nil, ""
	}

	// Todo: invalidate if too far away.
	return object, menuID
}

func (s *Session) HandleGossipHello(gguid *etc.Buffer) {
	id := s.decodeUnpackedGUID(gguid)

	object, menuID := s.GetValidGossipObject(id)
	if object == nil {
		return
	}

	fmt.Println("found object", object)

	// GameObjects will be supported in the future.

	if menuID == "" {
		fmt.Println("No menu found")
		return
	}

	menu := &packet.Gossip{
		Speaker:   id,
		TextEntry: 0,
	}

	// Quests should be offered no matter what.
	s.WS.ThinkOn(GossipEvent, s, menuID, 0, menu)

	s.SendGossip(menu)
}

func (s *Session) HandleGossipSelectOption(e *etc.Buffer) {
	id := s.decodeUnpackedGUID(e)
	option := e.ReadUint32()

	object, menuID := s.GetValidGossipObject(id)
	if object == nil {
		return
	}

	menu := &packet.Gossip{
		Speaker:   id,
		TextEntry: 0,
	}

	// Quests should be offered no matter what.
	s.WS.ThinkOn(GossipEvent, s, menuID, option, menu)
	s.SendGossip(menu)
}

func (s *Session) HandleGossipTextQuery(e *etc.Buffer) {
	entry := e.ReadUint32()

	var nt *wdb.NPCText
	s.DB().GetData(entry, &nt)

	resp := packet.NewWorldPacket(packet.SMSG_NPC_TEXT_UPDATE)
	resp.WriteUint32(entry)

	if nt == nil {
		for x := 0; x < 8; x++ {
			resp.WriteFloat32(0)
			resp.WriteCString("Hail, $r.")
			resp.WriteCString("Hail, $r.")
			resp.WriteUint32(0)
			resp.WriteUint32(0)
			resp.WriteUint32(0)
			resp.WriteUint32(0)
			resp.WriteUint32(0)
			resp.WriteUint32(0)
			resp.WriteUint32(0)
		}
	} else {
		x := 0

		for ; x < len(nt.Opts); x++ {
			opt := nt.Opts[x]
			resp.WriteFloat32(opt.Prob)
			text := opt.Text.GetLocalized(s.Locale)
			resp.WriteCString(text)
			resp.WriteCString(text)
			resp.WriteUint32(opt.Lang)

			em := 0

			for ; em < len(opt.Emote); em += 2 {
				e := opt.Emote[em]
				resp.WriteUint32(e.Delay)
				resp.WriteUint32(e.ID)
			}

			for ; em < 6; em += 2 {
				resp.WriteUint32(0)
				resp.WriteUint32(0)
			}

		}

		for ; x < 8; x++ {
			resp.WriteFloat32(0)
			resp.WriteCString("")
			resp.WriteCString("")
			resp.WriteUint32(0)
			resp.WriteUint32(0)
			resp.WriteUint32(0)
			resp.WriteUint32(0)
			resp.WriteUint32(0)
			resp.WriteUint32(0)
			resp.WriteUint32(0)
		}
	}

	s.SendAsync(resp)
}
