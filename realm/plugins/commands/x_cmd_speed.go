package commands

import (
	"github.com/superp00t/gophercraft/realm"
)

func cmdSpeed(s *realm.Session, speed float32) {
	if speed < .1 || speed > 100 {
		s.Warnf("speed must be [0.1 - 50.0]")
		return
	}

	if speed == 0 {
		speed = 1
	}

	s.ChangeDefaultSpeeds(speed)
	s.SyncSpeeds()
}
