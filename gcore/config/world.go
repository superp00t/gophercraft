package config

import (
	"crypto/tls"
	"fmt"

	"github.com/superp00t/gophercraft/gcore/sys"

	"github.com/go-yaml/yaml"
	"github.com/superp00t/etc"
)

type Flags map[string]interface{}

var (
	DefaultFlags = Flags{
		"world.maxVisibilityRange": float32(1000.0),
		"xp.rate":                  float32(1.0),
		"xp.startLevel":            byte(1),
	}
)

type WorldFile struct {
	Flags                 `yaml:"flags"`
	Version               uint32 `yaml:"version"`
	Listen                string `yaml:"world_listen"`
	Type                  string `yaml:"type"`
	RealmID               uint64 `yaml:"realm_id"`
	RealmName             string `yaml:"realm_name"`
	RealmDescription      string `yaml:"realm_description"`
	DBDriver              string `yaml:"db_driver"`
	DBURL                 string `yaml:"db_url"`
	PublicAddress         string `yaml:"public_address"`
	WardenEnabled         bool   `yaml:"warden_enable"`
	ShowSQL               bool   `yaml:"show_sql"`
	DatapackDir           string `yaml:"datapack_dir"`
	AuthServer            string `yaml:"auth_server"`
	AuthServerFingerprint string `yaml:"auth_server_fingerprint"`
}

type World struct {
	Path etc.Path
	WorldFile
	Certificate tls.Certificate
}

func LoadWorld(at string) (*World, error) {
	wc := new(World)

	wc.Path = etc.ParseSystemPath(at)

	c, err := wc.Path.Concat("config.yml").ReadAll()
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(c, &wc.WorldFile)
	if err != nil {
		return nil, err
	}

	wc.Certificate, err = tls.LoadX509KeyPair(
		wc.Path.Concat("cert.pem").Render(),
		wc.Path.Concat("key.pem").Render(),
	)

	if err != nil {
		return nil, err
	}

	wc.Flags = MergeFlagsWithDefaults(wc.Flags)

	if wc.DatapackDir == "" {
		dpackDir := wc.Path.Concat("datapacks")
		wc.DatapackDir = dpackDir.Render()
	}

	return wc, nil
}

const DefaultWorld = `# build ID
version: %d

# the internal IP address to listen on
world_listen: 0.0.0.0:8085

# The display name of your server. You can change this as you please.
realm_name: Placeholder Name %d

# Description of your server. This will appear in the Gophercraft website.
realm_description: Put the description for your server here!

# The reference ID of your server. This should remain constant to avoid losing your data.
realm_id: %d

# database driver
db_driver: mysql

# database URL
db_url: root:password@/gcraft_world_%d

# external address (should be accessible from the client's computer)
public_address: 0.0.0.0:8085

# Address of RPC server (replace 127.0.0.1 with host_external in gcraft_auth/config.yml)
auth_server: 127.0.0.1:3724

# RPC server fingerprint
auth_server_fingerprint: %s
`

func (a *Auth) GenerateDefaultWorld(version uint32, id uint64, at string) error {
	path := etc.ParseSystemPath(at)

	if path.IsExtant() {
		return ErrAlreadyExists
	}

	if err := path.MakeDir(); err != nil {
		return err
	}

	asf, err := sys.GetCertFileFingerprint(a.Path.Concat("cert.pem").Render())
	if err != nil {
		return err
	}

	wfile := fmt.Sprintf(DefaultWorld, version, id, id, id, asf)

	path.Concat("config.yml").WriteAll([]byte(wfile))

	path.Concat("datapacks").MakeDir()

	if err := GenerateTLSKeyPair(path.Render()); err != nil {
		return err
	}

	wsf, err := sys.GetCertFileFingerprint(path.Concat("cert.pem").Render())
	if err != nil {
		return err
	}

	realms := RealmsFile{}
	realmsFile := a.Path.Concat("realms.yml")

	if realmsFile.IsExtant() {
		rdata, err := realmsFile.ReadAll()
		if err != nil {
			return err
		}

		err = yaml.Unmarshal(rdata, &realms)
		if err != nil {
			return err
		}
	} else {
		realms.Realms = make(map[uint64]*Realm)
	}

	realms.Realms[id] = &Realm{
		FP: wsf,
	}

	rdata, err := yaml.Marshal(realms)
	if err != nil {
		return err
	}

	return realmsFile.WriteAll(rdata)
}

func (cf Flags) Get(key string) interface{} {
	data, ok := cf[key]
	if !ok {
		panic("no data for config " + key)
	}

	return data
}

func (cf Flags) Float32(key string) float32 {
	return cf.Get(key).(float32)
}

func (cf Flags) Byte(key string) byte {
	return cf.Get(key).(byte)
}

func MergeFlagsWithDefaults(input Flags) Flags {
	output := make(Flags)

	for k, v := range DefaultFlags {
		output[k] = v
	}

	for k, v := range input {
		output[k] = v
	}

	return output
}

func (c World) String() string {
	return "[" + c.RealmName + "] " + c.PublicAddress
}
