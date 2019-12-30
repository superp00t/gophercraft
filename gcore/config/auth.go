package config

import (
	"crypto/tls"
	"fmt"

	"github.com/go-yaml/yaml"
	"github.com/superp00t/etc"
)

var (
	ErrAlreadyExists = fmt.Errorf("config: already exists")
)

type AuthFile struct {
	HTTPInternal   string `yaml:"http_internal"`
	HostExternal   string `yaml:"host_external"`
	AuthListen     string `yaml:"auth_listen"`
	BnetListen     string `yaml:"bnet_listen"`
	BnetRESTListen string `yaml:"bnet_rest_listen"`
	DBDriver       string `yaml:"db_driver"`
	DBURL          string `yaml:"db_url"`
}

type RealmsFile struct {
	Realms map[uint64]*Realm `yaml:"realms"`
}

type Realm struct {
	FP     string `yaml:"fp"`
	Armory string `yaml:"armory,omitempty"`
}

type Auth struct {
	Path        etc.Path
	Certificate tls.Certificate
	AuthFile
}

func (a *Auth) RealmsFile() (*RealmsFile, error) {
	realms := &RealmsFile{}
	realmsFile := a.Path.Concat("realms.yml")

	if realmsFile.IsExtant() {
		rdata, err := realmsFile.ReadAll()
		if err != nil {
			return nil, err
		}

		err = yaml.Unmarshal(rdata, realms)
		if err != nil {
			return nil, err
		}
	} else {
		realms.Realms = make(map[uint64]*Realm)
	}

	return realms, nil
}

func LoadAuth(at string) (*Auth, error) {
	ac := new(Auth)
	ac.Path = etc.ParseSystemPath(at)
	c, err := ac.Path.Concat("config.yml").ReadAll()
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(c, &ac.AuthFile)
	if err != nil {
		return nil, err
	}

	ac.Certificate, err = tls.LoadX509KeyPair(
		ac.Path.Concat("cert.pem").Render(),
		ac.Path.Concat("key.pem").Render(),
	)

	if err != nil {
		return nil, err
	}

	return ac, nil
}

const DefaultAuth = `# the TCP/IP address to listen the Gophercraft HTTP API on.
# You can reverse proxy this however you like.
http_internal: 0.0.0.0:8086

# The public hostname of your Gophercraft API server.
# if left uncommented, it will be set to localhost
# this is needed to tell the client where the REST logon service is located.
#
# host_external: gcraft.example.com

# The TCP/IP addresses to listen Auth/Realmlist servers on.
# it's not a good idea not to change these ports.
auth_listen: 0.0.0.0:3724
bnet_listen: 0.0.0.0:1119
bnet_rest_listen: 0.0.0.0:1120

# ~~ DB OPTIONS ~~
# the go-xorm SQL driver to use.
db_driver: mysql

# the go-xorm SQL URL to use.
db_url: root:password@/gcraft_auth
`

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

	path.Concat("config.yml").WriteAll([]byte(DefaultAuth))

	return GenerateTLSKeyPair(path.Render())
}
