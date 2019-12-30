package packet

import (
	"fmt"
	"strings"

	"github.com/superp00t/etc"
)

type WhoRequest struct {
	LevelMin, LevelMax    uint32
	PlayerName, GuildName string
	RaceMask, ClassMask   uint32
	ZonesCount            uint32
	Strings               []string
}

func UnmarshalWhoRequest(b []byte) (*WhoRequest, error) {
	wr := new(WhoRequest)
	e := etc.FromBytes(b)

	wr.LevelMin = e.ReadUint32()
	wr.LevelMax = e.ReadUint32()
	wr.PlayerName = e.ReadCString()
	wr.GuildName = e.ReadCString()
	wr.RaceMask = e.ReadUint32()
	wr.ClassMask = e.ReadUint32()
	wr.ZonesCount = e.ReadUint32()
	strCount := e.ReadUint32()

	if wr.ZonesCount > 10 {
		return nil, fmt.Errorf("packet: too many zones")
	}

	if strCount > 4 {
		return nil, fmt.Errorf("packet: too many strings")
	}

	strs := []string{}

	for x := uint32(0); x < strCount; x++ {
		strs = append(strs, strings.ToLower(e.ReadCString()))
	}

	wr.Strings = strs
	return wr, nil
}

type WhoMatch struct {
	PlayerName string
	GuildName  string
	Level      uint32
	Class      uint32
	Race       uint32
	ZoneID     uint32
}

type Who struct {
	WhoMatches []WhoMatch
}

func (w *Who) Packet() *WorldPacket {
	p := NewWorldPacket(SMSG_WHO)

	displayCt := len(w.WhoMatches)
	if displayCt > 49 {
		displayCt = 49
	}

	p.WriteUint32(uint32(displayCt))
	p.WriteUint32(uint32(len(w.WhoMatches)))

	for _, m := range w.WhoMatches {
		p.WriteCString(m.PlayerName)
		p.WriteCString(m.GuildName)
		p.WriteUint32(m.Level)
		p.WriteUint32(m.Class)
		p.WriteUint32(m.Race)
		p.WriteUint32(m.ZoneID)
	}

	return p
}
