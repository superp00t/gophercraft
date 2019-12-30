package guid

import (
	"encoding/binary"
	"fmt"
	"io"
)

var (
	None GUID = Classic(0)
)

type GUID struct {
	Hi uint64
	Lo uint64
}

func (g GUID) HighType() HighType {
	return (HighType(g.Hi>>58) & 0x3F)
}

func (g GUID) RealmID() uint32 {
	return uint32(g.Hi>>42) & 0x1FFF
}

func (g GUID) Counter() uint64 {
	switch g.HighType() {
	case Transport:
		return (g.Hi >> 38) & uint64(0xFFFFF)
	default:
		break
	}
	return g.Lo & uint64(0x000000FFFFFFFFFF)
}

func (g GUID) String() string {
	// return fmt.Sprintf("(%s) 0x%016X", g.HighType(), g.Lo)

	switch g.HighType() {
	case Player:
		return fmt.Sprintf("Player-%d-%08X", g.RealmID(), g.Counter())
	case Null:
		return "Nil"
	default:
		return fmt.Sprintf("%s-%016X-%016X", g.HighType(), g.Hi, g.Lo)
	}
}

func rb(r io.Reader) uint8 {
	b := make([]byte, 1)
	r.Read(b)
	return b[0]
}

func decodeMasked64(mask uint8, r io.Reader) uint64 {
	var res uint64

	for i := uint64(0); i < 8; i++ {
		if (mask & (1 << i)) != 0 {
			res += uint64(rb(r)) << (i * 8)
		}
	}

	return res
}

func DecodePacked64(r io.Reader) uint64 {
	mask := rb(r)
	if mask == 0 {
		return 0
	}

	return decodeMasked64(mask, r)
}

func DecodePacked128(r io.Reader) GUID {
	loMask := rb(r)
	hiMask := rb(r)

	lo := decodeMasked64(loMask, r)
	hi := decodeMasked64(hiMask, r)

	return GUID{hi, lo}
}

func (g GUID) IsUnit() bool {
	return g.HighType() == Creature || g.HighType() == Player
}

func encodeMasked64(value uint64) (uint8, []byte) {
	bitMask := uint8(0)
	packGUID := make([]byte, 8)
	size := 0

	for i := uint64(0); value != 0; i++ {
		// Convert byte
		if (value & 0xFF) > 0 {
			bitMask |= uint8(1 << i)
			packGUID[size] = uint8(value & 0xFF)
			size++
		}

		// Read next byte
		value >>= 8
	}

	return bitMask, packGUID[:size]
}

func (g GUID) EncodePacked(version uint32, w io.Writer) {
	switch {
	case version <= 12340:
		mask, bytes := encodeMasked64(g.Classic())
		w.Write([]byte{mask})
		if mask > 0 {
			w.Write(bytes)
		}
	default:
		loMask, loBytes := encodeMasked64(g.Lo)
		hiMask, hiBytes := encodeMasked64(g.Hi)
		w.Write([]byte{loMask, hiMask})
		w.Write(loBytes)
		w.Write(hiBytes)
	}
}

func (g GUID) Classic() uint64 {
	highTypeClassic := htSupport[12340][g.HighType()]
	return (highTypeClassic << 48) | g.Lo
}

func (g GUID) EncodeUnpacked(version uint32, w io.Writer) {
	switch {
	case version <= 12340:
		e := make([]byte, 8)
		binary.LittleEndian.PutUint64(e, g.Classic())
		w.Write(e)
	default:
		panic(fmt.Errorf("update: can't encode this as packed in version %d", version))
	}
}

func (g GUID) Cmp(g2 GUID) int {
	if g == g2 {
		return 2 // exact match.
	}

	if g.HighType() == Player {
		if g.Counter() == g2.Counter() {
			return 1 // Same GUID counter, not the same realm ID
		}
	}

	return 0
}

func (g GUID) SetRealmID(realmID uint64) GUID {
	g2 := g
	g2.Hi |= (realmID << 42)
	return g2
}
