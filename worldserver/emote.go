package worldserver

import (
	"fmt"
	"math"

	"github.com/superp00t/etc"
	"github.com/superp00t/etc/yo"
	"github.com/superp00t/gophercraft/format/dbc"
	"github.com/superp00t/gophercraft/guid"
	"github.com/superp00t/gophercraft/packet"
	"github.com/superp00t/gophercraft/packet/update"
	"github.com/superp00t/gophercraft/worldserver/wdb"
)

func (s *Session) IsAlive() bool {
	return true
}

func (s *Session) CanSpeak() bool {
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
	s.SetStandState(uint8(anim))
}

func (s *Session) SetStandState(value uint8) {
	s.SetByte("StandState", value)
	s.Map().PropagateChanges(s.GUID())
}

func (s *Session) HandleTextEmote(e *etc.Buffer) {
	if !s.IsAlive() {
		return
	}

	if !s.CanSpeak() {
		return
	}

	textEmote := e.ReadUint32()
	emoteID := e.ReadUint32()
	target := s.decodeUnpackedGUID(e)

	yo.Warn(textEmote, emoteID)

	var emote *dbc.Ent_EmotesText
	wdb.GetData(textEmote, &emote)

	if emote == nil {
		s.Warnf("You appear to have sent an invalid emote command. Check to see if you have a base datapack installed.")
		return
	}

	switch emote.EmoteID {
	case 12, 13, 16, 0: //sleep, sit, kneel, none
	default:
		s.HandleEmoteCommand(emote.EmoteID)
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
	s.GUID().EncodeUnpacked(s.Build(), emoPacket)
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
	s.GUID().EncodeUnpacked(s.Build(), p)
	s.SendAreaAll(p)
}

func (s *Session) GetTarget() guid.GUID {
	return s.GetGUID("Target")
}

func (s *Session) HandleTarget(e *etc.Buffer) {
	tgt := s.decodeUnpackedGUID(e)

	s.SetGUID("Target", tgt)
	s.Map().PropagateChanges(s.GUID())
}

func (s *Session) SitChair(chair *GameObject) {
	chairPos := chair.Position()
	gobjt := s.GetGameObjectTemplateByEntry(chair.Entry())

	slots := gobjt.Data[0]
	height := gobjt.Data[1]

	if slots > 0 {
		lowestDist := s.Map().VisibilityDistance()

		xLowest := chairPos.X
		yLowest := chairPos.Y

		orthogOrientation := chairPos.O + float32(math.Pi)*0.5

		for i := uint32(0); i < slots; i++ {
			relDistance := (gobjt.Size*float32(i) - float32(gobjt.Size)*float32(slots-1)/2.0)

			xI := chairPos.X + relDistance*float32(math.Cos(float64(orthogOrientation)))
			yI := chairPos.X + relDistance*float32(math.Sin(float64(orthogOrientation)))

			thisDistance := s.Position().Point3.Dist2D(update.Point3{
				X: xI,
				Y: yI,
			})

			if thisDistance < lowestDist {
				lowestDist = thisDistance
				xLowest = xI
				yLowest = yI
			}
		}

		s.Teleport(s.CurrentMap, xLowest, yLowest, chairPos.Z, chairPos.O)
	} else {
		s.TeleportTo(s.CurrentMap, chairPos)
	}

	s.SetStandState(packet.UNIT_STAND_STATE_SIT_LOW_CHAIR + uint8(height))
}

func (s *Session) HandleSheathe(e *etc.Buffer) {
	state := e.ReadUint32()
	fmt.Println("sheathe", state)
	if state > 3 {
		state = 0
	}

	s.ValuesBlock.SetByte("SheathState", uint8(state))
	s.UpdatePlayer()
}
