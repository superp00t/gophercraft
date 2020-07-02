package worldserver

import (
	"github.com/superp00t/etc"
	"github.com/superp00t/etc/yo"
	"github.com/superp00t/gophercraft/packet"
	"github.com/superp00t/gophercraft/packet/update"
	"github.com/superp00t/gophercraft/vsn"
)

func (s *Session) SetFly(on bool) {
	mask := update.MoveFlagCanFly | update.MoveFlagFlying

	if s.Build().RemovedIn(vsn.V2_4_3) {
		if on {
			s.MovementInfo.Flags |= mask
		} else {
			s.MovementInfo.Flags &= ^mask
		}

		out := packet.NewWorldPacket(packet.MSG_MOVE_START_SWIM)
		s.encodePackedGUID(out, s.GUID())
		if err := update.EncodeMovementInfo(s.Build(), out, s.MovementInfo); err != nil {
			panic(err)
		}
		s.SendAsync(out)
	} else {
		var wType packet.WorldType
		if on {
			wType = packet.SMSG_MOVE_SET_CAN_FLY
		} else {
			wType = packet.SMSG_MOVE_UNSET_CAN_FLY
		}

		out := packet.NewWorldPacket(wType)
		s.encodePackedGUID(out, s.GUID())
		out.WriteUint32(0)
		s.SendAsync(out)
	}
}

func (s *Session) HandleUpdateMovement(minfo *update.MovementInfo) {
	// TODO: validate flags
	s.MovementInfo = minfo
}

func (s *Session) UpdatePosition(pos update.Position) {
	s.MovementInfo.Position = pos
}

func (s *Session) HandleMoves(t packet.WorldType, b []byte) {
	e, err := update.DecodeMovementInfo(s.Build(), etc.FromBytes(b))
	if err != nil {
		yo.Warn(err)
		return
	}

	// Important TODO: validate position
	s.HandleUpdateMovement(e)

	for _, v := range s.Map().NearSet(s) {
		// yo.Ok("Relaying moves", t, s.Char.Name, "->", v.Char.Name)
		out := packet.NewWorldPacket(t)
		s.encodePackedGUID(out, s.GUID())
		update.EncodeMovementInfo(v.Build(), out.Buffer, e)
		v.SendAsync(out)
	}
}
