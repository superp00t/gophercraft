package realm

import (
	"fmt"

	"github.com/superp00t/etc"
	"github.com/superp00t/etc/yo"
	"github.com/superp00t/gophercraft/guid"
	"github.com/superp00t/gophercraft/packet"
	"github.com/superp00t/gophercraft/packet/update"
	"github.com/superp00t/gophercraft/vsn"
)

func (s *Session) SetFly(on bool) {
	if s.Build().RemovedIn(vsn.V2_4_3) {
		// Hacky bullshit.
		// Flying wasn't actually implemented until version 2.0.
		// You can only move laterally and it's buggy as hell
		mask := update.MoveFlagCanFly | update.MoveFlagFlying

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
	yo.Spew(minfo)
	// TODO: validate flags
	s.MovementInfo = minfo
	s.UpdatePosition(s.MovementInfo.Position)
}

func (s *Session) UpdatePosition(pos update.Position) {
	// if pos.X == 0 && pos.Y == 0 && pos.Z == 0 {
	// 	panic("zero set")
	// }

	s.MovementInfo.Position = pos

	s.Char.Map = s.CurrentMap
	s.Char.X = s.MovementInfo.Position.X
	s.Char.Y = s.MovementInfo.Position.Y
	s.Char.Z = s.MovementInfo.Position.Z
	s.Char.O = s.MovementInfo.Position.O

	if _, err := s.DB().Where("id = ?", s.PlayerID()).Cols("x", "y", "z", "o", "map", "zone").Update(s.Char); err != nil {
		panic(err)
	}

	s.UpdateCameraPosition(true, pos)

	if s.Group != nil {
		s.Group.Lock()
		for _, v := range s.Group.Members {
			if v != s.GUID() {
				if partyMember, err := s.WS.GetSessionByGUID(v); err == nil {
					partyMember.SendPartyMemberStats(statsMaskAll, s)
				}
			}
		}
		s.Group.Unlock()
	}
}

func (s *Session) UpdateCameraPosition(syncSelf bool, pos update.Position) {
	visRange := s.WS.Config.Float32("Sync.VisibilityRange")

	s.GuardTrackedGUIDs.Lock()
	defer s.GuardTrackedGUIDs.Unlock()

	tGuids := s.TrackedGUIDs[:0]

	// TODO: time out object synchronization if the player only moved a little bit, or has moved very recently

	// Remove GUIDs if too far away
	for _, g := range s.TrackedGUIDs {
		active := s.Map().GetObject(g)
		if active == nil {
			s.SendObjectDelete(g)
		} else {
			if active.Movement().Position.Dist3D(s.Position().Point3) > visRange {
				s.SendObjectDelete(g)

				if syncSelf {
					if activeSession, ok := (active).(*Session); ok {
						activeSession.RemoveTrackedGUID(s.GUID())
						activeSession.SendObjectDelete(s.GUID())
					}
				}
			} else {
				tGuids = append(tGuids, g)
			}
		}
	}

	s.TrackedGUIDs = tGuids

	// Add new GUIDs if not found yet
	for _, nearObject := range s.Map().NearObjectsLimit(s, visRange) {
		if !s.isTrackedGUID(nearObject.GUID()) {
			s.TrackedGUIDs = append(s.TrackedGUIDs, nearObject.GUID())
			s.SendObjectCreate(nearObject)

			if syncSelf {
				// TODO: it may be more efficient to send these creates as multiple blocks in a SMSG_COMPRESSED_OBJECT_UPDATE
				// We may be appearing to a new player. Notify them of us.
				if nearSession, ok := (nearObject).(*Session); ok {
					nearSession.SendObjectCreate(s)
				}
			}
		}
	}
}

func (s *Session) HandleMoves(t packet.WorldType, b []byte) {
	data := etc.FromBytes(b)

	if s.Build().AddedIn(vsn.V3_3_5a) {
		gid := s.decodePackedGUID(data)
		fmt.Println("Move", gid)
	}

	e, err := update.DecodeMovementInfo(s.Build(), data)
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

func (s *Session) Movement() *update.MovementBlock {
	s.MovementInfo.Time = s.WS.UptimeMS()

	mData := &update.MovementBlock{
		Speeds:   s.MoveSpeeds,
		Position: s.MovementInfo.Position,
		Info:     s.MovementInfo,
	}

	mData.UpdateFlags |= update.UpdateFlagLiving
	mData.UpdateFlags |= update.UpdateFlagHasPosition

	mData.UpdateFlags |= update.UpdateFlagAll
	mData.UpdateFlags |= update.UpdateFlagHighGUID
	mData.All = 0x1 // 5875 only
	mData.HighGUID = 0x1
	return mData
}

func (s *Session) Position() update.Position {
	return s.MovementInfo.Position
}

func (s *Session) Speeds() update.Speeds {
	return s.MoveSpeeds
}

func (m *Map) UpdateSpeed(g guid.GUID, st update.SpeedType) {
	pair, ok := SpeedCodes[st]
	if !ok {
		return
	}

	m.Lock()
	obj := m.Objects[g]
	m.Unlock()

	if obj == nil {
		return
	}

	speeds := obj.Movement().Speeds

	if g.HighType() == guid.Player {
		sess := obj.(*Session)
		pkt := packet.NewWorldPacket(pair.Force)
		sess.encodePackedGUID(pkt, g)
		pkt.WriteUint32(0)
		if st == update.Run && sess.Build().AddedIn(vsn.V2_4_3) {
			pkt.WriteByte(0)
		}
		pkt.WriteFloat32(speeds[st])
		sess.SendAsync(pkt)
	}

	pkt := packet.NewWorldPacket(pair.Spline)
	g.EncodePacked(m.Phase.Server.Build(), pkt)
	pkt.WriteFloat32(speeds[st])

	m.NearSet(obj).Iter(func(s *Session) {
		s.SendAsync(pkt)
	})
}

type SpeedOpcodePair struct {
	Force  packet.WorldType
	Spline packet.WorldType
}

var (
	SpeedCodes = map[update.SpeedType]SpeedOpcodePair{
		update.Walk:           {packet.SMSG_FORCE_WALK_SPEED_CHANGE, packet.SMSG_SPLINE_SET_WALK_SPEED},
		update.Run:            {packet.SMSG_FORCE_RUN_SPEED_CHANGE, packet.SMSG_SPLINE_SET_RUN_SPEED},
		update.RunBackward:    {packet.SMSG_FORCE_RUN_BACK_SPEED_CHANGE, packet.SMSG_SPLINE_SET_RUN_BACK_SPEED},
		update.Swim:           {packet.SMSG_FORCE_SWIM_SPEED_CHANGE, packet.SMSG_SPLINE_SET_SWIM_SPEED},
		update.SwimBackward:   {packet.SMSG_FORCE_SWIM_BACK_SPEED_CHANGE, packet.SMSG_SPLINE_SET_SWIM_BACK_SPEED},
		update.Turn:           {packet.SMSG_FORCE_TURN_RATE_CHANGE, packet.SMSG_SPLINE_SET_TURN_RATE},
		update.Flight:         {packet.SMSG_FORCE_FLIGHT_SPEED_CHANGE, packet.SMSG_SPLINE_SET_FLIGHT_SPEED},
		update.FlightBackward: {packet.SMSG_FORCE_FLIGHT_BACK_SPEED_CHANGE, packet.SMSG_SPLINE_SET_FLIGHT_BACK_SPEED},
	}
)

func (s *Session) SyncSpeeds() {
	for _, v := range update.SpeedLists[s.Build()] {
		s.Map().UpdateSpeed(s.GUID(), v)
	}
	s.SyncTime()
}
