package commands

import "github.com/superp00t/gophercraft/realm"

func cmdModLevel(s *realm.Session, level int) {
	s.LevelUp(uint32(level))
}
