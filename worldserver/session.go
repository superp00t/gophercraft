package worldserver

import (
	"fmt"
	"io"
	"net"
	"strings"
	"sync"

	"github.com/superp00t/gophercraft/packet/update"
	"github.com/superp00t/gophercraft/worldserver/wdb"

	"github.com/superp00t/etc"
	"github.com/superp00t/etc/yo"
	"github.com/superp00t/gophercraft/guid"
	"github.com/superp00t/gophercraft/warden"

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
	Tier        sys.Tier
	State       SessionState
	GameAccount uint64
	SessionKey  []byte
	AddonData   []byte
	Warden      *warden.Warden
	WS          *WorldServer
	C           net.Conn
	Crypter     *packet.Crypter
	Char        *wdb.Character
	lInventory  sync.Mutex
	Inventory   map[guid.GUID]*Item
	summons     *summons

	// In-world data
	CurrentPhase string
	CurrentMap   uint32
	ZoneID       uint32
	// currently tracked objects
	lTrackedGUIDs sync.Mutex
	TrackedGUIDs  []guid.GUID

	*update.ValuesBlock

	PlayerSpeeds   update.Speeds
	SpeedModifier  float32
	PlayerPosition update.Position

	messageBroker chan *packet.WorldPacket
	brokerClosed  bool
	objectDebug   bool
}

func (s *Session) Living() bool {
	return true
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

func (s *Session) Position() update.Position {
	return s.PlayerPosition
}

func (s *Session) Speeds() update.Speeds {
	speeds := s.PlayerSpeeds
	for k := range speeds {
		speeds[k] = speeds[k] * s.SpeedModifier
	}

	return speeds
}

func (s *Session) ReadCrypt() (packet.WorldType, []byte, error) {
	frame, err := s.Crypter.ReadFrame()
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
		yo.Warn("Broker is clossed")
	}
}

func (s *Session) oldGUID() bool {
	return s.Version() <= 20000
}

func (s *Session) decodeUnpackedGUID(in io.Reader) guid.GUID {
	g, err := guid.DecodeUnpacked(s.Version(), in)
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
	g, err := guid.DecodePacked(s.Version(), in)
	if err != nil {
		yo.Warn(err)
		return guid.Nil
	}

	return s.convertClientGUID(g)
}

func (s *Session) SendSync(p *packet.WorldPacket) error {
	return s.Crypter.SendFrame(packet.Frame{
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

	gid, err := guid.DecodeUnpacked(s.Version(), e)
	if err != nil {
		panic(err)
	}

	yo.Println("Player join requested")

	var cl []wdb.Character

	s.DB().Find(&cl)

	for _, v := range cl {
		if v.ID == gid.Counter() {
			s.Char = &v
			yo.Println("GUID found for character", v.Name, gid)
			s.SetupOnLogin()
			return
		}
	}

	// Todo handle unknown GUID
}

func (s *Session) Handle() {
	for {
		f, err := s.Crypter.ReadFrame()
		if err != nil {
			yo.Println(err)
			if s.Char != nil {
				s.WS.PlayersL.Lock()
				if pls := s.WS.PlayerList[s.PlayerName()]; pls != nil {
					delete(s.WS.PlayerList, s.PlayerName())
				}
				s.WS.PlayersL.Unlock()
			}

			if s.State == InWorld {
				s.Map().RemoveObject(s.GUID())
			}

			s.Crypter.Conn.Close()
			s.brokerClosed = true
			close(s.messageBroker)
			return
		}

		yo.Println(f.Type, "requested", len(f.Data))

		if strings.HasPrefix(f.Type.String(), "WorldType(") {
			s.Crypter.Conn.Close()
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
			default:
				panic("unusable function type for " + f.Type.String())
			}
		} else {
			yo.Warn("Unauthorized packet sent from ", s.Crypter.Conn.RemoteAddr().String())
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
	return packet.Class(s.GetByteValue(update.UnitClass))
}
