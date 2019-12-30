package worldserver

import (
	"github.com/superp00t/etc/yo"
	"github.com/superp00t/gophercraft/packet"
)

func (s *Session) IntroductoryPackets() {
	if s.WS.Config.Version == 5875 {
		return
	}

	if len(s.AddonData) != 1 {
		s.SendAsync(packet.SendAddonsInfo(s.AddonData))
		yo.Println("Addon info sent")
	}

	v2 := packet.NewWorldPacket(packet.SMSG_CLIENTCACHE_VERSION)
	v2.WriteUint32(s.Version())
	s.SendAsync(v2)

	v3 := packet.NewWorldPacket(packet.SMSG_TUTORIAL_FLAGS)
	for i := 0; i < 8; i++ {
		v3.WriteUint32(0x111111)
	}
	s.SendAsync(v3)
}
