package vsn

import "fmt"

type Build uint32

var (
	names = map[Build]string{
		3368:  "0.5.3 (Alpha)",
		5875:  "1.12.1 (Vanilla)",
		8606:  "2.4.3 (TBC)",
		12340: "3.3.5a (WoTLK)",
	}
)

const (
	Alpha   Build = 3368
	V1_12_1 Build = 5875
	V2_4_3  Build = 8606
	V3_3_5a Build = 12340
)

func (b Build) String() string {
	if str := names[b]; str != "" {
		return fmt.Sprintf("%s (%d)", str, b)
	}

	return fmt.Sprintf("unknown version (%d)", b)
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
