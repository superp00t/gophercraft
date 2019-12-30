package worldserver

import (
	"github.com/superp00t/gophercraft/packet/update"
	"github.com/superp00t/gophercraft/worldserver/wdb"
)

type Item struct {
	wdb.Item
	*update.ValuesBlock
}

// func (s *Session) NewItem(it gcore.Item) *Item {
// 	i := &Item{it, update.NewValuesBlock()}
// 	g := guid.RealmSpecific(guid.Item, s.WS.RealmID, i.ID)
// 	i.SetGUIDValue(update.ObjectGUID, g)
// 	i.SetUint32Value(update.ObjectEntry, i.ItemID)
// 	i.SetTypeMask(s.Version(), guid.TypeMaskObject|guid.TypeMaskItem)
// 	i.SetFloat32Value(update.ObjectScaleX, 1.0)
// 	i.SetUint32Value(update.Item)
// 	return i
// }

// func (s *Session) InitInventoryManager() {
// 	var inv []gcore.Item
// 	s.WS.DB.Where("owner = ?", s.GUID().Counter()).Find(&inv)
// }
