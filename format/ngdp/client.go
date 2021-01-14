package ngdp

import (
	"encoding/hex"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/superp00t/gophercraft/format/ngdp/ccfg"
)

type Hash [16]byte

func (h Hash) String() string {
	return hex.EncodeToString(h[:])
}

type Client struct {
	*Agent
	Opt         uint64
	Dir         string
	CDN         *CDN
	Build       Hash
	BuildConfig BuildConfig
}

type CDN struct {
	Region     string   `ccfg:"Name"`
	Path       string   `ccfg:"Path"`
	Hosts      []string `ccfg:"Hosts"`
	Servers    []string `ccfg:"Servers"`
	ConfigPath string   `ccfg:"ConfigPath"`
}

type Version struct {
	Region        string
	BuildConfig   Hash
	CDNConfig     Hash
	KeyRing       Hash
	BuildId       uint32
	VersionsName  string
	ProductConfig Hash
}

// OpenOnline open a remote NGDP volume using the set CDN server
// Without setting additional options, the client will not touch the disk and will read directly from HTTP to memory.
func (ag *Agent) OpenOnline(programID string) (*Client, error) {
	c := new(Client)
	c.Agent = ag
	c.Opt |= OptUseTACTNetwork

	versionListFile, err := ag.DownloadFn(ag.HostServer + "/" + programID + "/versions")
	if err != nil {
		return nil, err
	}
	var versionList []Version
	err = ccfg.NewDecoder(versionListFile).Decode(&versionList)
	if err != nil {
		return nil, err
	}
	if len(versionList) == 0 {
		return nil, fmt.Errorf("no versions")
	}
	for _, version := range versionList {
		if version.Region == ag.Region {
			c.Build = version.BuildConfig
		}
	}

	cdnListFile, err := ag.DownloadFn(ag.HostServer + "/" + programID + "/cdns")
	if err != nil {
		return nil, err
	}
	var cdnList []CDN
	err = ccfg.NewDecoder(cdnListFile).Decode(&cdnList)
	if err != nil {
		return nil, err
	}
	if len(cdnList) == 0 {
		return nil, fmt.Errorf("ngdp: no cdns")
	}
	var cdn *CDN
	cdn = &cdnList[0]
	for i, cdnListing := range cdnList {
		if cdnListing.Region == ag.Region {
			cdn = &cdnList[i]
			break
		}
	}
	if len(cdn.Servers) == 0 {
		return nil, fmt.Errorf("ngdp: no servers")
	}
	c.CDN = cdn

	return c, c.Init()
}

func (cl *Client) Init() error {
	buildConfigFile, err := cl.openHash("config", cl.Build)
	if err != nil {
		return err
	}

	err = ccfg.NewDecoder(buildConfigFile).Decode(&cl.BuildConfig)
	if err != nil {
		return err
	} 
	return nil
}

// // OpenCASC open an existing CASC archive and use its configuration files to figure out what to download
// func (ag *Agent) OpenCASC(path string) (*Client, error) {
// 	var c Client
// 	c.Agent = ag
// 	c.Opt |= OptUseCASContainer
// 	c.Dir = filepath.Clean(path)
// }

func (cl *Client) realPath(path string) string {
	forwardSlash := strings.Split(path, "/")

	return filepath.Join(append([]string{cl.Dir}, forwardSlash...)...)
}

func (cl *Client) exists(path string) bool {
	if cl.Opt&OptUseCASContainer == 0 {
		panic("cannot check existence")
	}

	_, err := os.Stat(cl.realPath(path))
	return err == nil
}

func (cl *Client) openHash(kind string, h Hash) (io.ReadCloser, error) {
	str := h.String()
	p1, p2 := str[0:2], str[2:4]
	return cl.openRawFile(fmt.Sprintf("%s/%s/%s/%s", kind, p1, p2, str))
}

// Retrieve a file from TACT and/or CASC
func (cl *Client) openRawFile(path string) (io.ReadCloser, error) {
	if cl.Opt&OptUseCASContainer != 0 {
		if cl.exists(path) {
			return os.Open(cl.realPath(path))
		}
	}

	// if cl.Opt&(OptUseCASContainer|OptTACTNetwork)!=0  {
	// *download and return file as readerreader*

	if cl.Opt&OptUseTACTNetwork != 0 {
		u, err := url.Parse(cl.CDN.Servers[0])
		if err != nil {
			return nil, err
		}
		u.Path = "/" + cl.CDN.Path + "/" + path
		return cl.DownloadFn(u.String())
	}

	panic("flags")
}
