package commands

import "github.com/superp00t/gophercraft/realm"

func cmdAppear(s *realm.Session, name string) {
	player, err := s.WS.GetSessionByPlayerName(name)
	if err != nil {
		s.Warnf("no such player as '%s' found.", name)
		return
	}

	// todo: escape user input

	if player.CurrentPhase != s.CurrentPhase {
		s.Warnf("'%s' is currently in phase %d. You must join this phase if you want to appear at this player's location.", name, player.CurrentPhase)
		return
	}

	targetMap := player.CurrentMap

	s.TeleportTo(targetMap, player.Position())
}
