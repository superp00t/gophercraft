package worldserver

func x_Sound(c *C) {
	// port string
	if len(c.Args) < 1 {
		c.Session.Annf(".sound <soundID>")
		return
	}

	c.Session.Map().PlaySound(c.Uint32(0))
}
