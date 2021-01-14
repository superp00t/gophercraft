package realm

import (
	"github.com/superp00t/etc"
	"github.com/superp00t/gophercraft/guid"
	"github.com/superp00t/gophercraft/packet"
)

func (s *Session) QuestDone(q uint32) bool {
	return true
}

func (s *Session) HandleQuestgiverStatusQuery(e *etc.Buffer) {
	objectGUID := s.decodeUnpackedGUID(e)
	object := s.Map().GetObject(objectGUID)
	if object == nil {
		return
	}

	dialogStatus := uint8(1)

	switch object.TypeID() {
	case guid.TypeUnit:
	}

	s.SendQuestGiverStatus(objectGUID, dialogStatus)
}

func (s *Session) SendQuestGiverStatus(id guid.GUID, dialogStatus uint8) {
	status := packet.NewWorldPacket(packet.SMSG_QUESTGIVER_STATUS)
	id.EncodeUnpacked(s.Build(), status)
	status.WriteUint32(uint32(dialogStatus))
	s.SendAsync(status)
}
