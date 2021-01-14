package realm

import (
	"context"
	"crypto/rand"
	"crypto/tls"
	"fmt"
	"net"
	"os"
	"os/signal"
	"runtime"
	"runtime/pprof"
	"sync"
	"time"

	"github.com/superp00t/gophercraft/gcore/sys"
	"github.com/superp00t/gophercraft/i18n"
	"github.com/superp00t/gophercraft/vsn"

	"github.com/superp00t/etc/yo"
	"github.com/superp00t/gophercraft/datapack"
	"github.com/superp00t/gophercraft/gcore/config"
	"github.com/superp00t/gophercraft/guid"
	"github.com/superp00t/gophercraft/packet"
	"github.com/superp00t/gophercraft/realm/wdb"
)

type Server struct {
	Config     *config.World
	DB         *wdb.Core
	PhaseL     sync.Mutex
	Phases     map[string]*Phase
	PlayersL   sync.Mutex
	PlayerList map[string]*Session
	PackLoader *datapack.Loader
	Plugins    []*LoadedPlugin

	eventMgr          sync.Map
	handlers          *Handlers
	CommandHandlers   []Command
	AuthServiceClient sys.AuthServiceClient
	tlsConfig         *tls.Config
	DynamicCounters   map[guid.TypeID]uint64
	GuardCounters     sync.Mutex
	StartTime         time.Time
	TerrainMgr
	// Misc data stores
	LevelExperience           wdb.LevelExperience
	PlayerCreateInfo          []wdb.PlayerCreateInfo
	PlayerCreateItems         []wdb.PlayerCreateItem
	PlayerCreateActionButtons []wdb.PlayerCreateActionButton
	PlayerCreateAbilities     []wdb.PlayerCreateAbility
}

func Start(opts *config.World) error {
	if opts.CPUProfile != "" {
		f, err := os.Create(opts.CPUProfile)
		if err != nil {
			return err
		}

		pprof.StartCPUProfile(f)
	}

	// Graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		yo.Warn("Stopping realm server")
		if opts.CPUProfile != "" {
			yo.Warn("Closing CPU profile")
			pprof.StopCPUProfile()
		}
		os.Exit(0)
	}()

	ws := &Server{}
	ws.Config = opts
	ws.Phases = make(map[string]*Phase)
	ws.PlayerList = make(map[string]*Session)
	ws.DynamicCounters = make(map[guid.TypeID]uint64)
	ws.StartTime = time.Now()
	// Open database
	core, err := wdb.NewCore(opts.DBDriver, opts.DBURL)
	if err != nil {
		return err
	}

	ws.DB = core

	if opts.ShowSQL {
		ws.DB.ShowSQL(true)
	}

	if err := ws.loadPlugins(); err != nil {
		return err
	}

	// Open handles to ZIP archives and indices of flat folders.
	ws.PackLoader, err = datapack.Open(ws.Config.DatapackDir)
	if err != nil {
		return err
	}

	if err := ws.LoadDatapacks(); err != nil {
		return err
	}

	ws.initHandlers()

	// Remove unused chunks
	go ws.InitTerrainMgr()

	go ws.connectRPC()

	if opts.Version == vsn.Alpha {
		go ws.serveRedirect()
	}

	characterCt, err := ws.DB.Count(new(wdb.Character))
	if err != nil {
		return err
	}

	yo.Println("Gophercraft Core World Server successfully initialized!")
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

func (ws *Server) UptimeMS() uint32 {
	return uint32(time.Since(ws.StartTime) / time.Millisecond)
}

func (ws *Server) Build() vsn.Build {
	return vsn.Build(ws.Config.Version)
}

func (ws *Server) RealmID() uint64 {
	return ws.Config.RealmID
}

func (ws *Server) Handle(c net.Conn) {
	if ws.AuthServiceClient == nil {
		panic("no auth service")
	}

	yo.Println("["+ws.Config.RealmName+"] New worldserver connection from", c.RemoteAddr().String())

	if ws.Config.Version.AddedIn(vsn.V8_3_0) {
		yo.Ok("Modern protocol selected.")
		ws.handleModern(c)
		return
	}

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

	cmsg, err := packet.UnmarshalCMSGAuthSession(ws.Config.Version, buf[:wr])
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

	var aInfo *packet.AddonList

	if len(cmsg.AddonData) > 0 {
		var err error
		aInfo, err = packet.ParseAddonList(cmsg.Build, cmsg.AddonData)
		if err != nil {
			yo.Warn(err)
			loginFail(c)
			return
		}
		yo.Spew(aInfo)
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
		AddonData:     aInfo,
		SessionKey:    resp.SessionKey,
		WS:            ws,
		Connection:    crypt,
		Tier:          resp.Tier,
		Locale:        i18n.Locale(resp.Locale),
		messageBroker: make(chan *packet.WorldPacket, 64),
	}

	if session.AddonData != nil {
		aInfoPacket := packet.BuildServerAddonResponse(session.Build(), session.AddonData)
		session.SendSync(aInfoPacket)
		yo.Println("Addon info sent")
	}
	// TODO: Position user in queue if server is over capacity

	response := packet.NewWorldPacket(packet.SMSG_AUTH_RESPONSE)
	response.WriteByte(packet.AUTH_OK)
	response.WriteUint64(0)
	response.WriteByte(0)
	switch ws.Config.Version {
	case 5875:
	case 8606:
		response.WriteByte(1)
	case 12340:
		response.WriteByte(2) // expansion number 2
	}

	session.SendSync(response)

	yo.Println("Auth response sent,")

	session.Init()
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

func (s *Server) GetSessionByGUID(g guid.GUID) (*Session, error) {
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

func (ws *Server) GetPlayerNameByGUID(g guid.GUID) (string, error) {
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

func (s *Server) GetUnitNameByGUID(g guid.GUID) (string, error) {
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

func (s *Server) GetSessionByPlayerName(playerName string) (*Session, error) {
	s.PlayersL.Lock()
	session := s.PlayerList[playerName]
	s.PlayersL.Unlock()
	if session != nil {
		return session, nil
	}

	return nil, fmt.Errorf("no session for player '%s'", playerName)
}

func (s *Server) GetGUIDByPlayerName(playerName string) (guid.GUID, error) {
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

type ServerStats struct {
	Allocated      uint64
	TotalAllocated uint64
	SystemMemory   uint64
	NumGCCycles    uint32
	Goroutines     int
	Uptime         time.Duration
}

func (ws *Server) GetServerStats() *ServerStats {
	sstats := &ServerStats{}
	sstats.Goroutines = runtime.NumGoroutine()
	sstats.Uptime = time.Since(ws.StartTime)
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	sstats.Allocated = memStats.Alloc
	sstats.TotalAllocated = memStats.TotalAlloc
	sstats.SystemMemory = memStats.Sys
	sstats.NumGCCycles = memStats.NumGC
	return sstats
}
