package worldserver

import (
	"github.com/davecgh/go-spew/spew"
	"github.com/superp00t/etc/yo"
	"github.com/superp00t/gophercraft/guid"
	"github.com/superp00t/gophercraft/packet"
	"github.com/superp00t/gophercraft/packet/update"
)

func (s *Session) SendNewWorld(mapID uint32, pos update.Quaternion) {
	pkt := packet.NewWorldPacket(packet.SMSG_NEW_WORLD)
	pkt.WriteUint32(mapID)
	update.EncodeQuaternion(pkt.Buffer, pos)
	s.SendAsync(pkt)
}

func (s *Session) SendTeleportAck(g guid.GUID, mapID uint32, pos update.Quaternion) {
	pkt := packet.NewWorldPacket(packet.MSG_MOVE_TELEPORT_ACK)
	s.encodePackedGUID(pkt, g)
	pkt.WriteUint32(0)
	mi := &update.MovementInfo{
		Flags:    0,
		Position: pos,
	}
	update.EncodeMovementInfo(s.Version(), pkt.Buffer, mi)
	s.SendAsync(pkt)
}

func (s *Session) SendTransferPending(mapID uint32) {
	var pkt *packet.WorldPacket
	pkt = packet.NewWorldPacket(packet.SMSG_TRANSFER_PENDING)
	pkt.WriteUint32(mapID)
	s.SendAsync(pkt)
}

func (s *Session) HandleWorldportAck(b []byte) {
	// Tell the client they've successfully teleported.
	s.SendObjectCreate(s)
}

// TeleportTo teleports a player to a new location. This function should be called carefully.
func (s *Session) TeleportTo(mapID uint32, newPos update.Quaternion) {
	yo.Ok("Teleporting", s.PlayerName(), "to", mapID, spew.Sdump(newPos))

	if mapID == s.CurrentMap {
		s.SendTeleportAck(s.GUID(), mapID, newPos)

		for _, sess := range s.Map().NearbySessions(s) {
			// if our new location is too far from this player
			if newPos.Dist2D(sess.Position().Point3) > s.WS.Config.Float32("world.maxVisibilityRange") {
				// make player s disappear in their client
				sess.SendObjectDelete(s.GUID())
			} else { // or else
				// make player s jump to new position in nearby player's client
				sess.SendTeleportAck(s.GUID(), mapID, newPos)
			}
		}

		// set position
		s.PlayerPosition = newPos

		// Now that our new position is set, notify nearby players that they're here.
		for _, sess := range s.Map().NearbySessions(s) {
			sess.SendObjectCreate(s)
		}
	} else {
		// Open appropriate loading screen for mapID
		s.SendTransferPending(mapID)

		// Remove from old map and notify other clients of this player's disappearance
		s.Map().RemoveObject(s.GUID())

		// tell player they are deleted
		s.SendObjectDelete(s.GUID())

		s.PlayerPosition = newPos
		s.CurrentMap = mapID

		s.Map().AddObject(s)

		// trigger worldport ack once client is finished loading into zone
		s.SendNewWorld(mapID, newPos)
	}
}
