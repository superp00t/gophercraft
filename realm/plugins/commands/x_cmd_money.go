package commands

import (
	"github.com/superp00t/gophercraft/econ"
	"github.com/superp00t/gophercraft/realm"
)

func cmdMoney(s *realm.Session, add econ.Money) {
	s.AddMoney(add)
	s.Warnf("Added %s: new balance: %s", add, s.Char.Coinage)
}
