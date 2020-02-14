package worldserver

import (
	"fmt"
	"sort"
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

	ln := 0

	err := like(c.Session.DB().NewSession(), "port_id", portLoc).Limit(int(max)).Desc("port_id").Iterate(new(wdb.PortLocation), func(i int, bean interface{}) error {
		v := bean.(*wdb.PortLocation)
		c.Session.SystemChat(fmt.Sprintf("|cFFFFFFFF[%s]|r", v.Name))
		ln++
		return nil
	})
	if err != nil {
		panic(err)
	}

	c.Session.Warnf("%d/%d port locations returned.", ln, max)
}

type itemTemplateResult []wdb.ItemTemplate

func (itr itemTemplateResult) Len() int {
	return len(itr)
}

func (itr itemTemplateResult) Swap(i, j int) {
	_j := itr[j]
	_i := itr[i]
	itr[i] = _j
	itr[j] = _i
}

func (itr itemTemplateResult) Less(i, j int) bool {
	return itr[i].Entry < itr[j].Entry
}

func x_LookupItem(c *C) {
	itemName := c.ArgString()

	if itemName == "" {
		return
	}

	max := int64(75)

	ln := 0

	now := time.Now()

	var itr itemTemplateResult

	err := like(c.Session.DB().NewSession().Cols("name", "entry"), "name", itemName).Limit(int(max)).Find(&itr)
	if err != nil {
		panic(err)
	}

	sort.Sort(itr)
	for _, v := range itr {
		c.Session.SystemChat(fmt.Sprintf("%d - |cffffffff|Hitem:%d::::::::%d::::|h[%s]|h|r", v.Entry, v.Entry, c.Session.GetLevel(), v.Name))
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

	var gobj []wdb.GameObjectTemplate

	err := like(c.Session.DB().NewSession().Cols("name", "entry"), "name", gobjName).Limit(int(max)).Find(&gobj)
	if err != nil {
		panic(err)
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
