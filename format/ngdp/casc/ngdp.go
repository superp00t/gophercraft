package casc

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/superp00t/etc"
	"github.com/superp00t/gophercraft/blizzardry/bcfg"
)

type NGDP struct {
	ProgramCode string
	URL         *url.URL
}

func (n *NGDP) GetCDNs() map[string]*CDNListing {
	neu := new(url.URL)
	*neu = *n.URL
	neu.Path = fmt.Sprintf("/%s/cdns", n.ProgramCode)

	b := get(neu.String())
	return ParseCDNList(b)
}

func (n *NGDP) GetVersions() map[string]*VersionListing {
	neu := new(url.URL)
	*neu = *n.URL
	neu.Path = fmt.Sprintf("/%s/versions", n.ProgramCode)

	b := get(neu.String())
	return ParseVersionList(b)
}

type Opts struct {
	ProgramCode, URL string
}

func NewNGDP(o Opts) (*NGDP, error) {
	s := new(NGDP)
	var err error
	s.URL, err = url.Parse(o.URL)
	if err != nil {
		return nil, err
	}

	s.ProgramCode = o.ProgramCode

	return s, nil
}

// func (n *NGDP) Connect(country string) error {
// 	cns := n.GetCDNs()
// 	vs := n.GetVersions()

// 	if cns[country] == nil {
// 		return fmt.Errorf("country code %s not found.", country)
// 	}

// 	l := cns[country]
// 	host := l.Domains[0]
// 	vers := vs[country]

// 	u := url.URL{
// 		Scheme: "http",
// 		Host:   host,
// 		Path: fmt.Sprintf(
// 			"/%s/config/%s/%s/%s",
// 			l.Path,
// 			vers.CDNConfig[0:2],
// 			vers.CDNConfig[2:4],
// 			vers.CDNConfig,
// 		),
// 	}

// 	header, cfgFile, err := bcfg.Parse(get(u.String()).Bytes())
// 	fmt.Println(err, header, spew.Sdump(cfgFile))
// 	return nil
// }

type CDNListing struct {
	Path    string
	Domains []string
}

func ParseCDNList(i *etc.Buffer) map[string]*CDNListing {
	c := make(map[string]*CDNListing)
	i.ReadString('\n')

	for {
		s := i.ReadString('\n')
		if s == "" {
			break
		}

		cols := strings.Split(s, "|")
		c[cols[0]] = &CDNListing{
			cols[1],
			strings.Split(cols[2], " ")}
	}

	return c
}

type VersionListing struct {
	BuildConfig string
	CDNConfig   string
	Keyring     string
	BuildID     string
	Version     string
}

func ParseVersionList(i *etc.Buffer) map[string]*VersionListing {
	c := make(map[string]*VersionListing)
	i.ReadString('\n')

	for {
		s := i.ReadString('\n')
		if s == "" {
			break
		}
		cols := strings.Split(s, "|")
		if len(cols) < 2 {
			continue
		}
		c[cols[0]] = &VersionListing{
			cols[1],
			cols[2],
			cols[3],
			cols[4],
			cols[5],
		}
	}

	return c
}

func get(url string) *etc.Buffer {
	b := etc.NewBuffer()

	re, err := http.NewRequest("GET", url, nil)
	if err != nil {
		panic(err)
	}

	c := &http.Client{}
	r, err := c.Do(re)
	if err != nil {
		panic(err)
	}

	fmt.Println(re.Method, url, r.Status, fmt.Sprintf("(%d)", r.ContentLength))

	io.Copy(b, r.Body)
	fmt.Println("...")
	return b
}

type Archive struct {
	base                   []string
	n                      *NGDP
	v                      *VersionListing
	path, server           string
	buildConfig, cdnConfig *bcfg.Config
}

func (n *NGDP) getConfig(path, server, hash string) (*bcfg.Config, error) {
	u := url.URL{
		Scheme: "http",
		Host:   server,
		Path: fmt.Sprintf(
			"/%s/config/%s/%s/%s",
			path,
			hash[0:2],
			hash[2:4],
			hash,
		),
	}

	return bcfg.Parse(get(u.String()).Bytes())
}

func (n *NGDP) OpenArchive(path, server string, v *VersionListing) (*Archive, error) {
	a := new(Archive)
	a.n = n
	a.v = v
	a.path = path
	a.server = server

	cdnConfig, err := n.getConfig(a.path, server, a.v.CDNConfig)
	if err != nil {
		return nil, err
	}

	buildConfig, err := n.getConfig(a.path, server, a.v.BuildConfig)
	if err != nil {
		return nil, err
	}

	a.cdnConfig = cdnConfig
	a.buildConfig = buildConfig

	return a, nil
}

func (a *Archive) Sync(path string) error {
	if a.base != nil {
		return fmt.Errorf("Sync should be called only once.")
	}

	if !archiveExistsAt(path) {
		a.base = filepath.SplitList(path)
		mkdir(append(base, a.configPath()...))
		ioutil.WriteFile(
			filepath.Join(
				append(base,
					append(a.configPath(), a.v.CDNConfig)...)...), a.cdnConfig.Encode(), 0700)

		ioutil.WriteFile(
			filepath.Join(
				append(base,
					append(a.configPath(), a.v.BuildConfig)...)...), a.buildConfig.Encode(), 0700)
		mkdir(append(base, "data"))
		mkdir(append(base, "indices"))
		mkdir(append(base, "patch"))
	} else {
		return fmt.Errorf("Cannot yet resume archives.")
	}

	return nil
}

func (a *Archive) makeAll()

type KeyTable struct {
	FirstHash []byte
	EntryHash []byte
}

type KeyEntry struct {
	KeyCount         uint16
	DecompressedSize uint32
	Hash             []byte
	Keys             [][]byte
}

type LayoutEntry struct {
	Key            []byte
	StringIndex    uint32
	Unk1           uint8
	CompressedSize uint32
}
type EncodingFile struct {
	Version        uint8
	ChecksumSizeA  uint8
	ChecksumSizeB  uint8
	FlagsA, FlagsB uint16
	SizeA, SizeB   uint32
	Unk1           uint8

	LayoutStrings      []string
	KeyTableIndex      []KeyTable
	KeyTableEntries    []KeyEntry
	LayoutTableIndex   []KeyTable
	LayoutTableEntries []LayoutEntry
}

func (a *Archive) buildURL(typ, hash string) string {
	u := url.URL{
		Scheme: "http",
		Host:   a.server,
		Path: fmt.Sprintf(
			"/%s/%s/%s/%s/%s",
			a.path,
			typ,
			hash[0:2],
			hash[2:4],
			hash,
		),
	}

	return u.String()
}

func (a *Archive) GetEncodingFile() (*EncodingFile, error) {
	// glogger.Fatal(a.buildConfig.Data["encoding"])
	buf := get(a.buildURL("data", a.buildConfig.Data["encoding"][1]))
	e := new(EncodingFile)
	header := buf.ReadFixedString(2)

	if header != "EN" {
		return nil, fmt.Errorf("Header is not EN... %s", header)
	}

	e.ChecksumSizeA = buf.ReadByte()
	e.ChecksumSizeB = buf.ReadByte()
	e.FlagsA = buf.ReadUint16()
	e.FlagsB = buf.ReadUint16()
	e.SizeA = buf.ReadBigUint32()
	e.SizeB = buf.ReadBigUint32()
	e.Unk1 = buf.ReadByte()
	StringSize := buf.ReadBigUint32()
	e.LayoutStrings = make([]string, StringSize)

	for i := uint32(0); i < StringSize; i++ {
		e.LayoutStrings[i] = buf.ReadCString()
	}

	e.KeyTableIndex = make([]KeyTable, e.SizeA)
	e.KeyTableEntries = make([]KeyEntry, e.SizeA)
	for i := uint32(0); i < e.SizeA; i++ {
		k1 := buf.ReadBytes(16)
		k2 := buf.ReadBytes(16)
		e.KeyTableIndex[i] = KeyTable{k1, k2}
	}

	for i := uint32(0); i < e.SizeA; i++ {
		entry := etc.MkBuffer(buf.ReadBytes(4096))
		ent := KeyEntry{}
		ent.KeyCount = entry.ReadUint16()
		ent.DecompressedSize = entry.ReadBigUint32()
		ent.Hash = entry.ReadBytes(16)
		ent.Keys = make([][]byte, ent.KeyCount)
		for x := uint16(0); x < ent.KeyCount; x++ {
			ent.Keys[x] = entry.ReadBytes(16)
		}
		e.KeyTableEntries[i] = ent
	}

	e.LayoutTableIndex = make([]KeyTable, e.SizeB)
	e.LayoutTableEntries = make([]LayoutEntry, e.SizeB)

	for i := uint32(0); i < e.SizeB; i++ {
		k1 := buf.ReadBytes(16)
		k2 := buf.ReadBytes(16)
		e.LayoutTableIndex[i] = KeyTable{k1, k2}
	}
	for i := uint32(0); i < e.SizeB; i++ {
		t := LayoutEntry{}
		pd := etc.MkBuffer(buf.ReadBytes(4096))
		t.Key = pd.ReadBytes(16)
		t.StringIndex = pd.ReadBigUint32()
		t.Unk1 = pd.ReadByte()
		t.CompressedSize = pd.ReadBigUint32()
		e.LayoutTableEntries[i] = t
	}

	return e, nil
}

func (a *Archive) configPath() []string {
	return []string{"config", a.v.CDNConfig[0:2], a.v.CDNConfig[2:4]}
}

func archiveExistsAt(path string) bool {
	if _, err := os.Stat(path); err == nil {
		return true
	}

	return false
}

func mkdir(path []string) {
	os.MkdirAll(filepath.Join(path...), 0700)
}
