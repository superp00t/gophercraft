package commands

import (
	"github.com/superp00t/gophercraft/realm"
	"github.com/superp00t/gophercraft/vsn"
)

func cmdFly(s *realm.Session, on bool) {
	if s.Build() < vsn.V2_4_3 {
		s.Warnf("Flight is not implemented in version %s.", s.Build())
		s.Warnf("Only lateral movement is allowed: You can use .xgps <distance> <up/down> to move vertically.")
	}

	s.Warnf("Flight activated: %v. To turn off: .gm fly off", on)

	s.SetFly(on)
}
