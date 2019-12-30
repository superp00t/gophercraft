package worldserver

import "github.com/superp00t/gophercraft/packet"

func (s *Session) GetFreeTalentPoints() uint32 {
	return 100
}

func (s *Session) GetSpecsCount() uint8 {
	return 0
}

func (s *Session) GetActiveSpec() uint8 {
	return 0
}

func (s *Session) SendPlayerTalentsInfoData() {
	p := packet.NewWorldPacket(packet.SMSG_TALENTS_INFO)
	p.WriteUint32(s.GetFreeTalentPoints())
	p.WriteByte(s.GetActiveSpec())

	if s.GetSpecsCount() > 0 {
		// TODO: send talent data
		//	for s := 0; s
	}

	s.SendAsync(p)
}
