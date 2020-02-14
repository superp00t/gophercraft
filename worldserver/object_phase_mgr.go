package worldserver

import (
	"sync"

	"github.com/davecgh/go-spew/spew"
	"github.com/superp00t/gophercraft/gcore/config"
	"github.com/superp00t/gophercraft/packet/update"

	"github.com/superp00t/etc/yo"
	"github.com/superp00t/gophercraft/guid"
	"github.com/superp00t/gophercraft/packet"
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
	Living() bool
	Position() update.Position
	Speeds() update.Speeds
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

func (ws *WorldServer) BuildObjectUpdate(mask update.ValueMask, data *update.Data, forceCompress ...bool) (*packet.WorldPacket, error) {
	// serialize attributes according to mask
	encoded, err := update.Marshal(ws.Config.Version, mask, data)
	if err != nil {
		return nil, err
	}

	// initialize uncompressed packet, sending this is sometimes more efficient than enabling zlib compression
	uPacket := packet.NewWorldPacket(packet.SMSG_UPDATE_OBJECT)
	uPacket.Write(encoded)

	// detect if compression has been forcefully disabled/enabled
	compressionEnabled := false
	if len(forceCompress) > 0 {
		compressionEnabled = forceCompress[0]
	} else {
		// compression is enabled if the packet is over 512 bytes.
		compressionEnabled = uPacket.Len() > 512
	}

	if compressionEnabled {
		cPacket := packet.NewWorldPacket(packet.SMSG_COMPRESSED_UPDATE_OBJECT)
		compressedData := packet.Compress(uPacket.Bytes())
		cPacket.WriteUint32(uint32(uPacket.Len()))
		cPacket.Write(compressedData)
		return cPacket, nil
	}

	return uPacket, nil
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
	delete(m.Objects, id)
	m.Unlock()

	for _, v := range m.NearSet(obj) {
		v.SendObjectDelete(id)
	}
}

type WorldObjectSet []WorldObject

func (m *Map) VisibilityDistance() float32 {
	return m.Config().Float32("world.maxVisibilityRange")
}

func (m *Map) NearObjects(nearTo WorldObject) WorldObjectSet {
	return m.NearObjectsLimit(nearTo, m.VisibilityDistance())
}

func (m *Map) NearObjectsLimit(nearTo WorldObject, limit float32) WorldObjectSet {
	m.Lock()
	var wo []WorldObject
	for _, obj := range m.Objects {
		if nearTo.GUID() != obj.GUID() {
			if nearTo.Position().Dist2D(obj.Position().Point3) <= limit {
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
				if nearTo.Position().Dist2D(plyr.Position().Point3) <= limit {
					s = append(s, session)
				}
			}
		}
	}
	m.Unlock()
	return s
}

func (s *Session) SendUpdateData(mask update.ValueMask, udata *update.Data) {
	packet, err := s.WS.BuildObjectUpdate(mask, udata)
	if err != nil {
		panic(err)
	}

	s.SendAsync(packet)
}

func (s *Session) SendMovementUpdate(wo WorldObject) {
	mData := &update.MovementBlock{}

	if wo.Living() {
		mData.UpdateFlags |= update.UpdateFlagLiving
		mData.UpdateFlags |= update.UpdateFlagHasPosition
		mData.Speeds = wo.Speeds()
		s.Warnf("%s", spew.Sdump(mData.Speeds))

		if wo.TypeID() == guid.TypePlayer {
			mData.UpdateFlags |= update.UpdateFlagAll
			mData.All = 0x1
		}

		if wo.GUID() == s.GUID() {
			mData.UpdateFlags |= update.UpdateFlagSelf
		}

		mData.Info = &update.MovementInfo{
			Flags:    0,
			Time:     packet.GetMSTime(),
			Position: wo.Position(),
		}
	} else {
		mData.UpdateFlags |= update.UpdateFlagHasPosition
		mData.Position = wo.Position()
	}

	s.SendUpdateData(0, &update.Data{
		Blocks: []update.Block{
			{wo.GUID(), mData},
		},
	})
}

func (m *Map) UpdateMovement(wo WorldObject) {
	if sess, ok := wo.(*Session); ok {
		sess.SendMovementUpdate(wo)
	}

	for _, v := range m.NearSet(wo) {
		v.SendMovementUpdate(wo)
	}
}

// buildCreate creates a an SMSG_UPDATE_OBJECT structure filled out with necessary data for loading an Object into the game.
func buildCreate(obj Object, self bool) *update.Data {
	opcode := update.CreateObject
	switch obj.TypeID() {
	case guid.TypePlayer, guid.TypeUnit:
		opcode = update.SpawnObject
	}

	mData := &update.MovementBlock{
		UpdateFlags: 0,
	}

	// WorldObjects have
	if wo, ok := obj.(WorldObject); ok {
		if wo.Living() {
			mData.UpdateFlags |= update.UpdateFlagLiving
			mData.UpdateFlags |= update.UpdateFlagHasPosition
			mData.Speeds = wo.Speeds()

			if wo.TypeID() == guid.TypePlayer {
				mData.UpdateFlags |= update.UpdateFlagAll
				mData.All = 0x1
			}

			if self {
				mData.UpdateFlags |= update.UpdateFlagSelf
			}

			mData.Info = &update.MovementInfo{
				Flags:    0,
				Time:     packet.GetMSTime(),
				Position: wo.Position(),
			}
		} else {
			mData.UpdateFlags |= update.UpdateFlagHasPosition
			mData.Position = wo.Position()
		}
	}

	sp := &update.CreateBlock{
		opcode,
		obj.TypeID(),
		mData,
		obj.Values(),
	}

	packet := &update.Data{
		Blocks: []update.Block{update.Block{obj.GUID(), sp}},
	}

	return packet
}

func (s *Session) SendObjectCreate(wo Object) {
	name, _ := s.WS.GetUnitNameByGUID(wo.GUID())

	if s.objectDebug {
		s.Warnf("Sending create of %s (%s)", wo.GUID(), name)

		sass, ok := wo.(*Session)
		if ok {
			s.Warnf("Map: %d", sass.CurrentMap)
		}

		wobj, ok := wo.(WorldObject)
		if ok {
			s.Warnf("Position: %+v", wobj.Position())
		}
	}

	self := s.GUID() == wo.GUID()
	uData := buildCreate(wo, self)
	relMask := s.WS.queryRelationshipMask(wo.GUID(), s.GUID())

	s.SendUpdateData(update.ValuesCreate|relMask, uData)
}

func (s *Session) SendObjectDelete(g guid.GUID) {
	pkt := packet.NewWorldPacket(packet.SMSG_DESTROY_OBJECT)
	g.EncodeUnpacked(s.Version(), pkt)
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

func (ws *WorldServer) queryRelationshipMask(src, target guid.GUID) update.ValueMask {
	if src == target {
		return update.ValuesPrivate
	}

	// todo: determine party relationship

	return update.ValuesNone
}

// todo: handle byte field updates
func (m *Map) ModifyObject(id guid.GUID, changes map[update.Global]interface{}) {
	yo.Ok("Locking map...")
	m.Lock()
	yo.Ok("acquired map...")
	o := m.Objects[id]
	m.Unlock()

	if o == nil {
		yo.Warn("modify request for unknown object", id)
		return
	}
	valuesStore := o.Values()

	// store changes
	valuesStore.ModifyAndLock(changes)

	uo := &update.Data{
		Blocks: []update.Block{
			update.Block{id, valuesStore},
		},
	}

	// if the Object is a player session
	if session, ok := o.(*Session); ok {
		session.SendUpdateData(update.ValuesPrivate|update.ValuesParty, uo)
	}

	// transmit appropriate changes.
	for _, v := range m.NearSet(o) {
		v.SendUpdateData(m.Phase.Server.queryRelationshipMask(o.GUID(), v.GUID()), uo)
	}

	valuesStore.ClearChangesAndUnlock()
}

func (m *Map) PropagateChanges(id guid.GUID) {
	m.Lock()
	o := m.Objects[id]
	m.Unlock()

	valuesStore := o.Values()

	uo := &update.Data{
		Blocks: []update.Block{
			update.Block{id, valuesStore},
		},
	}

	// if the Object is a player session
	if session, ok := o.(*Session); ok {
		session.SendUpdateData(update.ValuesPrivate|update.ValuesParty, uo)
	}

	// transmit appropriate changes.
	for _, v := range m.NearSet(o) {
		v.SendUpdateData(m.Phase.Server.queryRelationshipMask(o.GUID(), v.GUID()), uo)
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
