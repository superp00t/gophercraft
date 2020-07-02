package worldserver

import (
	"fmt"
	"time"

	"github.com/superp00t/etc/yo"
	"github.com/superp00t/gophercraft/worldserver/wdb"
	"xorm.io/xorm"
)

// this file contains various .lookup commands.

func x_LookupTeleport(c *C) {
	portLoc := c.String(0)

	yo.Spew(c.Args)

	if portLoc == "" {
		return
	}

	fmt.Println("searchin string", portLoc)

	max := int64(75)

	var locations []*wdb.PortLocation

	err := wdb.SearchTemplates(portLoc, max, &locations)
	if err != nil {
		c.Session.Warnf("%s", err)
		return
	}

	for _, loc := range locations {
		c.Session.SystemChat(fmt.Sprintf("|cFFFFFFFF[%s]|r", loc.ID))
	}

	c.Session.Warnf("%d/%d port locations returned.", len(locations), max)
}

func x_LookupItem(c *C) {
	itemName := c.ArgString()

	if itemName == "" {
		return
	}

	max := int64(75)

	ln := 0

	now := time.Now()

	var items []*wdb.ItemTemplate
	if err := wdb.SearchTemplates(itemName, max, &items); err != nil {
		c.Session.Warnf("%s", err)
		return
	}

	for _, v := range items {
		c.Session.SystemChat(fmt.Sprintf("%s (%d) - |cffffffff|Hitem:%d::::::::%d::::|h[%s]|h|r", v.ID, v.Entry, v.Entry, c.Session.GetLevel(), v.Name))
		ln++
	}

	elapsed := time.Since(now)

	c.Session.Warnf("%d items returned in %v. (maximum query: %d)", ln, elapsed, max)
}

func x_LookupGameObject(c *C) {
	gobjName := c.ArgString()
	if gobjName == "" {
		return
	}

	max := int64(75)

	ln := 0

	now := time.Now()

	var gobj []*wdb.GameObjectTemplate

	err := wdb.SearchTemplates(gobjName, max, &gobj)
	if err != nil {
		c.Session.Warnf("%s", err)
		return
	}

	for _, v := range gobj {
		c.Session.SystemChat(fmt.Sprintf("%d - |cffffffff|Hgameobject_entry:%d|h[%s]|h|r", v.Entry, v.Entry, v.Name))
		ln++
	}

	elapsed := time.Since(now)

	c.Session.Warnf("%d GameObjects returned in %v. (maximum query: %d)", ln, elapsed, max)
}

func like(s *xorm.Session, columnName string, searchName string) *xorm.Session {
	return s.Where(columnName+" regexp ?", searchName)
}
