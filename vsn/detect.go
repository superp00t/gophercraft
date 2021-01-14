package vsn

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/superp00t/etc"
	"github.com/superp00t/etc/yo"
	"github.com/superp00t/gophercraft/format/mpq"

	reader "github.com/Velocidex/binparsergen/reader"
	pe "github.com/Velocidex/go-pe"
)

type VolumeType uint8

const (
	NotAVolume VolumeType = iota
	MPQ
	NGDP
)

var (
	CoreVersion = "0.4"

	ErrInvalidPath = fmt.Errorf("vsn: invalid game path")
)

type Selector string

func (sl Selector) Match(b Build) bool {
	s := strings.Split(string(sl), "-")

	min := float64(0)
	max := math.Inf(1)

	if s[0] != "" {
		var err error
		min, err = strconv.ParseFloat(s[0], 64)
		if err != nil {
			panic(err)
		}
	}

	if s[1] != "" {
		var err error
		max, err = strconv.ParseFloat(s[1], 64)
		if err != nil {
			panic(err)
		}
	}

	fb := float64(b)

	return fb >= min && fb <= max
}

type componentinfo struct {
	XMLName   xml.Name `xml:"componentinfo"`
	Format    int      `xml:"format,attr"`
	Component struct {
		XMLName xml.Name `xml:"component"`
		Name    string   `xml:"name,attr"`
		Version uint32   `xml:"version,attr"`
	}
}

func DetectGame(folder string) (Build, error) {
	path := etc.ParseSystemPath(folder)

	if len(path) == 0 {
		return 0, ErrInvalidPath
	}

	head := path[len(path)-1]

	exes := []string{"WoWClient.exe", "WoW.exe", "Wow.exe", "WoW-64.exe", "Wow-64.exe"}

	for _, e := range exes {
		if e == head {
			return detectEXEBuild(folder)
		}
	}

	for _, e := range exes {
		pExePath := path.Concat(e)

		if pExePath.IsExtant() {
			return detectEXEBuild(path.Concat(e).Render())
		}
	}

	if head == "Data" {
		return DetectGame(path[:len(path)-1].Render())
	}

	// After all attempts have failed, bruteforce MPQs for version number. Some volumes will contain an XML file with the build ID enclosed.
	// This method is useful on Mac OS where the executable does not contain accessible metadata.
	var metaBuild Build

	if err := filepath.Walk(path.Render(), func(path string, info os.FileInfo, err error) error {
		vname := strings.ToLower(info.Name())

		if strings.HasPrefix(vname, "patch-") && strings.HasSuffix(vname, ".mpq") {
			volume, err := mpq.Open(path)
			if err != nil {
				return err
			}

			yo.Warn("[Bruteforce Version Detection] Checking", path)

			for _, filePath := range volume.ListFiles() {
				lFilePath := strings.ToLower(filePath)
				if strings.HasPrefix(lFilePath, "component.wow-") && strings.HasSuffix(lFilePath, ".txt") {
					xmlFile, err := volume.OpenFile(lFilePath)
					if err != nil {
						return err
					}

					xmlData, err := xmlFile.ReadBlock()
					if err != nil {
						return err
					}

					xmlFile.Close()

					var ci componentinfo

					err = xml.Unmarshal(xmlData, &ci)
					if err != nil {
						return err
					}

					pBuild := Build(ci.Component.Version)
					if pBuild > metaBuild {
						metaBuild = pBuild
					}

					break
				}
			}
			volume.Close()
		}

		return nil
	}); err != nil {
		return 0, err
	}

	if metaBuild != 0 {
		return metaBuild, nil
	}

	return 0, fmt.Errorf("version: could not find executable or other parseable version info")
}

func DetectVolumeLocation(folder string) (VolumeType, string, error) {
	path := etc.ParseSystemPath(folder)

	if len(path) == 0 {
		return 0, "", ErrInvalidPath
	}

	head := strings.ToLower(path[len(path)-1])

	if head != "data" {
		if path.Concat("Data").IsExtant() {
			return DetectVolumeLocation(path.Concat("Data").Render())
		}

		return 0, "", ErrInvalidPath
	}

	f, err := ioutil.ReadDir(path.Render())
	if err != nil {
		return 0, "", err
	}

	for _, fl := range f {
		if strings.HasSuffix(strings.ToLower(fl.Name()), ".mpq") {
			return MPQ, path.Render(), nil
		}
	}

	if path.Exists("config") {
		return NGDP, "", nil
	}

	return NotAVolume, "", ErrInvalidPath
}

// todo: implement test cases
func detectEXEBuild(path string) (Build, error) {
	fmt.Println("Analyzing", path, "...")

	// This may not always be the case.
	if strings.HasSuffix(path, "WoWClient.exe") {
		return Alpha, nil
	}

	file, err := os.Open(path)
	if err != nil {
		return 0, err
	}

	defer file.Close()

	reader, err := reader.NewPagedReader(file, 4096, 100)
	if err != nil {
		return 0, err
	}

	pe_file, err := pe.NewPEFile(reader)
	if err != nil {
		return 0, err
	}

	vinfo := pe_file.VersionInformation["FileVersion"]
	elements := strings.Split(vinfo, ", ")

	head := elements[len(elements)-1]

	i, err := strconv.ParseInt(head, 0, 64)
	if err != nil {
		return 0, err
	}

	return Build(i), nil
}
