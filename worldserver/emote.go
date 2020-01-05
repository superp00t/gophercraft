package worldserver

import (
	"github.com/superp00t/etc"
	"github.com/superp00t/etc/yo"
	"github.com/superp00t/gophercraft/format/dbc"
	"github.com/superp00t/gophercraft/guid"
	"github.com/superp00t/gophercraft/packet"
	"github.com/superp00t/gophercraft/packet/update"
)

func (s *Session) isAlive() bool {
	return true
}

func (s *Session) canSpeak() bool {
	return true
}

func (s *Session) HandleStandStateChange(e *etc.Buffer) {
	anim := e.ReadUint32()

	// Validation
	switch anim {
	case packet.UNIT_STAND_STATE_STAND, packet.UNIT_STAND_STATE_KNEEL, packet.UNIT_STAND_STATE_SIT, packet.UNIT_STAND_STATE_SLEEP:
		break
	default:
		return
	}

	// Broadcast new stand state to server
	s.Map().ModifyObject(s.GUID(), map[update.Global]interface{}{
		update.UnitStandState: uint8(anim),
	})
}

func (s *Session) HandleTextEmote(e *etc.Buffer) {
	if !s.isAlive() {
		return
	}

	if !s.canSpeak() {
		return
	}

	textEmote := e.ReadUint32()
	emoteID := e.ReadUint32()
	target := s.decodeUnpackedGUID(e)

	yo.Warn(textEmote, emoteID)

	var emotes []dbc.Ent_EmotesText
	if err := s.DB().Where("id = ?", textEmote).Find(&emotes); err != nil {
		yo.Fatal(err)
	}

	if len(emotes) == 0 {
		s.Warnf("You appear to have sent an invalid emote command. Check to see if you have a base datapack installed.")
		return
	}

	em := emotes[0]

	switch em.EmoteID {
	case 12, 13, 16, 0: //sleep, sit, kneel, none
	default:
		s.HandleEmoteCommand(em.EmoteID)
	}

	var data string
	var err error
	if target != guid.Nil {
		data, err = s.WS.GetUnitNameByGUID(target)
		if err != nil {
			s.Warnf("%s guid=%s", err.Error(), target.Summary())
			return
		}
	}

	// // toSelfPacket := packet.NewWorldPacket(packet.SMSG_TEXT_EMOTE)
	emoPacket := packet.NewWorldPacket(packet.SMSG_TEXT_EMOTE)
	s.GUID().EncodeUnpacked(s.Version(), emoPacket)
	emoPacket.WriteUint32(textEmote)
	emoPacket.WriteUint32(emoteID)
	emoPacket.WriteUint32(uint32(len(data)) + 1)
	emoPacket.Write([]byte(data))
	emoPacket.WriteByte(0)

	s.SendAreaAll(emoPacket)
}

func (s *Session) HandleEmoteCommand(emoteID uint32) {
	p := packet.NewWorldPacket(packet.SMSG_EMOTE)
	p.WriteUint32(emoteID)
	s.GUID().EncodeUnpacked(s.Version(), p)
	s.SendAreaAll(p)
}

func (s *Session) GetTarget() guid.GUID {
	return s.GetGUIDValue(update.UnitTarget)
}

func (s *Session) HandleTarget(e *etc.Buffer) {
	tgt := s.decodeUnpackedGUID(e)

	s.Warnf("Targeting %s", tgt)

	s.Map().ModifyObject(s.GUID(), map[update.Global]interface{}{
		update.UnitTarget: tgt,
	})
}
