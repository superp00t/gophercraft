package commands

import (
	"fmt"
	"time"

	"github.com/superp00t/gophercraft/realm"
	"github.com/superp00t/gophercraft/realm/wdb"
)

// this file contains various .lookup commands.

func cmdLookupTeleport(s *realm.Session, portLoc string) {
	if portLoc == "" {
		return
	}

	max := int64(75)

	var locations []*wdb.PortLocation

	err := s.DB().SearchTemplates(portLoc, max, &locations)
	if err != nil {
		s.Warnf("%s", err)
		return
	}

	for _, loc := range locations {
		s.SystemChat(fmt.Sprintf("|cFFFFFFFF[%s]|r", loc.ID))
	}

	s.Warnf("%d/%d port locations returned.", len(locations), max)
}

func cmdLookupItem(s *realm.Session, itemName string) {
	if itemName == "" {
		return
	}

	max := int64(75)

	ln := 0

	now := time.Now()

	var items []*wdb.ItemTemplate
	if err := s.DB().SearchTemplates(itemName, max, &items); err != nil {
		s.Warnf("%s", err)
		return
	}

	for _, v := range items {
		s.SystemChat(fmt.Sprintf("%s (%d) - |cffffffff|Hitem:%d::::::::%d::::|h[%s]|h|r", v.ID, v.Entry, v.Entry, s.GetLevel(), v.Name))
		ln++
	}

	elapsed := time.Since(now)

	s.Warnf("%d items returned in %v. (maximum query: %d)", ln, elapsed, max)
}

func cmdLookupGameObject(s *realm.Session, gobjName string) {
	if gobjName == "" {
		return
	}

	max := int64(75)

	ln := 0

	now := time.Now()

	var gobj []*wdb.GameObjectTemplate

	err := s.DB().SearchTemplates(gobjName, max, &gobj)
	if err != nil {
		s.Warnf("%s", err)
		return
	}

	for _, v := range gobj {
		s.SystemChat(fmt.Sprintf("%d - |cffffffff|Hgameobject_entry:%d|h[%s]|h|r", v.Entry, v.Entry, v.Name))
		ln++
	}

	elapsed := time.Since(now)

	s.Warnf("%d GameObjects returned in %v. (maximum query: %d)", ln, elapsed, max)
}
