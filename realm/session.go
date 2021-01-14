package realm

import (
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/superp00t/gophercraft/format/terrain"
	"github.com/superp00t/gophercraft/i18n"
	"github.com/superp00t/gophercraft/packet/update"
	"github.com/superp00t/gophercraft/realm/wdb"
	"github.com/superp00t/gophercraft/vsn"

	"github.com/superp00t/etc"
	"github.com/superp00t/etc/yo"
	"github.com/superp00t/gophercraft/crypto/warden"
	"github.com/superp00t/gophercraft/guid"

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
	WS          *Server
	Connection  *packet.Connection
	State       SessionState
	Warden      *warden.Warden
	Tier        sys.Tier
	Locale      i18n.Locale
	Account     uint64
	GameAccount uint64
	SessionKey  []byte
	GuardProps  sync.Mutex
	Props       []wdb.PropID

	AddonData  *packet.AddonList
	Char       *wdb.Character
	lInventory sync.Mutex
	Inventory  map[guid.GUID]*Item

	// In-world data
	CurrentPhase      string
	CurrentMap        uint32
	CurrentArea       uint32
	CurrentChunkIndex *terrain.TileChunkLookupIndex
	ZoneID            uint32

	// currently tracked objects
	GuardTrackedGUIDs sync.Mutex
	TrackedGUIDs      []guid.GUID

	*update.ValuesBlock

	MoveSpeeds   update.Speeds
	MovementInfo *update.MovementInfo

	LocationUpdateTimer *time.Ticker
	KillTimers          chan bool

	messageBroker chan *packet.WorldPacket
	brokerClosed  bool

	// Social
	Group       *Group
	GroupInvite guid.GUID
	summons     *summons
}

func (session *Session) Init() {
	session.messageBroker = make(chan *packet.WorldPacket, 64)

	if session.Build().AddedIn(vsn.NewCryptSystem) {
		var as packet.AuthResponse
	}

	var props []wdb.AccountProp

	if err := session.DB().Where("id = ?", session.Account).Find(&props); err != nil {
		panic(err)
	}
	session.Props = make([]wdb.PropID, len(props))
	for i, prop := range props {
		session.Props[i] = prop.Prop
	}

	if session.WS.Config.WardenEnabled {
		session.InitWarden()
	}

	session.KillTimers = make(chan bool)
	session.LocationUpdateTimer = time.NewTicker(1 * time.Second)

	go func() {
		for {
			select {
			case <-session.LocationUpdateTimer.C:
				session.UpdateArea()
			case <-session.KillTimers:
				return
			}
		}
	}()

	go func() {
		for {
			data, ok := <-session.messageBroker
			if !ok {
				return
			
			}

			if err := session.SendSync(data); err != nil {
				yo.Warn(err)
				session.Connection.Close()
				session.Cleanup()
				return
			}
		}
	}()

	session.SendSessionMetadata()
	session.State = CharacterSelectMenu
	session.Handle()
}

func (s *Session) TypeID() guid.TypeID {
	// activeplayer
	return guid.TypePlayer
}

func (s *Session) GUID() guid.GUID {
	if s.Char == nil {
		return guid.Nil
	}
	return guid.RealmSpecific(guid.Player, s.WS.RealmID(), s.Char.ID)
}

func (s *Session) Values() *update.ValuesBlock {
	return s.ValuesBlock
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
	// TODO: Omit for types that not realm specific
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

	s.LocationUpdateTimer.Stop()
	s.Connection.Conn.Close()

	if !s.brokerClosed {
		s.brokerClosed = true
		close(s.messageBroker)
		s.KillTimers <- true
	}
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

		// if strings.HasPrefix(f.Type.String(), "WorldType(") {
		// 	s.Connection.Conn.Close()
		// 	s.Cleanup()
		// 	continue
		// }

		h, ok := s.WS.handlers.Map[f.Type]
		if !ok {
			continue
		}

		dispatchHandler := func() {
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

		var opts Options
		for _, opt := range h.Options {
			if err := opt(f.Type, f.Data, &opts); err != nil {
				// todo: close connection upon fatal error
				yo.Warn(err)
			}
		}

		// Most opcodes are better left in this goroutine, but spawn a new one if specified
		if opts.OptionFlags&OptionFlagAsync != 0 {
			go dispatchHandler()
		} else {
			dispatchHandler()
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

func (s *Session) HandleRealmSplit(split *etc.Buffer) {
	splitReq := split.ReadInt32() // realm ID perhaps?
	yo.Ok("User requested realm split", splitReq)

	response := packet.NewWorldPacket(packet.SMSG_REALM_SPLIT)
	response.WriteInt32(splitReq)
	response.WriteInt32(0) // split state
	response.WriteCString("01/01/01")

	s.SendAsync(response)
}

func (s *Session) HandleUITimeRequest() {
	resp := packet.NewWorldPacket(packet.SMSG_UI_TIME)
	resp.WriteUint32(uint32(time.Now().Unix()))
	s.SendAsync(resp)
}

func (s *Session) DebugGUID(dbg guid.GUID) string {
	switch dbg.HighType() {
	case guid.Player:
		plyr, err := s.WS.GetSessionByGUID(dbg)
		if err != nil {
			return dbg.String()
		}
		return dbg.String() + " (" + plyr.PlayerName() + ")"
	case guid.Item:
		it, ok := s.Inventory[dbg]
		if !ok {
			return dbg.String()
		}

		var tpl *wdb.ItemTemplate
		s.DB().GetData(it.ItemID, &tpl)
		if tpl == nil {
			return dbg.String()
		}

		return dbg.String() + " (" + tpl.Name.String() + ")"
	default:
		return dbg.String()
	}
}
