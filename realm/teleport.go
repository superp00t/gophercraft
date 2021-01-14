package realm

import (
	"math"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/superp00t/etc"
	"github.com/superp00t/etc/yo"
	"github.com/superp00t/gophercraft/format/dbc"
	"github.com/superp00t/gophercraft/guid"
	"github.com/superp00t/gophercraft/packet"
	"github.com/superp00t/gophercraft/packet/update"
	"github.com/superp00t/gophercraft/realm/wdb"
	"github.com/superp00t/gophercraft/vsn"
)

func (s *Session) SendNewWorld(mapID uint32, pos update.Position) {
	pkt := packet.NewWorldPacket(packet.SMSG_NEW_WORLD)
	if s.Build().AddedIn(vsn.V1_12_1) {
		pkt.WriteUint32(mapID)
	} else {
		pkt.WriteByte(uint8(mapID))
	}
	update.EncodePosition(pkt.Buffer, pos)
	s.SendAsync(pkt)
}

func (s *Session) SendTeleportAck(g guid.GUID, mapID uint32, pos update.Position) {
	if s.Build().RemovedIn(vsn.V1_12_1) {
		pkt := packet.NewWorldPacket(packet.SMSG_MOVE_WORLDPORT_ACK)
		mi := &update.MovementInfo{
			TransportPosition: pos,
			Position:          pos,
		}
		update.EncodeMovementInfo(s.Build(), pkt, mi)
		s.SendAsync(pkt)
		return
	}

	pkt := packet.NewWorldPacket(packet.MSG_MOVE_TELEPORT_ACK)
	s.encodePackedGUID(pkt, g)
	pkt.WriteUint32(0)
	mi := &update.MovementInfo{
		Flags:    0,
		Position: pos,
	}
	update.EncodeMovementInfo(s.Build(), pkt.Buffer, mi)
	s.SendAsync(pkt)
}

func (s *Session) SendTransferPending(mapID uint32) {
	var pkt *packet.WorldPacket
	pkt = packet.NewWorldPacket(packet.SMSG_TRANSFER_PENDING)
	pkt.WriteUint32(mapID)
	s.SendAsync(pkt)
}

func (s *Session) HandleWorldportAck(b []byte) {
	// Resend inventory
	for _, v := range s.Inventory {
		s.SendObjectCreate(v)
	}

	// Tell the client they've successfully teleported.
	s.SendObjectCreate(s)

	// Notify our client of existing objects in this map.
	for _, obj := range s.Map().NearObjects(s) {
		s.SendObjectCreate(obj)
	}

	s.CurrentChunkIndex = nil
	s.CurrentArea = 0

	s.SyncTime()
	s.UpdateArea()
}

// TeleportTo teleports a player to a new location. This function should be called carefully.
func (s *Session) TeleportTo(mapID uint32, newPos update.Position) {
	yo.Ok("Teleporting", s.PlayerName(), "to", mapID, spew.Sdump(newPos))

	if mapID == s.CurrentMap {
		s.SendTeleportAck(s.GUID(), mapID, newPos)

		for _, sess := range s.Map().NearSet(s) {
			// if our new location is too far from this player
			if newPos.Dist2D(sess.Position().Point3) > s.WS.Config.Float32("Sync.VisibilityRange") {
				// make player s disappear in their client
				sess.SendObjectDelete(s.GUID())
			} else { // or else
				// make player s jump to new position in nearby player's client
				sess.SendTeleportAck(s.GUID(), mapID, newPos)
			}
		}

		// set position
		s.UpdatePosition(newPos)

		// Now that our new position is set, notify nearby players that we are here.
		for _, sess := range s.Map().NearSet(s) {
			sess.SendObjectCreate(s)
		}

		s.CurrentChunkIndex = nil
		s.CurrentArea = 0

		s.SyncTime()
		s.UpdateArea()
	} else {
		// Open appropriate loading screen for mapID
		s.SendTransferPending(mapID)

		// Remove from old map and notify other clients of this player's disappearance
		s.Map().RemoveObject(s.GUID())

		// tell player they are deleted
		s.SendObjectDelete(s.GUID())

		s.UpdatePosition(newPos)
		s.CurrentMap = mapID

		s.Map().AddObject(s)

		// trigger worldport ack once client is finished loading into zone
		s.SendNewWorld(mapID, newPos)
	}
}

func (s *Session) Teleport(mapID uint32, x, y, z, o float32) {
	s.TeleportTo(mapID, update.Position{update.Point3{x, y, z}, o})
}

// important todo: verify zone ID against geometric bounds of area:
// Just a dummy function until ADT parsing is available
func (s *Session) ValidateZoneWithPosition(zoneID uint32) bool {
	return true
}

func (s *Session) HandleZoneUpdate(e *etc.Buffer) {
	zoneID := e.ReadUint32()

	if !s.ValidateZoneWithPosition(zoneID) {
		return
	}

	_, err := s.DB().Cols("zone").Where("id = ?", s.PlayerID()).Update(&wdb.Character{
		Zone: zoneID,
	})

	if err != nil {
		panic(err)
	}

	s.ZoneID = zoneID
}

func (s *Session) HandleAreaTrigger(e *etc.Buffer) {
	triggerID := e.ReadUint32()

	var aTrigger *dbc.Ent_AreaTrigger
	s.DB().GetData(triggerID, &aTrigger)
	if aTrigger == nil {
		yo.Warn("Player tried to call non existent area trigger", triggerID)
		return
	}
	// delta is safe radius
	const delta float32 = 5.0

	pos := s.Position()

	if !isPointInAreaTriggerZone(aTrigger, s.CurrentMap, pos.X, pos.Y, pos.Z, delta) {
		yo.Warn("Player", s.PlayerName(), "tried to teleport to area trigger", triggerID, "without being in correct position")
		return
	}

	yo.Ok("Area trigger", triggerID)

	s.WS.ThinkOn(AreaTriggerEvent, triggerID, s)
}

func (s *Session) SendRequiredLevelZoneError(lvl int) {
	s.Alertf("You must be at least level %d to enter this zone.", lvl)
}

func (s *Session) SendRequiredItemZoneError(itemID string) {
	tpl := s.GetItemTemplate(wdb.Item{
		ItemID: itemID,
	})
	s.Alertf("You must have %s to enter this zone.", tpl.Name)
}

func (s *Session) SendRequiredQuestZoneError(questEntry uint32) {
	// todo: send quest name
	s.Alertf("You must solve a quest to enter this zone.")
}

func fabs(i float32) float32 {
	return float32(math.Abs(float64(i)))
}

func isPointInAreaTriggerZone(atEntry *dbc.Ent_AreaTrigger, mapID uint32, x, y, z, delta float32) bool {
	if mapID != atEntry.MapID {
		yo.Warn("incorrect mapID", mapID)
		return false
	}

	if atEntry.Radius > 0 {
		// if we have radius check it
		dist2 := (x-atEntry.X)*(x-atEntry.X) + (y-atEntry.Y)*(y-atEntry.Y) + (z-atEntry.Z)*(z-atEntry.Z)
		if dist2 > (atEntry.Radius+delta)*(atEntry.Radius+delta) {
			return false
		}
	} else {
		// we have only extent

		// rotate the players position instead of rotating the whole cube, that way we can make a simplified
		// is-in-cube check and we have to calculate only one point instead of 4

		// 2PI = 360, keep in mind that ingame orientation is counter-clockwise
		rotation := float64(2)*math.Pi - float64(atEntry.BoxO)
		sinVal := float32(math.Sin(rotation))
		cosVal := float32(math.Cos(rotation))

		playerBoxDistX := x - atEntry.X
		playerBoxDistY := y - atEntry.Y

		rotPlayerX := float32(atEntry.X + playerBoxDistX*cosVal - playerBoxDistY*sinVal)
		rotPlayerY := float32(atEntry.Y + playerBoxDistY*cosVal + playerBoxDistX*sinVal)

		// box edges are parallel to coordiante axis, so we can treat every dimension independently :D
		dz := z - atEntry.Z
		dx := rotPlayerX - atEntry.X
		dy := rotPlayerY - atEntry.Y
		if (fabs(dx) > atEntry.BoxX/2+delta) ||
			(fabs(dy) > atEntry.BoxY/2+delta) ||
			(fabs(dz) > atEntry.BoxO/2+delta) {
			return false
		}
	}

	return true
}

func (c *Session) PhaseTeleportTo(phaseID string, mapID uint32, pos update.Position) {
	if phaseID == c.CurrentPhase {
		c.TeleportTo(mapID, pos)
	} else {
		panic("nyi")
	}
}

func (c *Session) HandleSummonResponse(e *etc.Buffer) {
	if !c.IsAlive() {
		return
	}

	if c.summons == nil {
		return
	}

	c.PhaseTeleportTo(c.summons.Phase, c.summons.Map, c.summons.Pos)
	c.summons = nil
}

type summons struct {
	Phase string
	Map   uint32
	Pos   update.Position
}

func (s *Session) SetSummonLocation(phase string, mapID uint32, pos update.Position) {
	s.summons = &summons{
		phase,
		mapID,
		pos,
	}
}

func (s *Session) SendSummonRequest(summoner guid.GUID, zoneID uint32, timeout time.Duration) {
	pkt := packet.NewWorldPacket(packet.SMSG_SUMMON_REQUEST)
	summoner.EncodeUnpacked(s.Build(), pkt)
	pkt.WriteUint32(zoneID)
	pkt.WriteUint32(uint32(timeout / time.Millisecond))
	s.SendAsync(pkt)
}

func (s *Session) HandleWorldTeleport(e *etc.Buffer) {
	if !s.IsGM() {
		return
	}

	// time := e.ReadUint32()
	e.ReadUint32()

	var mapID uint32
	if s.Build().AddedIn(vsn.V1_12_1) {
		mapID = e.ReadUint32()
	} else {
		// Alpha shenanigans
		mapID = uint32(e.ReadByte())
	}

	var newPos update.Position
	newPos.X = e.ReadFloat32()
	newPos.Y = e.ReadFloat32()
	newPos.Z = e.ReadFloat32()
	newPos.O = e.ReadFloat32()

	s.TeleportTo(mapID, newPos)
}
