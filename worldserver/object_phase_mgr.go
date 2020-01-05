package worldserver

import (
	"sync"

	"github.com/superp00t/gophercraft/gcore/config"
	"github.com/superp00t/gophercraft/packet/update"

	"github.com/superp00t/etc/yo"
	"github.com/superp00t/gophercraft/guid"
	"github.com/superp00t/gophercraft/packet"
)

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
	Position() update.Quaternion
	Speeds() update.Speeds
}

// TODO: delete phases when players are not there.
func (ws *WorldServer) Phase(i uint32) *Phase {
	ws.PhaseL.Lock()
	ph := ws.Phases[i]
	if ph == nil {
		ph = &Phase{
			Maps:   make(map[uint32]*Map),
			Server: ws,
		}

		ws.Phases[i] = ph
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

func (m *Map) AddObject(obj WorldObject) error {
	m.Lock()
	m.Objects[obj.GUID()] = obj
	m.Unlock()

	// Send spawn packet to nearby players.
	for _, v := range m.NearbySessions(obj) {
		v.SendObjectCreate(obj)
	}

	return nil
}

func (m *Map) RemoveObject(id guid.GUID) {
	m.Lock()
	obj := m.Objects[id]
	delete(m.Objects, id)
	m.Unlock()

	for _, v := range m.NearbySessions(obj) {
		v.SendObjectDelete(id)
	}
}

// NearbySessions enumerates a list of players close to (less than or equal to world.maxVisibilityRange) a game object on a map.
func (m *Map) NearbySessions(nearTo WorldObject) []*Session {
	return m.NearbyLimit(nearTo, m.Config().Float32("world.maxVisibilityRange"))
}

func (m *Map) NearbyLimit(nearTo WorldObject, limit float32) []*Session {
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
	packet := &update.Data{
		Blocks: []update.Block{
			update.Block{g, &update.DeleteObjectsBlock{
				update.DeleteFarObjects,
				[]guid.GUID{g},
			}},
		},
	}

	s.SendUpdateData(update.ValuesNone, packet)
}

func (s *Session) SendAreaAll(p *packet.WorldPacket) {
	s.SendAsync(p)

	for _, v := range s.WS.Phase(s.CurrentPhase).Map(s.CurrentMap).NearbySessions(s) {
		v.SendAsync(p)
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
	for _, v := range m.NearbySessions(o) {
		v.SendUpdateData(m.Phase.Server.queryRelationshipMask(o.GUID(), v.GUID()), uo)
	}

	valuesStore.ClearChangesAndUnlock()
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
