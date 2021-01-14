package realm

import (
	"fmt"
	"sync"

	"github.com/superp00t/etc"

	"github.com/superp00t/gophercraft/gcore/config"
	"github.com/superp00t/gophercraft/packet/update"

	"github.com/superp00t/gophercraft/guid"
	"github.com/superp00t/gophercraft/packet"
	_ "github.com/superp00t/gophercraft/packet/update/descriptorsupport"
	"github.com/superp00t/gophercraft/realm/wdb"
)

type SessionSet []*Session

// Phase describes a plane of existence which contains multiple maps.
type Phase struct {
	sync.Mutex
	Server *Server
	Maps   map[uint32]*Map
}

type InstanceScripting interface {
	AddCreature(id string, mapID uint32, x, y, z, o float32)
}

// Map describes a single map in the world
type Map struct {
	sync.Mutex

	Phase   *Phase
	Objects map[guid.GUID]WorldObject
}

type Object interface {
	GUID() guid.GUID
	TypeID() guid.TypeID
	// Values should return a pointer to underlying data, not a reference to a copy.
	Values() *update.ValuesBlock
}

// Objects that have a presence in the world in a specific location (players, creatures)
type WorldObject interface {
	Object
	// movement data
	Movement() *update.MovementBlock
}

// TODO: delete phases when players are not there.
func (ws *Server) Phase(id string) *Phase {
	ws.PhaseL.Lock()
	ph := ws.Phases[id]
	if ph == nil {
		ph = &Phase{
			Maps:   make(map[uint32]*Map),
			Server: ws,
		}

		ws.Phases[id] = ph
	}
	ws.PhaseL.Unlock()
	return ph
}

func (ph *Phase) Map(i uint32) *Map {
	ph.Lock()
	defer ph.Unlock()

	if ph.Maps[i] == nil {
		mp := new(Map)
		mp.Phase = ph
		mp.Objects = make(map[guid.GUID]WorldObject)
		ph.Maps[i] = mp
	}

	return ph.Maps[i]
}

const (
	UseCompressionSmartly int = iota
	ForceCompressionOff
	ForceCompressionOn
)

const updatePacketLengthThreshold = 100

func (s *Session) SendRawUpdateObjectData(encoded []byte, forceCompression int) {
	// initialize uncompressed packet, sending this is sometimes more efficient than enabling zlib compression
	sPacket := packet.NewWorldPacket(packet.SMSG_UPDATE_OBJECT)

	var compressionEnabled = false

	// detect if compression has been forcefully disabled/enabled
	switch forceCompression {
	case UseCompressionSmartly:
		// compression is disabled if there is no benefit
		compressionEnabled = len(encoded) > updatePacketLengthThreshold
	case ForceCompressionOff:
		compressionEnabled = false
	case ForceCompressionOn:
		compressionEnabled = true
	}

	uncompressedLength := uint32(len(encoded))

	if compressionEnabled {
		sPacket = packet.NewWorldPacket(packet.SMSG_COMPRESSED_UPDATE_OBJECT)
		compressedData := packet.Compress(encoded)
		sPacket.WriteUint32(uncompressedLength)
		sPacket.Write(compressedData)

		if s.HasProp(ObjectDebug) {
			compressionRatio := float64(len(compressedData)) / float64(uncompressedLength)

			s.printfObjMgr("Sending compressed update. %d => %d bytes compression ratio: %f", uncompressedLength, sPacket.Len(), compressionRatio)
		}

	} else {
		if s.HasProp(ObjectDebug) {
			s.printfObjMgr("Sending uncompressed update. %d bytes", uncompressedLength)
		}
		sPacket.Write(encoded)
	}

	s.SendAsync(sPacket)
}

func (m *Map) GetObject(g guid.GUID) WorldObject {
	m.Lock()
	wo := m.Objects[g]
	m.Unlock()
	return wo
}

func (m *Map) AddObject(obj WorldObject) error {
	m.Lock()
	m.Objects[obj.GUID()] = obj
	m.Unlock()

	// Send spawn packet to nearby players.
	for _, v := range m.NearSet(obj) {
		v.SendObjectCreate(obj)
		v.AddTrackedGUID(obj.GUID())
	}

	return nil
}

func (m *Map) RemoveObject(id guid.GUID) {
	m.Lock()

	obj := m.Objects[id]
	if obj == nil {
		m.Unlock()
		return
	}

	delete(m.Objects, id)
	m.Unlock()

	for _, v := range m.NearSet(obj) {
		v.SendObjectDelete(id)
		v.RemoveTrackedGUID(id)
	}
}

type WorldObjectSet []WorldObject

func (m *Map) VisibilityDistance() float32 {
	return m.Config().Float32("Sync.VisibilityRange")
}

func (m *Map) NearObjects(nearTo WorldObject) WorldObjectSet {
	return m.NearObjectsLimit(nearTo, m.VisibilityDistance())
}

func (m *Map) NearObjectsLimit(nearTo WorldObject, limit float32) WorldObjectSet {
	m.Lock()
	var wo []WorldObject
	for _, obj := range m.Objects {
		if nearTo.GUID() != obj.GUID() {
			if nearTo.Movement().Position.Dist3D(obj.Movement().Position.Point3) <= limit {
				wo = append(wo, obj)
			}
		}
	}
	m.Unlock()
	return wo
}

func (wos WorldObjectSet) Iter(iterFunc func(WorldObject)) {
	for _, wo := range wos {
		iterFunc(wo)
	}
}

// NeaSet enumerates a list of players close to (less than or equal to world.maxVisibilityRange) a game object on a map.
func (m *Map) NearSet(nearTo WorldObject) SessionSet {
	return m.NearSetLimit(nearTo, m.VisibilityDistance())
}

func (m *Map) NearSetLimit(nearTo WorldObject, limit float32) SessionSet {
	m.Lock()
	var s []*Session
	for _, plyr := range m.Objects {
		switch session := plyr.(type) {
		case *Session:
			if session.GUID() != nearTo.GUID() {
				if nearTo.Movement().Position.Dist3D(plyr.Movement().Position.Point3) <= limit {
					s = append(s, session)
				}
			}
		}
	}
	m.Unlock()
	return s
}

func (s *Session) SendObjectChanges(viewMask update.VisibilityFlags, object Object) {
	packet := etc.NewBuffer()

	enc, err := update.NewEncoder(s.Build(), packet, 1)
	if err != nil {
		panic(err)
	}

	if s.HasProp(ObjectDebug) {
		s.printfObjMgr("Updating %s: %s", object.GUID(), object.Values().ChangeMask)
	}

	if err = enc.AddBlock(object.GUID(), object.Values(), viewMask); err != nil {
		panic(err)
	}

	s.SendRawUpdateObjectData(packet.Bytes(), 0)
}

func setNumBlocks(buffer *etc.Buffer, num int) {
	wpos := buffer.Wpos()
	buffer.SeekW(0)
	buffer.WriteUint32(uint32(num))
	buffer.SeekW(wpos)
}

// Send spawn packets for 1+ objects. This function is optimized to fit as many Create/Spawn blocks into one compressed packet as possible.
func (s *Session) SendObjectCreate(objects ...Object) {
	if len(objects) == 0 {
		return
	}

	if s.HasProp(ObjectDebug) {
		s.printfObjMgr("Spawning %d objects", len(objects))
	}

	uPacket := etc.NewBuffer()

	const compression = UseCompressionSmartly

	// The number of blocks will be written afterward with setNumBlocks, use zero as placeholder value.
	enc, err := update.NewEncoder(s.Build(), uPacket, 0)
	if err != nil {
		panic(err)
	}

	var numBlocks int

	for _, wo := range objects {
		// Which fields should be included?
		viewMask := s.WS.queryRelationshipMask(wo.GUID(), s.GUID())

		movementBlock := &update.MovementBlock{}
		if wobj, ok := wo.(WorldObject); ok {
			movementBlock = wobj.Movement()
			if wo.GUID() == s.GUID() {
				movementBlock.UpdateFlags |= update.UpdateFlagSelf
			}
		}

		blockType := update.CreateObject

		if wo.TypeID() == guid.TypeUnit || wo.TypeID() == guid.TypePlayer {
			blockType = update.SpawnObject
		}

		// Guesstimate whether adding this block will overflow the maximum packet length (doesn't need to be perfect)
		if uPacket.Len()+5000 > packet.MaxLength {
			setNumBlocks(uPacket, numBlocks)

			if s.HasProp(ObjectDebug) {
				s.printfObjMgr("Sending object update (uncompressed: %d bytes, %d blocks)", uPacket.Len(), numBlocks)
			}

			// If so, send packet and create a new one
			s.SendRawUpdateObjectData(uPacket.Bytes(), compression)
			uPacket = etc.NewBuffer()

			numBlocks = 0

			enc, err = update.NewEncoder(s.Build(), uPacket, 0)
			if err != nil {
				panic(err)
			}
		}

		if s.HasProp(ObjectDebug) {
			s.printfObjMgr("Creating %s", wo.GUID())
		}

		// Serialize create data to buffer
		if err = enc.AddBlock(wo.GUID(), &update.CreateBlock{
			BlockType:     blockType,
			ObjectType:    wo.TypeID(),
			MovementBlock: movementBlock,
			ValuesBlock:   wo.Values(),
		}, viewMask); err != nil {
			panic(err)
		}

		numBlocks++

		if uPacket.Len() > packet.MaxLength {
			panic("maximum packet length for object creates exceeded")
		}
	}

	if uPacket.Len() > 0 {
		if s.HasProp(ObjectDebug) {
			s.printfObjMgr("Sending object update (uncompressed: %d bytes, %d blocks)", uPacket.Len(), numBlocks)
		}
		setNumBlocks(uPacket, numBlocks)
		s.SendRawUpdateObjectData(uPacket.Bytes(), compression)
	}
}

func (s *Session) SendObjectDelete(g guid.GUID) {
	if s.HasProp(ObjectDebug) {
		s.printfObjMgr("Deleting %s", g)
	}

	pkt := packet.NewWorldPacket(packet.SMSG_DESTROY_OBJECT)
	g.EncodeUnpacked(s.Build(), pkt)
	s.SendAsync(pkt)
}

func (s *Session) SendAreaAll(p *packet.WorldPacket) {
	s.SendAsync(p)

	// broadcast
	s.Map().NearSet(s).Send(p)
}

func (s SessionSet) Send(p *packet.WorldPacket) {
	for _, v := range s {
		v.SendAsync(p)
	}
}

func (s SessionSet) Iter(iterFunc func(*Session)) {
	for _, sess := range s {
		iterFunc(sess)
	}
}

func (ws *Server) queryRelationshipMask(src, target guid.GUID) update.VisibilityFlags {
	if src == target {
		return update.Owner
	}

	// todo: determine party relationship

	return 0
}

func (m *Map) PropagateChanges(id guid.GUID) {
	m.Lock()
	o := m.Objects[id]
	m.Unlock()

	valuesStore := o.Values()

	// if the Object is a player session, send them their own changes.
	if session, ok := o.(*Session); ok {
		session.SendObjectChanges(m.Phase.Server.queryRelationshipMask(o.GUID(), o.GUID()), session)
	}

	// transmit appropriate changes.
	for _, v := range m.NearSet(o) {
		v.SendObjectChanges(m.Phase.Server.queryRelationshipMask(o.GUID(), v.GUID()), o)
	}

	valuesStore.ClearChanges()
}

func (m *Map) Config() *config.World {
	return m.Phase.Server.Config
}

// Send updated fields directly to client. Use for setting private fields.
func (s *Session) UpdateSelf() {
	s.SendObjectChanges(update.Owner, s)
	s.ClearChanges()
}

// Broadcast changes to nearby players
func (s *Session) UpdatePlayer() {
	s.Map().PropagateChanges(s.GUID())
}

func (ws *Server) AllSessions() SessionSet {
	ss := make(SessionSet, len(ws.PlayerList))
	index := 0
	for _, v := range ws.PlayerList {
		ss[index] = v
		index++
	}
	return ss
}

func (ws *Server) NextDynamicCounter(typeID guid.TypeID) uint64 {
	ws.GuardCounters.Lock()
	next := ws.DynamicCounters[typeID] + 1
	ws.DynamicCounters[typeID] = next
	ws.GuardCounters.Unlock()
	return next
}

func (s *Session) AddTrackedGUID(g guid.GUID) {
	s.GuardTrackedGUIDs.Lock()
	defer s.GuardTrackedGUIDs.Unlock()

	if s.isTrackedGUID(g) {
		return
	}

	s.TrackedGUIDs = append(s.TrackedGUIDs, g)
}

func (s *Session) RemoveTrackedGUID(g guid.GUID) {
	s.GuardTrackedGUIDs.Lock()
	defer s.GuardTrackedGUIDs.Unlock()

	idx := -1

	for i, v := range s.TrackedGUIDs {
		if v == g {
			idx = i
			break
		}
	}

	if idx >= 0 {
		s.TrackedGUIDs = append(s.TrackedGUIDs[:idx], s.TrackedGUIDs[idx+1:]...)
	}
}

func (ph *Phase) AddCreature(id string, mapID uint32, x, y, z, o float32) {
	var cr *wdb.CreatureTemplate
	ph.Server.DB.GetData(id, &cr)
	if cr == nil {
		panic(fmt.Errorf("No CreatureTemplate could be found with the ID %s", id))
		return
	}

	creature := ph.Server.NewCreature(cr, update.Position{
		Point3: update.Point3{
			x, y, z,
		},
		O: o,
	})
	ph.Map(mapID).AddObject(creature)
}
