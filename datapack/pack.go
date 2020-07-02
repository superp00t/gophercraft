// Package datapack allows users to extend Gophercraft with ease.
// It describes a simple format for loading data into Gophercraft and compatible clients
// Planned formats include:
//   Lua scripts
//   3d objects/textures
//   Map geometry
//   Sound files
//   JSON files for patching the server and client databases
package datapack

import (
	"archive/zip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"sort"

	"github.com/superp00t/etc"
	"github.com/superp00t/gophercraft/datapack/text"
)

type WriteFile io.WriteCloser

// Driver describes a mechanism for loading a datapack.
type Driver interface {
	Init(at string) (Opts, error)
	ReadFile(at string) (File, error)
	WriteFile(at string) (WriteFile, error)
	List() []string
	Close() error
}

type Opts uint64

const None Opts = 0

const (
	Read Opts = 1 << iota
	Write
)

var (
	drivers = map[string]func() Driver{
		"flatfile": func() Driver {
			return new(flatFile)
		},

		"archive": func() Driver {
			return new(archive)
		},
	}
)

func RegisterDriver(key string, value func() Driver) {
	if drivers[key] != nil {
		panic("datapack: " + key + " already registered")
	}

	drivers[key] = value
}

type PackConfig struct {
	Name          string
	Description   string
	Author        string
	Version       string
	Depends       []string
	ServerScripts []string
	ClientScripts []string
}

type Pack struct {
	PackConfig
	Opts
	Driver
}

func (p *Pack) WriteBytes(path string, data []byte) error {
	file, err := p.WriteFile(path)
	if err != nil {
		return err
	}

	_, err = file.Write(data)
	if err != nil {
		return err
	}

	err = file.Close()
	return err
}

func (p *Pack) ReadBytes(path string) ([]byte, error) {
	file, err := p.Driver.ReadFile(path)
	if err != nil {
		return nil, err
	}

	b, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	if err := file.Close(); err != nil {
		return nil, err
	}

	return b, nil
}

func OpenPack(path string) (*Pack, error) {
	p := new(Pack)
	driv := ""
	if etc.ParseSystemPath(path).IsDirectory() {
		driv = "flatfile"
	} else {
		driv = "archive"
	}

	var err error

	p.Driver = drivers[driv]()
	p.Opts, err = p.Driver.Init(path)
	if err != nil {
		return nil, err
	}

	yb, err := p.ReadBytes("Pack.txt")
	if err != nil {
		return nil, err
	}

	err = text.Unmarshal(yb, &p.PackConfig)
	if err != nil {
		return nil, err
	}

	return p, nil
}

func Open(directory string) (*Loader, error) {
	bp := etc.ParseSystemPath(directory)

	l := new(Loader)

	packs, err := ioutil.ReadDir(directory)
	if err != nil {
		return nil, err
	}

	var s []string

	for _, v := range packs {
		s = append(s, v.Name())
	}

	sort.Strings(s)

	for _, v := range s {
		path := bp.Concat(v).Render()
		pack, err := OpenPack(path)
		if err != nil {
			return nil, err
		}
		l.Volumes = append(l.Volumes, pack)
	}

	return l, nil
}

func (p *Pack) Exists(path string) bool {
	list := p.List()
	for _, v := range list {
		if v == path {
			return true
		}
	}

	return false
}

func Author(cfg PackConfig) (*Pack, error) {
	tempToken := etc.GenerateRandomUUID().String()
	tempDir := etc.TmpDirectory().Concat(tempToken)
	if err := tempDir.MakeDir(); err != nil {
		return nil, err
	}

	data, err := text.Marshal(cfg)
	if err != nil {
		panic(err)
	}
	tempDir.Concat("Pack.txt").WriteAll(data)

	pack, err := OpenPack(tempDir.Render())
	if err != nil {
		return nil, err
	}

	return pack, nil
}

func (p *Pack) ZipToFile(filename string) error {
	os.Remove(filename)

	newZipFile, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer newZipFile.Close()

	zipWriter := zip.NewWriter(newZipFile)
	defer zipWriter.Close()
	// Add files to zip
	for _, file := range p.List() {
		rdr, err := p.ReadFile(file)
		if err != nil {
			fmt.Println(file)
			panic(err)
			return err
		}

		if err = addFileToZip(zipWriter, file, rdr); err != nil {
			return err
		}

		rdr.Close()
	}
	return nil
}

func (p *Pack) Delete() error {
	switch f := p.Driver.(type) {
	case *flatFile:
		return os.RemoveAll(f.Base.Render())
	case *archive:
		return os.Remove(f.Path)
	default:
		return fmt.Errorf("unknown pack type")
	}
}
