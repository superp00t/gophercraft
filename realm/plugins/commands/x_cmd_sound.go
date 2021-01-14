package commands

import "github.com/superp00t/gophercraft/realm"

func cmdSound(s *realm.Session, sound uint32) {
	s.Map().PlaySound(sound)
}

func cmdMusic(s *realm.Session, music uint32) {
	s.Map().PlayMusic(music)
}
