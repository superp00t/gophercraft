package commands

import (
	"math"

	"github.com/superp00t/gophercraft/realm"
)

func cmdXGPS(s *realm.Session, yards float32, direction string) {
	if direction == "" {
		direction = "f"
	}

	direction = direction[:1]

	if direction == "b" {
		direction = "f"
		yards = -yards
	}

	if direction == "u" {
		pos := s.Position()
		pos.Z = pos.Z + yards
		s.TeleportTo(s.CurrentMap, pos)
		return
	}

	if direction == "d" {
		pos := s.Position()
		pos.Z = pos.Z - yards
		s.TeleportTo(s.CurrentMap, pos)
		return
	}

	pos := s.Position()

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

	s.TeleportTo(s.CurrentMap, pos)
}
