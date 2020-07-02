package config

import (
	"crypto/tls"
	"fmt"
	"strconv"

	"github.com/superp00t/gophercraft/datapack/text"
	"github.com/superp00t/gophercraft/gcore/sys"
	"github.com/superp00t/gophercraft/vsn"

	"github.com/superp00t/etc"
)

type WorldVars map[string]string

var (
	Presets = map[string]WorldVars{}
)

func init() {
	Presets["allgm"] = WorldVars{
		"XP.Rate":       "1.0",
		"XP.StartLevel": "255",

		"Weather.On": "true",

		"Sync.VisibilityRange": "250.0",

		"PVP.Deathmatch":         "false",
		"PVP.AtWar":              "false",
		"PVP.LanguageBarrier":    "false",
		"PVP.CrossFactionGroups": "true",
	}
}

type WorldFile struct {
	Version               vsn.Build
	Listen                string
	Type                  string
	RealmID               uint64
	RealmName             string
	RealmDescription      string
	DBDriver              string
	DBURL                 string
	PublicAddress         string
	WardenEnabled         bool
	ShowSQL               bool
	DatapackDir           string
	AuthServer            string
	AuthServerFingerprint string
	WorldVars
}

type World struct {
	Path etc.Path
	WorldFile
	Certificate tls.Certificate
}

func LoadWorld(at string) (*World, error) {
	wc := new(World)

	wc.Path = etc.ParseSystemPath(at)

	c, err := wc.Path.Concat("World.txt").ReadAll()
	if err != nil {
		return nil, err
	}

	err = text.Unmarshal(c, &wc.WorldFile)
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

	if wc.DatapackDir == "" {
		dpackDir := wc.Path.Concat("Datapacks")
		wc.DatapackDir = dpackDir.Render()
	}

	if wc.WorldFile.WorldVars != nil {
		// Merge presets with user vars
		preset := wc.WorldFile.WorldVars["Config.Preset"]

		if preset != "" {
			wv := WorldVars{}

			if Presets[preset] == nil {
				return nil, fmt.Errorf("Config.Preset '%s' not found", preset)
			}

			for k, v := range Presets[preset] {
				wv[k] = v
			}

			for k, v := range wc.WorldFile.WorldVars {
				wv[k] = v
			}

			wc.WorldFile.WorldVars = wv
		}
	}

	return wc, nil
}

const DefaultWorld = `{
	// build ID
	Version %d

	// the internal IP address to listen on
	Listen 0.0.0.0:8085

	// The display name of your server. You can change this as you please.
	RealmName "Placeholder Name %d"

	// Description of your server. This will appear in the Gophercraft website.
	RealmDescription "Put the description for your server here!"

	// DO NOT EDIT after first run.
	RealmID %d

	// database driver
	DBDriver mysql

	// database URL
	DBURL root:password@/gcraft_world_%d

	// external address (should be accessible from the client's computer)
	PublicAddress 0.0.0.0:8085

	// Address of RPC server (replace 127.0.0.1 with host_external in gcraft_auth/Config.txt)
	AuthServer %s

	// RPC server fingerprint
	AuthServerFingerprint %s

	WorldVars
	{
		// Presets set default world vars for what kind of game you want to play.
		Config.Preset allgm
	}
}
`

func GenerateDefaultWorld(version uint32, id uint64, at string, authServer, serverFingerprint string) error {
	path := etc.ParseSystemPath(at)

	if path.IsExtant() {
		return ErrAlreadyExists
	}

	if err := path.MakeDir(); err != nil {
		return err
	}

	wfile := fmt.Sprintf(DefaultWorld, version, id, id, id, authServer, serverFingerprint)

	path.Concat("World.txt").WriteAll([]byte(wfile))

	path.Concat("Datapacks").MakeDir()

	if err := GenerateTLSKeyPair(path.Render()); err != nil {
		return err
	}

	return nil
}

func (a *Auth) GenerateDefaultWorld(version uint32, id uint64, at string) error {
	asf, err := sys.GetCertFileFingerprint(a.Path.Concat("cert.pem").Render())
	if err != nil {
		return err
	}

	if err = GenerateDefaultWorld(version, id, at, "127.0.0.1:3724", asf); err != nil {
		return nil
	}

	wsf, err := sys.GetCertFileFingerprint(etc.ParseSystemPath(at).Concat("cert.pem").Render())
	if err != nil {
		return err
	}

	realms := RealmsFile{}
	realmsFile := a.Path.Concat("Realms.txt")

	if realmsFile.IsExtant() {
		rdata, err := realmsFile.ReadAll()
		if err != nil {
			return err
		}

		err = text.Unmarshal(rdata, &realms)
		if err != nil {
			return err
		}
	} else {
		realms.Realms = make(map[uint64]Realm)
	}

	realms.Realms[id] = Realm{
		FP: wsf,
	}

	rdata, err := text.Marshal(realms)
	if err != nil {
		return err
	}

	return realmsFile.WriteAll(rdata)
}

func (c *World) String() string {
	return "[" + c.RealmName + "] @" + c.PublicAddress
}

func (c *World) GetString(name string) string {
	dat := c.WorldVars[name]
	return dat
}

func (c *World) Bool(name string) bool {
	on, err := strconv.ParseBool(c.GetString(name))
	if err != nil {
		panic(err)
	}
	return on
}

func (c *World) Int64(name string) int64 {
	i, err := strconv.ParseInt(c.GetString(name), 0, 64)
	if err != nil {
		panic(err)
	}
	return i
}

func (c *World) Uint64(name string) uint64 {
	u32, err := strconv.ParseUint(c.GetString(name), 0, 64)
	if err != nil {
		panic(err)
	}
	return u32
}

func (c *World) Uint32(name string) uint32 {
	u32, err := strconv.ParseUint(c.GetString(name), 0, 32)
	if err != nil {
		panic(err)
	}
	return uint32(u32)
}

func (c *World) Float32(name string) float32 {
	f32, err := strconv.ParseFloat(c.GetString(name), 32)
	if err != nil {
		panic(err)
	}

	return float32(f32)
}
