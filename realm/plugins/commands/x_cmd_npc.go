package commands

import (
	"github.com/superp00t/gophercraft/realm"
	"github.com/superp00t/gophercraft/realm/wdb"
)

func cmdAddNPC(s *realm.Session, npcID string) {
	var cr *wdb.CreatureTemplate
	s.DB().GetData(npcID, &cr)
	if cr == nil {
		s.Warnf("No CreatureTemplate could be found with the ID %s", npcID)
		return
	}

	creature := s.WS.NewCreature(cr, s.Position())
	s.Map().AddObject(creature)

	s.Warnf("Object spawned successfully: %s", creature.GUID())
}
