package vsn

import (
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"strconv"
	"strings"

	"github.com/superp00t/etc"

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
	CoreVersion = "0.3"

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

func DetectGame(folder string) (Build, error) {
	path := etc.ParseSystemPath(folder)

	if len(path) == 0 {
		return 0, ErrInvalidPath
	}

	fmt.Println("detecting game in ", folder)

	head := path[len(path)-1]

	exes := []string{"WoW.exe", "Wow.exe", "WoW-64.exe", "Wow-64.exe"}

	for _, e := range exes {
		if e == head {
			return detectEXEBuild(folder)
		}
	}

	for _, e := range exes {
		if path.Exists(e) {
			return detectEXEBuild(path.Concat(e).Render())
		}
	}

	if head == "Data" {
		return DetectGame(path[:len(path)-1].Render())
	}

	return 0, fmt.Errorf("version: could not find executable")
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
		if strings.HasSuffix(fl.Name(), ".MPQ") {
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
	fmt.Println("reading ", path)

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
