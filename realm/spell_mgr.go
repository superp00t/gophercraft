package realm

import (
	"encoding/hex"

	"github.com/superp00t/etc"
	"github.com/superp00t/gophercraft/packet"
	"github.com/superp00t/gophercraft/realm/wdb"
	"github.com/superp00t/gophercraft/vsn"
)

func (s *Session) SendLoginSpell() {
	p := packet.NewWorldPacket(packet.SMSG_SPELL_GO)
	s.encodePackedGUID(p, s.GUID())
	// s.encodePackedGUID(p, s.GUID())

	// p.WriteByte(0)     //  pending cast
	p.WriteUint32(836) // login

	p.WriteUint16(0) // flags
	p.WriteUint32(packet.GetMSTime())
	p.WriteUint32(0) // flags
	p.WriteUint32(0) // ticks count

	s.SendAsync(p)
}

type SpellMgr struct {
}

func (s *Session) shex(pt packet.WorldType, data string) {
	hx, err := hex.DecodeString(data)
	if err != nil {
		panic(err)
	}
	p := packet.NewWorldPacket(pt)
	p.Buffer = etc.FromBytes(hx)
	s.SendAsync(p)
}

func (s *Session) SendSpellList() {
	// s.shex(packet.SMSG_INITIAL_SPELLS, "0060002a8500006f8200006c820000d77d000074760000bb620000b2620000b06200009a620000946200008f620000896200007c62000073620000705d00000b5600009454000093540000087400004650000045500000064f0000863100002b8500004850000047500000412d0000a52300009c230000752300009469000076230000510000009006000078620000370c0000cb00000089040000ee020000d501000047000000c6000000e63c0000c5000000c40000007c860000c80000006b000000c7000000670300000a0100009e020000621c00009c0400001a5900000a020000ca0000009d0200007e140000c2200000530d0000e30000009a090000a4020000630100003a2d0000cb1900009313000000080000ca0b00004e0900009909000008010000ea0b00007f0a0000050a0000cc0a0000070a0000cc000000af090000cb0c0000250d0000b7060000b514000059180000a20200006618000067180000957600004d1900004e190000212200009a190000631c000043480000bb1c00000000")

	p := packet.NewWorldPacket(packet.SMSG_INITIAL_SPELLS)

	var spells []wdb.LearnedAbility
	s.DB().Where("player = ?", s.PlayerID()).Find(&spells)

	spells = append(spells, wdb.LearnedAbility{
		Player: s.PlayerID(),
		Spell:  668,
	}, wdb.LearnedAbility{
		Player: s.PlayerID(),
		Spell:  669,
	})

	p.WriteByte(0)
	p.WriteUint16(uint16(len(spells)))

	for _, spell := range spells {
		p.WriteUint32(spell.Spell)
		if s.Build().AddedIn(vsn.V3_3_5a) {
			p.WriteUint16(0) // unk
		}
	}

	// cooldowns
	p.WriteUint16(0)

	s.SendAsync(p)
}
