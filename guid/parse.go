package guid

import (
	"encoding/binary"
	"fmt"
	"io"
	"strconv"
	"strings"
)

var Nil = GUID{0, 0}

func Classic(u64 uint64) GUID {
	if u64 == 0 {
		return Nil
	}

	hiBits := (u64 >> 48) & 0x0000FFFF

	realHt := Null

	for ht, mask := range htSupport[12340] {
		if hiBits == mask {
			realHt = ht
		}
	}

	g := GUID{}
	g.Hi = uint64(realHt) << 58
	g.Lo = u64 & 0xFFFFFFFFFFFF
	return g
}

func DecodePacked(version uint32, reader io.Reader) (GUID, error) {
	switch {
	case version <= 12340:
		u64 := DecodePacked64(reader)
		fixed := Classic(u64)
		return fixed, nil
	default:
		g := DecodePacked128(reader)
		return g, nil
	}
}

func DecodeUnpacked(version uint32, reader io.Reader) (GUID, error) {
	switch {
	case version <= 12340:
		var bytes [8]byte
		reader.Read(bytes[:])
		u := binary.LittleEndian.Uint64(bytes[:])
		return Classic(u), nil
	// Modern version, no conversion necessary!
	default:
		var bytes [16]byte
		reader.Read(bytes[:])
		low := binary.LittleEndian.Uint64(bytes[0:8])
		high := binary.LittleEndian.Uint64(bytes[8:16])
		return GUID{Lo: low, Hi: high}, nil
	}
}

func FromString(s string) (GUID, error) {
	if len(s) < 2 {
		return Nil, fmt.Errorf("guid: too short")
	}

	// hex format
	if s[0] == '0' && s[1] == 'x' {
		switch {
		case len(s) <= 18:
			i, err := strconv.ParseUint(s, 0, 64)
			if err != nil {
				return Nil, err
			}

			return Classic(i), nil
		default:
			return Nil, fmt.Errorf("guid: unknown guid format")
		}
	}

	arguments := strings.Split(s, "-")

	if len(arguments) < 2 {
		return Nil, fmt.Errorf("guid: not enough parameters")
	}

	ht := Null

	switch arguments[0] {
	case "Player":
		realmid, err := strconv.ParseUint(arguments[1], 0, 64)
		if err != nil {
			return Nil, err
		}

		counter, err := strconv.ParseUint(arguments[2], 16, 64)
		if err != nil {
			return Nil, err
		}
		return RealmSpecific(Player, realmid, counter), nil
	case "Unit":
		ht = Creature
	case "Creature":
		ht = Creature
	case "Item":
		ht = Item
	case "GameObject":
		ht = GameObject
	case "DynamicObject":
		ht = DynamicObject
	default:
		return Nil, fmt.Errorf("guid: unknown type '%s'", arguments[0])
	}

	return Nil, fmt.Errorf("guid: cannot process this type %s at the moment", ht)
}
