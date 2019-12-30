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

	"github.com/superp00t/etc/yo"
	"github.com/superp00t/gophercraft/datapack"
	"github.com/superp00t/gophercraft/format/dbc"
	"github.com/superp00t/gophercraft/gcore/config"
	"github.com/superp00t/gophercraft/guid"
	"github.com/superp00t/gophercraft/packet"
	"github.com/superp00t/gophercraft/worldserver/wdb"
)

type WorldServer struct {
	RealmID           uint64
	Config            *config.World
	DB                *wdb.Core
	PhaseL            sync.Mutex
	Phases            map[uint32]*Phase
	PlayersL          sync.Mutex
	PlayerList        map[string]*Session
	PackLoader        *datapack.Loader
	handlers          *Handlers
	AuthServiceClient sys.AuthServiceClient
	tlsConfig         *tls.Config
}

func Start(opts *config.World) error {
	ws := &WorldServer{}
	ws.Config = opts
	ws.Phases = make(map[uint32]*Phase)
	ws.PlayerList = make(map[string]*Session)
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

	for {
		c, err := l.Accept()
		if err != nil {
			continue
		}

		go ws.Handle(c)
	}
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

	dat := smsg.Encode(uint16(ws.Config.Version))
	c.Write(dat)

	buf := make([]byte, 512)
	wr, err := c.Read(buf)
	if err != nil {
		yo.Println(err)
		c.Close()
		return
	}

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

	resp, err := ws.AuthServiceClient.VerifyWorld(context.Background(), &sys.VerifyWorldQuery{
		RealmID:     ws.Config.RealmID,
		Build:       cmsg.Build,
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

	crypt := packet.NewCrypter(cmsg.Build, c, resp.SessionKey, true)
	session := &Session{
		GameAccount: resp.GameAccount,
		AddonData:   cmsg.AddonData,
		SessionKey:  resp.SessionKey,
		WS:          ws,
		C:           c,
		Crypter:     crypt,
		Tier:        resp.Tier,
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

type Char struct {
	Race   packet.Race
	Gender uint8
}

func (s *WorldServer) GetNative(race packet.Race, gender uint8) uint32 {
	var races dbc.Ent_ChrRaces
	found, _ := s.DB.Where("id = ?", race).Get(&races)
	if !found {
		return 2838
	}

	if gender == 1 {
		return races.FemaleDisplayID
	}

	return races.MaleDisplayID
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
		return "", fmt.Errorf("cannot name this type")
	}
}

func (s *Session) SendAuthWaitQueue(position uint32) {
	p := packet.NewWorldPacket(packet.SMSG_AUTH_RESPONSE)
	p.WriteByte(packet.AUTH_WAIT_QUEUE)
	p.WriteUint32(position)
	p.WriteByte(0)
}
