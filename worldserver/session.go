package worldserver

import (
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/superp00t/gophercraft/packet/update"
	"github.com/superp00t/gophercraft/worldserver/wdb"

	"github.com/superp00t/etc"
	"github.com/superp00t/etc/yo"
	"github.com/superp00t/gophercraft/guid"
	"github.com/superp00t/gophercraft/warden"

	"github.com/superp00t/gophercraft/gcore/config"
	"github.com/superp00t/gophercraft/gcore/sys"
	"github.com/superp00t/gophercraft/packet"
)

type SessionState uint8

const (
	CharacterSelectMenu SessionState = iota
	InWorld
)

type Session struct {
	// Account data
	WS          *WorldServer
	Connection  *packet.Connection
	State       SessionState
	Warden      *warden.Warden
	Tier        sys.Tier
	GameAccount uint64
	SessionKey  []byte
	AddonData   []byte
	Char        *wdb.Character
	lInventory  sync.Mutex
	Inventory   map[guid.GUID]*Item

	// In-world data
	CurrentPhase string
	CurrentMap   uint32
	ZoneID       uint32
	Group        *Group
	GroupInvite  guid.GUID
	summons      *summons
	// currently tracked objects
	lTrackedGUIDs sync.Mutex
	TrackedGUIDs  []guid.GUID

	*update.ValuesBlock

	MoveSpeeds   update.Speeds
	MovementInfo *update.MovementInfo

	messageBroker chan *packet.WorldPacket
	brokerClosed  bool
	objectDebug   bool
}

func (s *Session) TypeID() guid.TypeID {
	// activeplayer
	return guid.TypePlayer
}

func (s *Session) GUID() guid.GUID {
	return guid.RealmSpecific(guid.Player, s.WS.RealmID(), s.Char.ID)
}

func (s *Session) Values() *update.ValuesBlock {
	return s.ValuesBlock
}

func (s *Session) Movement() *update.MovementBlock {
	mData := &update.MovementBlock{
		Speeds:   s.MoveSpeeds,
		Position: s.MovementInfo.Position,
		Info:     s.MovementInfo,
	}

	mData.UpdateFlags |= update.UpdateFlagLiving
	mData.UpdateFlags |= update.UpdateFlagHasPosition

	mData.UpdateFlags |= update.UpdateFlagAll
	mData.All = 0x1
	return mData
}

func (s *Session) Position() update.Position {
	return s.MovementInfo.Position
}

func (s *Session) Speeds() update.Speeds {
	return s.MoveSpeeds
}

func (s *Session) ReadCrypt() (packet.WorldType, []byte, error) {
	frame, err := s.Connection.ReadFrame()
	if err != nil {
		return 0, nil, err
	}

	return frame.Type, frame.Data, nil
}

// todo: make more consistent
func (s *Session) SendAsync(p *packet.WorldPacket) {
	if !s.brokerClosed {
		s.messageBroker <- p
	} else {
		yo.Warn("Broker is closed")
	}
}

func (s *Session) oldGUID() bool {
	// Patch 6.0.2
	return s.Build() < 19027
}

func (s *Session) decodeUnpackedGUID(in io.Reader) guid.GUID {
	g, err := guid.DecodeUnpacked(s.Build(), in)
	if err != nil {
		return guid.Nil
	}

	return s.convertClientGUID(g)
}

func (s *Session) convertClientGUID(g guid.GUID) guid.GUID {
	// The realm ID isn't present in older versions. We still have to add it in so the GUIDs are equal server side.
	if s.oldGUID() && g != guid.Nil {
		g = g.SetRealmID(s.WS.RealmID())
	}

	if g.RealmID() == 0 && g.Counter() == 0 && g.HighType() == guid.Player {
		g = guid.Nil
	}

	return g
}

func (s *Session) decodePackedGUID(in io.Reader) guid.GUID {
	g, err := guid.DecodePacked(s.Build(), in)
	if err != nil {
		yo.Warn(err)
		return guid.Nil
	}

	return s.convertClientGUID(g)
}

func (s *Session) SendSync(p *packet.WorldPacket) error {
	return s.Connection.SendFrame(packet.Frame{
		Type: p.Type,
		Data: p.Finish(),
	})
}

func (s *Session) HandlePong(e *etc.Buffer) {
	ping := e.ReadUint32()
	latency := e.ReadUint32()
	yo.Println("Ping: ", ping, "Latency", latency)
	pkt := packet.NewWorldPacket(packet.SMSG_PONG)
	pkt.WriteUint32(ping)
	s.SendAsync(pkt)
}

func (s *Session) DB() *wdb.Core {
	return s.WS.DB
}

func (s *Session) HandleJoin(e *etc.Buffer) {
	if s.State == InWorld {
		return
	}

	// todo: handle player already in world

	gid := s.decodeUnpackedGUID(e)
	if gid == guid.Nil {
		return
	}

	yo.Println("Player join requested", gid)

	if sess, _ := s.WS.GetSessionByGUID(gid); sess != nil {
		s.SendLoginFailure(packet.CharLoginDuplicateCharacter)
		return
	}

	var chr wdb.Character

	found, err := s.DB().Where("game_account = ?", s.GameAccount).Where("id = ?", gid.Counter()).Get(&chr)
	if err != nil {
		panic(err)
	}

	if found {
		s.Char = &chr
		yo.Println("GUID found for character", chr.Name, gid)
		s.SetupOnLogin()
		return
	}

	// Todo handle unknown GUID
	s.SendLoginFailure(packet.CharLoginNoCharacter)
}

func (s *Session) Cleanup() {
	if s.Char != nil {
		s.CleanupPlayer()
	}

	s.Connection.Conn.Close()
	s.brokerClosed = true
	close(s.messageBroker)
}

func (s *Session) Handle() {
	for {
		f, err := s.Connection.ReadFrame()
		if err != nil {
			yo.Println(err)
			s.Cleanup()
			return
		}

		yo.Println(f.Type, "requested", len(f.Data))

		if strings.HasPrefix(f.Type.String(), "WorldType(") {
			s.Connection.Conn.Close()
			s.Cleanup()
			continue
		}

		h, ok := s.WS.handlers.Map[f.Type]
		if !ok {
			continue
		}

		if h.RequiredState <= s.State {
			switch fn := h.Fn.(type) {
			case func(*Session, []byte):
				fn(s, f.Data)
			case func(*Session, packet.WorldType, []byte):
				fn(s, f.Type, f.Data)
			case func(*Session, *etc.Buffer):
				fn(s, etc.FromBytes(f.Data))
			case func(*Session):
				fn(s)
			default:
				panic("unusable function type for " + f.Type.String())
			}
		} else {
			yo.Warn("Unauthorized packet sent from ", s.Connection.Conn.RemoteAddr().String())
		}
	}
}

func PowerType(class packet.Class) uint8 {
	// switch class {
	// 	case packet.CLASS_
	// }
	return packet.MANA
}

func (s *Session) SendPet(b []byte) {
	pkt := packet.NewWorldPacket(packet.SMSG_PET_NAME_QUERY_RESPONSE)
	pkt.WriteUint32(0)
	pkt.WriteUint64(0)
	s.SendAsync(pkt)
	yo.Ok("Sent pet response")
}

func (s *Session) Alertf(format string, args ...interface{}) {
	s.SendAlertText(fmt.Sprintf(format, args...))
}

func (s *Session) SendAlertText(data string) {
	pkt := packet.NewWorldPacket(packet.SMSG_AREA_TRIGGER_MESSAGE)
	pkt.WriteUint32(uint32(len(data) + 1))
	pkt.WriteCString(data)
	s.SendAsync(pkt)
}

func (s *Session) Map() *Map {
	return s.WS.Phase(s.CurrentPhase).Map(s.CurrentMap)
}

func (s *Session) GetPlayerClass() packet.Class {
	return packet.Class(s.GetByte("Class"))
}

func (s *Session) Config() *config.World {
	return s.WS.Config
}
