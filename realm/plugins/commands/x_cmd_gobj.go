package commands

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/superp00t/gophercraft/packet/chat"
	"github.com/superp00t/gophercraft/realm"
)

func cmdSpawnGameobject(s *realm.Session, arg string) {
	var gobID string

	if strings.Contains(arg, ":") {
		if strings.Contains(arg, "|") {
			link, err := chat.ParseMarkup(arg)
			if err != nil {
				s.Warnf("%s", err)
				return
			}

			hlink, err := link.ExtractHyperlinkData()
			if err != nil {
				s.Warnf("%s", err)
				return
			}

			if hlink.Type != "gameobject_entry" || len(hlink.Fields) == 0 {
				s.Warnf("wrong link type")
				return
			}

			gob := hlink.Fields[0]
			gobID = "go:" + gob
		} else {
			gobID = arg
		}
	} else {
		// numerical
		u, err := strconv.ParseUint(arg, 10, 32)
		if err != nil {
			s.Warnf("%s", err)
			return
		}

		gobID = fmt.Sprintf("go:%d", u)
	}

	err := s.Map().SpawnGameObject(gobID, s.Position())
	if err != nil {
		s.Warnf("%s", err)
	}
}
