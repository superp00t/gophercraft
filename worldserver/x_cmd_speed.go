package worldserver

import (
	"github.com/superp00t/gophercraft/guid"
	"github.com/superp00t/gophercraft/packet"
	"github.com/superp00t/gophercraft/packet/update"
)

type SpeedOpcodePair struct {
	Force  packet.WorldType
	Spline packet.WorldType
}

var (
	SpeedCodes = map[update.SpeedType]SpeedOpcodePair{
		update.Walk:         {packet.SMSG_FORCE_WALK_SPEED_CHANGE, packet.SMSG_SPLINE_SET_WALK_SPEED},
		update.Run:          {packet.SMSG_FORCE_RUN_SPEED_CHANGE, packet.SMSG_SPLINE_SET_RUN_SPEED},
		update.RunBackward:  {packet.SMSG_FORCE_RUN_BACK_SPEED_CHANGE, packet.SMSG_SPLINE_SET_RUN_BACK_SPEED},
		update.Swim:         {packet.SMSG_FORCE_SWIM_SPEED_CHANGE, packet.SMSG_SPLINE_SET_SWIM_SPEED},
		update.SwimBackward: {packet.SMSG_FORCE_SWIM_BACK_SPEED_CHANGE, packet.SMSG_SPLINE_SET_SWIM_BACK_SPEED},
		update.Turn:         {packet.SMSG_FORCE_TURN_RATE_CHANGE, packet.SMSG_SPLINE_SET_TURN_RATE},
	}
)

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

	speeds := obj.Speeds()

	if g.HighType() == guid.Player {
		sess := obj.(*Session)
		pkt := packet.NewWorldPacket(pair.Force)
		sess.encodePackedGUID(pkt, g)
		pkt.WriteUint32(0)
		pkt.WriteFloat32(speeds[st])
		sess.SendAsync(pkt)
	}

	pkt := packet.NewWorldPacket(pair.Spline)
	g.EncodePacked(m.Phase.Server.Config.Version, pkt)
	pkt.WriteFloat32(speeds[st])

	m.NearSet(obj).Iter(func(s *Session) {
		s.SendAsync(pkt)
	})
}

func x_Speed(c *C) {
	speed := c.Float32(0)
	if speed < .1 || speed > 50 {
		c.Session.Warnf("speed must be [0.1 - 50.0]")
		return
	}

	if speed == 0 {
		speed = 1
	}

	c.Session.PlayerSpeeds[update.Run] = speed
	c.Session.Map().UpdateSpeed(c.Session.GUID(), update.Run)
}
