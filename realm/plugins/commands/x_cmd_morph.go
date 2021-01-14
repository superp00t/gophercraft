package commands

import (
	"github.com/superp00t/etc/yo"
	"github.com/superp00t/gophercraft/realm"
)

func cmdMorph(s *realm.Session, displayID uint32) {
	yo.Ok("Morphing to ", displayID)

	s.SetUint32("DisplayID", displayID)
	s.UpdatePlayer()
}
