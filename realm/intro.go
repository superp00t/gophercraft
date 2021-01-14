package realm

import (
	"github.com/superp00t/gophercraft/packet"
	"github.com/superp00t/gophercraft/vsn"
)

func (s *Session) SendSessionMetadata() {
	if s.Build().RemovedIn(vsn.V2_4_3) {
		return
	}

	if s.Build().AddedIn(vsn.V3_3_5a) {
		v2 := packet.NewWorldPacket(packet.SMSG_CLIENTCACHE_VERSION)
		v2.WriteUint32(uint32(s.Build()))
		s.SendAsync(v2)
	}
}

func (s *Session) SendUnlearnSpell() {
	p := packet.NewWorldPacket(packet.SMSG_SEND_UNLEARN_SPELLS)
	p.WriteUint32(0)
	s.SendAsync(p)
}

func (s *Session) SendWorldLoginMetadata() {

}

func (s *Session) SendMetadataAfterSpawn() {
	if s.Build().AddedIn(vsn.V2_4_3) {

	}
}
