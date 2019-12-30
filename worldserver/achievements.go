package worldserver

import (
	"github.com/superp00t/gophercraft/packet"
)

func (s *Session) SendAllAcheivementData() {
	p := packet.NewWorldPacket(packet.SMSG_ALL_ACHIEVEMENT_DATA)
	p.WriteUint32(0)
	s.SendAsync(p)
}
