package commands

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/superp00t/gophercraft/packet/chat"
	"github.com/superp00t/gophercraft/realm"
	"github.com/superp00t/gophercraft/realm/wdb"
)

func cmdAddItem(s *realm.Session, id string, ct int) {
	var itemID string

	if ct == 0 {
		ct = 1
	}

	if len(id) == 0 {
		return
	}

	// ent:xxxx
	if strings.Count(id, ":") == 1 {
		itemID = id
	} else if id[0] == '|' {
		// extract item ID from link
		mk, err := chat.ParseMarkup(id)
		if err != nil {
			s.Warnf("error parsing link: %s", err)
			return
		}

		hlink, err := mk.ExtractHyperlinkData()
		if err != nil {
			s.Warnf("%s", err)
			return
		}

		if hlink.Type != "item" {
			s.Warnf("not an item link")
			return
		}

		u, err := strconv.ParseUint(hlink.Fields[0], 10, 32)
		if err != nil {
			s.Warnf("%s", err)
			return
		}

		var tpl *wdb.ItemTemplate
		s.DB().GetData(uint32(u), &tpl)
		if tpl != nil {
			err := s.AddItem(tpl.ID, ct, true, true)
			if err != nil {
				s.Warnf("%s", err)
			}
			return
		} else {
			s.Warnf("no such item as %d", u)
		}
	} else {
		u, err := strconv.ParseUint(id, 10, 32)
		if err != nil {
			s.Warnf("%s", err)
			return
		}

		itemID = fmt.Sprintf("it:%d", u)
	}

	err := s.AddItem(itemID, ct, true, true)
	if err != nil {
		s.Warnf("%s", err)
	}
}
