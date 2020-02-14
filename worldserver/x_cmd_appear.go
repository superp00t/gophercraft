package worldserver

func x_Appear(c *C) {
	name := c.String(0)

	c.Session.WS.PlayersL.Lock()
	player := c.Session.WS.PlayerList[name]
	c.Session.WS.PlayersL.Unlock()

	// todo: escape user input

	if player == nil {
		c.Session.Warnf("no such player as '%s' found.", name)
		return
	}

	if player.CurrentPhase != c.Session.CurrentPhase {
		c.Session.Warnf("'%s' is currently in phase %d. You must join this phase if you want to appear at this player's location.", name, player.CurrentPhase)
		return
	}

	targetMap := player.CurrentMap

	c.Session.TeleportTo(targetMap, player.Position())
}
