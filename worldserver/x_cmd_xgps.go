package worldserver

import "math"

func x_XGPS(c *C) {
	yards := c.Float32(0)
	direction := c.String(1)

	if direction == "" {
		direction = "f"
	}

	direction = direction[:1]

	if direction == "b" {
		direction = "f"
		yards = -yards
	}

	if direction == "u" {
		pos := c.Session.Position()
		pos.Z = pos.Z + yards
		c.Session.TeleportTo(c.Session.CurrentMap, pos)
		return
	}

	if direction == "d" {
		pos := c.Session.Position()
		pos.Z = pos.Z - yards
		c.Session.TeleportTo(c.Session.CurrentMap, pos)
		return
	}

	pos := c.Session.Position()

	projection := pos.O

	// 90 degrees in Radians.
	r90 := float32(1.5708)

	// turn projection 90 to the left.
	if direction == "l" {
		projection = pos.O + r90
	}

	// turn projection 90 to the right
	if direction == "r" {
		projection = pos.O - r90
	}

	pos.X = pos.X + yards*float32(math.Cos(float64(projection)))
	pos.Y = pos.Y + yards*float32(math.Sin(float64(projection)))

	c.Session.TeleportTo(c.Session.CurrentMap, pos)
}
