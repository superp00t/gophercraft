package worldserver

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/superp00t/gophercraft/worldserver/wdb"

	"github.com/superp00t/etc/yo"
	"github.com/superp00t/gophercraft/packet/chat"
)

func x_AddItem(c *C) {
	var itemID string
	id := c.String(0)
	ct := c.Int(1)

	if ct == 0 {
		ct = 1
	}

	if len(id) == 0 {
		yo.Spew(c.Args)
		return
	}

	// ent:xxxx
	if strings.Count(id, ":") == 1 {
		itemID = id
	} else if id[0] == '|' {
		// extract item ID from link
		mk, err := chat.ParseMarkup(id)
		if err != nil {
			c.Session.Warnf("error parsing link: %s", err)
			return
		}

		hlink, err := mk.ExtractHyperlinkData()
		if err != nil {
			c.Session.Warnf("%s", err)
			return
		}

		if hlink.Type != "item" {
			c.Session.Warnf("not an item link")
			return
		}

		u, err := strconv.ParseUint(hlink.Fields[0], 10, 32)
		if err != nil {
			c.Session.Warnf("%s", err)
			return
		}

		var tpl *wdb.ItemTemplate
		wdb.GetData(uint32(u), &tpl)
		if tpl != nil {
			err := c.Session.AddItem(tpl.ID, ct, true, true)
			if err != nil {
				c.Session.Warnf("%s", err)
			}
			return
		} else {
			c.Session.Warnf("no such item as %d", u)
		}
	} else {
		u, err := strconv.ParseUint(id, 10, 32)
		if err != nil {
			c.Session.Warnf("%s", err)
			return
		}

		itemID = fmt.Sprintf("it:%d", u)
	}

	err := c.Session.AddItem(itemID, ct, true, true)
	if err != nil {
		c.Session.Warnf("%s", err)
	}
}
