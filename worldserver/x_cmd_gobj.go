package worldserver

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/superp00t/gophercraft/packet/chat"
)

func x_Gobject_spawn(c *C) {
	var gobID string

	arg := c.String(0)

	if strings.Contains(arg, ":") {
		if strings.Contains(arg, "|") {
			link, err := chat.ParseMarkup(arg)
			if err != nil {
				c.Session.Warnf("%s", err)
				return
			}

			hlink, err := link.ExtractHyperlinkData()
			if err != nil {
				c.Session.Warnf("%s", err)
				return
			}

			if hlink.Type != "gameobject_entry" || len(hlink.Fields) == 0 {
				c.Session.Warnf("wrong link type")
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
			c.Session.Warnf("%s", err)
			return
		}

		gobID = fmt.Sprintf("go:%d", u)
	}

	err := c.Session.Map().SpawnGameObject(gobID, c.Session.Position())
	if err != nil {
		c.Session.Warnf("%s", err)
	}
}
