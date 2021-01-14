package commands

import (
	"time"

	"github.com/superp00t/gophercraft/realm"
)

func cmdSummon(s *realm.Session, playerName string) {
	plyr, err := s.WS.GetSessionByPlayerName(playerName)
	if err != nil {
		s.NoSuchPlayer(playerName)
		return
	}

	if plyr.GUID() == s.GUID() {
		s.Warnf("You can't summon yourself!")
		return
	}

	plyr.SetSummonLocation(s.CurrentPhase, s.CurrentMap, s.Position())
	plyr.SendSummonRequest(s.GUID(), s.ZoneID, 2*time.Minute)
}
