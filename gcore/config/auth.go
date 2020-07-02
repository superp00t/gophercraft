package config

import (
	"crypto/tls"
	"fmt"

	"github.com/superp00t/etc"
	"github.com/superp00t/gophercraft/datapack/text"
)

var (
	ErrAlreadyExists = fmt.Errorf("config: already exists")
)

type AuthFile struct {
	HTTPInternal         string
	HostExternal         string
	AuthListen           string
	BnetListen           string
	BnetRESTListen       string
	AlphaRealmlistListen string
	DBDriver             string
	DBURL                string
}

type RealmsFile struct {
	Realms map[uint64]Realm
}

type Realm struct {
	FP     string
	Armory string
}

type Auth struct {
	Path        etc.Path
	Certificate tls.Certificate
	AuthFile
}

func (a *Auth) RealmsFile() (*RealmsFile, error) {
	realms := &RealmsFile{}
	realmsFile := a.Path.Concat("Realms.txt")

	if realmsFile.IsExtant() {
		rdata, err := realmsFile.ReadAll()
		if err != nil {
			return nil, err
		}

		err = text.Unmarshal(rdata, realms)
		if err != nil {
			return nil, err
		}
	} else {
		realms.Realms = make(map[uint64]Realm)
	}

	return realms, nil
}

func LoadAuth(at string) (*Auth, error) {
	ac := new(Auth)
	ac.Path = etc.ParseSystemPath(at)
	c, err := ac.Path.Concat("Auth.txt").ReadAll()
	if err != nil {
		return nil, err
	}

	err = text.Unmarshal(c, &ac.AuthFile)
	if err != nil {
		return nil, err
	}

	ac.Certificate, err = tls.LoadX509KeyPair(
		ac.Path.Concat("cert.pem").Render(),
		ac.Path.Concat("key.pem").Render(),
	)

	if ac.HostExternal == "" {
		ac.HostExternal = "localhost"
	}

	if err != nil {
		return nil, err
	}

	return ac, nil
}

const DefaultAuth = `{
	// the TCP/IP address to listen the Gophercraft HTTP API on.
  // You can reverse proxy this however you like.
	HTTPInternal 0.0.0.0:8086

	// The public hostname of your Gophercraft API server.
	// if left uncommented, it will be set to localhost
	// this is needed to tell the client where the REST logon service is located.
	// 
	// HostExternal gcraft.example.com

	// The TCP/IP addresses to listen Auth/Realmlist servers on.
	// Keep these unchanged, unless you really know what you're doing.
	AuthListen 0.0.0.0:3724
	BnetListen 0.0.0.0:1119
	BnetRESTListen 0.0.0.0:1120

	// Database options
	// the go-xorm SQL driver to use.
	DBDriver mysql

	// the go-xorm SQL URL to use.
	DBURL root:password@/gcraft_auth

	// Alpha: uncomment this to use the Alpha protocol.
	// AlphaRealmlistListen 0.0.0.0:9100
}`

func GenerateTLSKeyPair(at string) error {
	dir := etc.ParseSystemPath(at)
	return genPair(
		dir.Concat("cert.pem").Render(),
		dir.Concat("key.pem").Render())
}

func GenerateDefaultAuth(at string) error {
	path := etc.ParseSystemPath(at)

	if path.IsExtant() {
		return ErrAlreadyExists
	}

	if err := path.MakeDir(); err != nil {
		return err
	}

	path.Concat("Auth.txt").WriteAll([]byte(DefaultAuth))

	return GenerateTLSKeyPair(path.Render())
}
