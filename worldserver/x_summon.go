package worldserver

import "time"

func x_Summon(c *C) {
	playername := c.String(0)

	c.Session.WS.PlayersL.Lock()
	defer c.Session.WS.PlayersL.Unlock()
	plyr := c.Session.WS.PlayerList[playername]
	if plyr == nil {
		c.Session.NoSuchPlayer(playername)
		return
	}

	if plyr.GUID() == c.Session.GUID() {
		c.Session.Warnf("You can't summon yourself!")
		return
	}

	plyr.SetSummonLocation(c.Session.CurrentPhase, c.Session.CurrentMap, c.Session.Position())
	plyr.SendSummonRequest(c.Session.GUID(), c.Session.ZoneID, 2*time.Minute)
}
