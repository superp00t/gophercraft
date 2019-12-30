package client

import (
	"bytes"
	"crypto/rand"
	"crypto/sha1"
	"fmt"
	"net"

	"github.com/davecgh/go-spew/spew"
	"github.com/superp00t/gophercraft/auth"
	"github.com/superp00t/gophercraft/guid"
	"github.com/superp00t/gophercraft/packet"
	"github.com/superp00t/gophercraft/warden"
)

type Config struct {
	Version            uint32
	Username, Password string
	Player             string
	Realmlist          string
}

type Client struct {
	Auth       *auth.Client
	Player     string
	PlayerGUID guid.GUID
	Config     *Config
	RealmList  *packet.RealmList_S
	Warden     *warden.Warden
	Crypter    *packet.Crypter
	SessionKey []byte
	Handlers   map[packet.WorldType]*ClientHandler
}

func New(cfg Config) (*Client, error) {
	c := &Client{}
	c.Player = cfg.Playername
	c.Config = cfg
	var err error
	c.Auth, err = auth.Login(c.Config.Version, c.Config.Realmlist, c.Config.Username, c.Config.Password)
	if err != nil {
		return nil, err
	}

	c.RealmList, err = c.Auth.GetRealmlist()
	if err != nil {
		return nil, err
	}

	return c, nil
}

func (cl *Client) Connect(ip string) error {
	conn, err := net.Dial("tcp", ip)
	if err != nil {
		return err
	}

	buf := make([]byte, 512)
	if _, err := conn.Read(buf); err != nil {
		return err
	}

	gp, err := packet.UnmarshalSMSGAuthPacket(cl.Cfg.Build, buf)
	if err != nil {
		return err
	}

	seed := randomBuffer(4)
	h := hash(
		[]byte(cl.Cfg.Username),
		[]byte{0, 0, 0, 0},
		seed,
		gp.Salt,
		cl.Auth.SessionKey,
	)

	app := &packet.CMSGAuthSession{
		Build:     uint32(cl.Cfg.Build),
		Account:   cl.Cfg.Username,
		Seed:      seed,
		Digest:    h,
		AddonData: packet.ClientAddonData,
	}

	if _, err = wc.Write(app.Encode()); err != nil {
		return nil
	}

	cl.Handlers = make(map[packet.WorldType]*ClientHandler)
	cl.World = wc
	cl.Crypter = packet.NewCrypter(cl.Cfg.Build, wc, cl.SessionKey, false)

	return cl.Handle()
}

func hash(input ...[]byte) []byte {
	bt := sha1.Sum(bytes.Join(input, nil))
	return bt[:]
}

func randomBuffer(l int) []byte {
	b := make([]byte, l)
	rand.Read(b)
	return b
}
