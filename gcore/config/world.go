package config

import (
	"crypto/tls"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/superp00t/gophercraft/datapack/text"
	"github.com/superp00t/gophercraft/gcore/sys"
	"github.com/superp00t/gophercraft/vsn"

	"github.com/superp00t/etc"
)

type WorldVars map[string]string

var (
	Presets = map[RealmType]WorldVars{}
)

func init() {
	Presets[RealmTypeRP] = WorldVars{
		"XP.Rate": "1.0",

		"Weather.On": "true",

		"Sync.VisibilityRange": "250.0",

		"PVP.Deathmatch":         "false",
		"PVP.AtWar":              "false",
		"PVP.CrossFactionGroups": "true",

		"Chat.LanguageBarrier":   "false",
		"Char.StartLevel":        "255",
		"Char.StartingCinematic": "true",
	}
}

type RealmType uint8

const (
	RealmTypeNone   = 0
	RealmTypePvP    = 1
	RealmTypeNormal = 4
	RealmTypeRP     = 6
	RealmTypeRP_PvP = 8
)

func (rt RealmType) EncodeWord() (string, error) {
	switch rt {
	case RealmTypeNone:
		return "None", nil
	case RealmTypePvP:
		return "PvP", nil
	case RealmTypeNormal:
		return "Normal", nil
	case RealmTypeRP:
		return "RP", nil
	case RealmTypeRP_PvP:
		return "RP-PvP", nil
	default:
		return "", fmt.Errorf("gcore: unknown RealmType %d", rt)
	}
}

func (RealmType) DecodeWord(out reflect.Value, wdata string) error {
	data := strings.ToLower(wdata)

	var rt RealmType

	switch data {
	case "none", "":
		rt = RealmTypeNone
	case "pvp":
		rt = RealmTypePvP
	case "normal":
		rt = RealmTypeNormal
	case "rp":
		rt = RealmTypeRP
	case "rp-pvp":
		rt = RealmTypeRP_PvP
	default:
		return fmt.Errorf("gcore: unrecognized realm type %s", wdata)
	}

	out.Set(reflect.ValueOf(rt))

	return nil
}

type WorldFile struct {
	Version               vsn.Build
	Listen                string
	RealmID               uint64
	RealmType             RealmType
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
	CPUProfile            string
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
		preset := wc.WorldFile.RealmType

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

	return wc, nil
}

const DefaultWorld = `{
	// build ID
	Version %d

	// the internal IP address to listen on
	Listen 0.0.0.0:8085

	// The display name of your server. You can change this as you please.
	RealmName "%s"

	// The type of server you want to create. A server of the type "RP" will allow for an all-GM roleplaying experience.
	// Note: changing this value will also change the default WorldVars.
	// You can always override these with custom values to fine-tune your server to the desired behavior!
	RealmType RP	

	// Description of your server. This will appear in the Gophercraft website.
	RealmDescription "Put the description for your server here!"

	// Editing this can lead to duplicate entries in the realm list.
	RealmID %d

	// database driver
	DBDriver %s

	// database URL
	DBURL "%s"

	// external address (should be accessible from the client's computer)
	PublicAddress 127.0.0.1:8085

	// Address of RPC server (replace 127.0.0.1 with host_external in gcraft_auth/Config.txt)
	AuthServer %s

	// RPC server fingerprint
	AuthServerFingerprint %s

	// Uncomment to perform CPU usage profiling.
	// CPUProfile "cpu.prof"

	WorldVars
	{
	}
}
`

func GenerateDefaultWorld(version uint32, name string, id uint64, sqlDriver, sqlDB string, at string, authServer, serverFingerprint string) error {
	path := etc.ParseSystemPath(at)

	if path.IsExtant() {
		return ErrAlreadyExists
	}

	if err := path.MakeDir(); err != nil {
		return err
	}

	wfile := fmt.Sprintf(DefaultWorld, version, name, id, sqlDriver, sqlDB, authServer, serverFingerprint)

	path.Concat("World.txt").WriteAll([]byte(wfile))

	path.Concat("Datapacks").MakeDir()

	if err := GenerateTLSKeyPair(path.Render()); err != nil {
		return err
	}

	return nil
}

func (a *Auth) GenerateDefaultWorld(version uint32, name string, id uint64, sqlDriver, sqlDB string, at string) error {
	asf, err := sys.GetCertFileFingerprint(a.Path.Concat("cert.pem").Render())
	if err != nil {
		return err
	}

	if err = GenerateDefaultWorld(version, name, id, sqlDriver, sqlDB, at, "127.0.0.1:3724", asf); err != nil {
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
