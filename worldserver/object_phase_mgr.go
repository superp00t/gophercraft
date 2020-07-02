package worldserver

import (
	"fmt"
	"sync"

	"github.com/superp00t/etc"

	"github.com/superp00t/gophercraft/gcore/config"
	"github.com/superp00t/gophercraft/packet/update"

	"github.com/superp00t/gophercraft/guid"
	"github.com/superp00t/gophercraft/packet"
	_ "github.com/superp00t/gophercraft/packet/update/descriptorsupport"
)

type SessionSet []*Session

// Phase describes a dimension which contains multiple maps.
type Phase struct {
	sync.Mutex
	Server *WorldServer
	Maps   map[uint32]*Map
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
	// Values should return a mutable reference, not a generated copy of the underlying data.
	Values() *update.ValuesBlock
}

// Objects that have a presence in the world in a specific location, i.e. not Items or Containers which have no position.
type WorldObject interface {
	Object
	// movement data
	Movement() *update.MovementBlock
}

// TODO: delete phases when players are not there.
func (ws *WorldServer) Phase(id string) *Phase {
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

func (s *Session) SendRawUpdateObjectData(encoded []byte, forceCompression int) {
	// initialize uncompressed packet, sending this is sometimes more efficient than enabling zlib compression
	sPacket := packet.NewWorldPacket(packet.SMSG_UPDATE_OBJECT)

	var compressionEnabled = false

	// detect if compression has been forcefully disabled/enabled
	switch forceCompression {
	case UseCompressionSmartly:
		// compression is enabled if the packet is over 512 bytes.
		compressionEnabled = len(encoded) > 100
	case ForceCompressionOff:
		compressionEnabled = false
	case ForceCompressionOn:
		compressionEnabled = true
	}

	if compressionEnabled {
		sPacket = packet.NewWorldPacket(packet.SMSG_COMPRESSED_UPDATE_OBJECT)
		compressedData := packet.Compress(encoded)
		sPacket.WriteUint32(uint32(len(encoded)))
		sPacket.Write(compressedData)
	} else {
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
			if nearTo.Movement().Position.Dist2D(obj.Movement().Position.Point3) <= limit {
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
				if nearTo.Movement().Position.Dist2D(plyr.Movement().Position.Point3) <= limit {
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

	if err = enc.AddBlock(object.GUID(), object.Values(), viewMask); err != nil {
		panic(err)
	}

	s.SendRawUpdateObjectData(packet.Bytes(), 0)
}

func (s *Session) SendObjectCreate(wo Object) {
	fmt.Println("Creating", wo.GUID())
	name, _ := s.WS.GetUnitNameByGUID(wo.GUID())

	if s.objectDebug {
		s.Warnf("Sending create of %s (%s)", wo.GUID(), name)

		sass, ok := wo.(*Session)
		if ok {
			s.Warnf("Map: %d", sass.CurrentMap)
		}

		wobj, ok := wo.(WorldObject)
		if ok {
			s.Warnf("Position: %+v", wobj.Movement().Position)
		}
	}

	viewMask := s.WS.queryRelationshipMask(wo.GUID(), s.GUID())
	packet := etc.NewBuffer()

	enc, err := update.NewEncoder(s.Build(), packet, 1)
	if err != nil {
		panic(err)
	}

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

	if err = enc.AddBlock(wo.GUID(), &update.CreateBlock{
		BlockType:     blockType,
		ObjectType:    wo.TypeID(),
		MovementBlock: movementBlock,
		ValuesBlock:   wo.Values(),
	}, viewMask); err != nil {
		panic(err)
	}

	s.SendRawUpdateObjectData(packet.Bytes(), 0)
}

func (s *Session) SendObjectDelete(g guid.GUID) {
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

func (ws *WorldServer) queryRelationshipMask(src, target guid.GUID) update.VisibilityFlags {
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

	// if the Object is a player session
	if session, ok := o.(*Session); ok {
		session.SendObjectChanges(m.Phase.Server.queryRelationshipMask(o.GUID(), o.GUID()), session)
	}

	// transmit appropriate changes.
	for _, v := range m.NearSet(o) {
		v.SendObjectChanges(m.Phase.Server.queryRelationshipMask(o.GUID(), v.GUID()), o)
	}

	valuesStore.ClearChanges()
}

func (m *Map) PlaySound(id uint32) {
	m.Lock()
	for _, v := range m.Objects {
		if s, ok := v.(*Session); ok {
			s.SendPlaySound(id)
		}
	}
	m.Unlock()
}

func (s *Session) SendPlaySound(id uint32) {
	pkt := packet.NewWorldPacket(packet.SMSG_PLAY_SOUND)
	pkt.WriteUint32(id)
	s.SendAsync(pkt)
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
