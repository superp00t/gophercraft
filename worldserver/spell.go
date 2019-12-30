package worldserver

import "github.com/superp00t/gophercraft/packet"

func (s *Session) SendLoginSpell() {
	p := packet.NewWorldPacket(packet.SMSG_SPELL_GO)
	s.encodePackedGUID(p, s.GUID())
	s.encodePackedGUID(p, s.GUID())

	p.WriteByte(0)     //  pending cast
	p.WriteUint32(836) // login
	p.WriteUint32(0)   // flags
	p.WriteUint32(0)   // ticks count

	s.SendAsync(p)
}
