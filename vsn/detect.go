package vsn

import (
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"

	"github.com/superp00t/etc"

	reader "github.com/Velocidex/binparsergen/reader"
	pe "github.com/Velocidex/go-pe"
)

var (
	CoreVersion = "0.1"
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

type Build uint32

var (
	names = map[Build]string{
		5875:  "1.12.1 Vanilla (5875)",
		12340: "3.3.5a Wrath of the Lich King (12340)",
	}
)

func (b Build) String() string {
	if str := names[b]; str != "" {
		return str
	}

	return fmt.Sprintf("unknown version (%d)", b)
}

func DetectGame(folder string) (Build, error) {
	path := etc.ParseSystemPath(folder)

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
