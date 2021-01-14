//Package vsn provides utilities for handling protocol versions, as well as Gophercraft software versions.
package vsn

import "fmt"

type Build uint32

var (
	names = map[Build]string{
		3368:  "Alpha",
		5875:  "Vanilla",
		8606:  "TBC",
		12340: "WoTLK",
		33369: "BfA",
	}
)

const (
	Alpha   Build = 3368
	V1_12_1 Build = 5875
	V2_4_3  Build = 8606
	V3_3_5a Build = 12340
	V8_3_0  Build = 33369
)

const NewAuthSystem = V8_3_0
const NewCryptSystem = V8_3_0

func (b Build) String() string {
	info := details[b]
	if info == nil {
		return fmt.Sprintf("unknown version (%d)", b)
	}

	str := fmt.Sprintf("%d.%d.%d", info.MajorVersion, info.MinorVersion, info.BugfixVersion)
	if info.HotfixVersion != "" {
		str += info.HotfixVersion
	}

	if name := names[b]; name != "" {
		str += " " + name
	}

	return fmt.Sprintf("%s (%d)", str, b)
}

func (hasFeature Build) AddedIn(update Build) bool {
	if hasFeature >= update {
		return true
	}

	return false
}

func (hasFeature Build) RemovedIn(update Build) bool {
	if hasFeature < update {
		return true
	}

	return false
}
