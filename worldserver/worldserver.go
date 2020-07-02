package worldserver

import (
	"context"
	"crypto/rand"
	"crypto/tls"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/superp00t/gophercraft/gcore/sys"
	"github.com/superp00t/gophercraft/vsn"

	"github.com/superp00t/etc/yo"
	"github.com/superp00t/gophercraft/datapack"
	"github.com/superp00t/gophercraft/gcore/config"
	"github.com/superp00t/gophercraft/guid"
	"github.com/superp00t/gophercraft/packet"
	"github.com/superp00t/gophercraft/worldserver/script"
	"github.com/superp00t/gophercraft/worldserver/wdb"
)

type WorldServer struct {
	Config            *config.World
	DB                *wdb.Core
	PhaseL            sync.Mutex
	Phases            map[string]*Phase
	PlayersL          sync.Mutex
	PlayerList        map[string]*Session
	PackLoader        *datapack.Loader
	ScriptEngine      *script.Engine
	scriptFunc        chan func() error
	eventMgr          sync.Map
	handlers          *Handlers
	AuthServiceClient sys.AuthServiceClient
	tlsConfig         *tls.Config
	gameObjectCounter uint64
	StartTime         time.Time
}

func Start(opts *config.World) error {
	ws := &WorldServer{}
	ws.Config = opts
	ws.Phases = make(map[string]*Phase)
	ws.PlayerList = make(map[string]*Session)
	ws.scriptFunc = make(chan func() error)
	ws.gameObjectCounter = 1
	ws.StartTime = time.Now()
	core, err := wdb.NewCore(opts.DBDriver, opts.DBURL)
	if err != nil {
		return err
	}

	ws.DB = core

	if opts.ShowSQL {
		ws.DB.ShowSQL(true)
	}

	ws.PackLoader, err = datapack.Open(ws.Config.DatapackDir)
	if err != nil {
		return err
	}

	if err := ws.LoadDatapacks(); err != nil {
		return err
	}

	ws.initHandlers()

	go ws.connectRPC()

	if opts.Version == vsn.Alpha {
		go ws.serveRedirect()
	}

	characterCt, err := ws.DB.Count(new(wdb.Character))
	if err != nil {
		return err
	}

	yo.Println("Gophercraft Core World Server database opened without issue.")
	yo.Println(characterCt, "characters exist on this realm.")

	l, err := net.Listen("tcp", opts.Listen)
	if err != nil {
		return err
	}

	yo.Ok("Worldserver started to listen at", opts.Listen)

	for {
		c, err := l.Accept()
		if err != nil {
			continue
		}

		go ws.Handle(c)
	}
}

func (ws *WorldServer) Build() vsn.Build {
	return vsn.Build(ws.Config.Version)
}

func (ws *WorldServer) RealmID() uint64 {
	return ws.Config.RealmID
}

func (ws *WorldServer) Handle(c net.Conn) {
	if ws.AuthServiceClient == nil {
		panic("no auth service")
	}

	yo.Println("["+ws.Config.RealmName+"] New worldserver connection from", c.RemoteAddr().String())

	// Authentication
	salt := randomSalt(4)

	smsg := &packet.SMSGAuthPacket{
		Salt:  salt,
		Seed1: randomSalt(16),
		Seed2: randomSalt(16),
	}

	dat := smsg.Encode(ws.Config.Version)
	c.Write(dat)

	if ws.Config.Version.RemovedIn(vsn.V1_12_1) {
		ws.handleAlpha(c)
		return
	}

	buf := make([]byte, 512)
	wr, err := c.Read(buf)
	if err != nil {
		yo.Println(err)
		c.Close()
		return
	}

	yo.Spew(buf)

	cmsg, err := packet.UnmarshalCMSGAuthSession(buf[:wr])
	if err != nil {
		yo.Println("Invalid protocol: ", err)
		c.Close()
		return
	}

	yo.Spew(cmsg)

	if cmsg.Build != ws.Config.Version {
		yo.Warn("Client attempted to join with invalid build", cmsg.Build)
		c.Close()
		return
	}

	yo.Ok("Accepted connection with version", cmsg.Build)

	// Invoke the GRPC server.
	// This connects back to gcraft_core_auth, checking to see if this client is actually a registered user.
	resp, err := ws.AuthServiceClient.VerifyWorld(context.Background(), &sys.VerifyWorldQuery{
		RealmID:     ws.Config.RealmID,
		Build:       uint32(cmsg.Build),
		Account:     cmsg.Account,
		GameAccount: "Zero",
		IP:          c.RemoteAddr().String(),
		Digest:      cmsg.Digest,
		Salt:        salt,
		Seed:        cmsg.Seed,
	})

	if err != nil {
		yo.Warn(err)
		loginFail(c)
		return
	}

	switch resp.Status {
	case sys.Status_SysOK:
	default:
		yo.Warn("Login for user", cmsg.Account, "failed")
		loginFail(c)
		return
	}

	crypt := packet.NewConnection(cmsg.Build, c, resp.SessionKey, true)
	session := &Session{
		GameAccount:   resp.GameAccount,
		AddonData:     cmsg.AddonData,
		SessionKey:    resp.SessionKey,
		WS:            ws,
		Connection:    crypt,
		Tier:          resp.Tier,
		messageBroker: make(chan *packet.WorldPacket, 64),
	}

	// TODO: Position user in queue if server is over capacity

	response := packet.NewWorldPacket(packet.SMSG_AUTH_RESPONSE)
	response.WriteByte(packet.AUTH_OK)
	response.WriteUint64(0)
	response.WriteByte(0)
	switch ws.Config.Version {
	case 5875:
	case 12340:
		response.Write([]byte{2}) // expansion number 2
	}

	session.SendSync(response)

	yo.Println("Auth response sent,")

	if session.WS.Config.WardenEnabled {
		session.InitWarden()
	}

	go func() {
		for {
			data, ok := <-session.messageBroker
			if !ok {
				return
			}

			if err := session.SendSync(data); err != nil {
				yo.Warn(err)
				return
			}
		}
	}()

	session.IntroductoryPackets()
	session.State = CharacterSelectMenu
	session.Handle()
}

func randomSalt(le int) []byte {
	salt := make([]byte, le)
	rand.Read(salt)
	return salt
}

func loginFail(c net.Conn) {
	yo.Println("Login failure")
	wp := packet.NewWorldPacket(packet.SMSG_AUTH_RESPONSE)
	wp.Write([]byte{packet.AUTH_REJECT})
	c.Write(wp.Finish())
	time.Sleep(1 * time.Second)
	c.Close()
	return
}

func hash(input ...[]byte) []byte {
	return packet.Hash(input...)
}

func (s *WorldServer) GetSessionByGUID(g guid.GUID) (*Session, error) {
	s.PlayersL.Lock()
	defer s.PlayersL.Unlock()
	for _, v := range s.PlayerList {
		if v.State == InWorld {
			if v.GUID() == g {
				return v, nil
			}
		}
	}

	return nil, fmt.Errorf("could not find session corresponding to input")
}

func (ws *WorldServer) GetPlayerNameByGUID(g guid.GUID) (string, error) {
	plyr, err := ws.GetSessionByGUID(g)
	if err == nil {
		return plyr.PlayerName(), nil
	}

	var chr wdb.Character

	found, err := ws.DB.Where("id = ?", g.Counter()).Get(&chr)
	if err != nil {
		return "", err
	}

	if !found {
		return "", fmt.Errorf("no character found for guid %s", g)
	}

	return chr.Name, nil
}

func (s *WorldServer) GetUnitNameByGUID(g guid.GUID) (string, error) {
	if g == guid.Nil {
		return "", nil
	}

	switch g.HighType() {
	case guid.Player:
		plyr, err := s.GetSessionByGUID(g)
		if err != nil {
			return "", err
		}
		return plyr.PlayerName(), nil
	case guid.Creature:
		return "", fmt.Errorf("npc names nyi")
	default:
		return "", fmt.Errorf("cannot name this type (%s)", g.HighType())
	}
}

func (s *WorldServer) GetSessionByPlayerName(playerName string) (*Session, error) {
	s.PlayersL.Lock()
	session := s.PlayerList[playerName]
	s.PlayersL.Unlock()
	if session != nil {
		return session, nil
	}

	return nil, fmt.Errorf("no session for player '%s'", playerName)
}

func (s *WorldServer) GetGUIDByPlayerName(playerName string) (guid.GUID, error) {
	s.PlayersL.Lock()
	session := s.PlayerList[playerName]
	s.PlayersL.Unlock()
	if session != nil {
		return session.GUID(), nil
	}

	var chr wdb.Character
	found, err := s.DB.Where("name = ?", playerName).Get(&chr)
	if err != nil {
		return guid.Nil, err
	}

	if !found {
		return guid.Nil, fmt.Errorf("no player by the name of %s", playerName)
	}

	return guid.RealmSpecific(guid.Player, s.RealmID(), chr.ID), nil
}

func (s *Session) SendAuthWaitQueue(position uint32) {
	p := packet.NewWorldPacket(packet.SMSG_AUTH_RESPONSE)
	p.WriteByte(packet.AUTH_WAIT_QUEUE)
	p.WriteUint32(position)
	p.WriteByte(0)
	s.SendSync(p)
}
