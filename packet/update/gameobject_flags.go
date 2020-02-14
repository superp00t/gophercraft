package update

import "strings"

type GameObjectFlags uint32

const (
	GOInUse GameObjectFlags = 1 << iota
	GOLocked
	GOUntargetable
	GOTransport
	GOUnselectable
	GONoDespawn
	GOTriggered
	GODamaged
	GODestroyed
)

func (gof GameObjectFlags) String() string {
	var s []string
	if gof&GOInUse != 0 {
		s = append(s, "InUse")
	}
	if gof&GOLocked != 0 {
		s = append(s, "Locked")
	}
	if gof&GOUntargetable != 0 {
		s = append(s, "Untargetable")
	}
	if gof&GOTransport != 0 {
		s = append(s, "Transport")
	}
	if gof&GOUnselectable != 0 {
		s = append(s, "Unselectable")
	}
	if gof&GONoDespawn != 0 {
		s = append(s, "NoDespawn")
	}
	if gof&GOTriggered != 0 {
		s = append(s, "Triggered")
	}
	if gof&GODamaged != 0 {
		s = append(s, "Damaged")
	}
	if gof&GODestroyed != 0 {
		s = append(s, "Destroyed")
	}
	if len(s) == 0 {
		return ""
	}

	return strings.Join(s, "|")
}

func ParseGameObjectFlags(str string) (GameObjectFlags, error) {
	var o GameObjectFlags

	if str == "" {
		return 0, nil
	}

	s := strings.Split(str, "|")

	for _, v := range s {
		if v == "Locked" {
			o |= GOLocked
		}
		if v == "Untargetable" {
			o |= GOUntargetable
		}
		if v == "Transport" {
			o |= GOTransport
		}
		if v == "Unselectable" {
			o |= GOUnselectable
		}
		if v == "NoDespawn" {
			o |= GONoDespawn
		}
		if v == "Triggered" {
			o |= GOTriggered
		}
		if v == "Damaged" {
			o |= GODamaged
		}
		if v == "Destroyed" {
			o |= GODestroyed
		}
	}

	return o, nil
}
