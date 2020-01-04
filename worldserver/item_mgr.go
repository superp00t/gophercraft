package worldserver

import (
	"github.com/superp00t/gophercraft/packet/update"
)

type Item struct {
	*update.ValuesBlock
}

// func (s *Session) NewItem(it wdb.Item) *Item {
// 	i := &Item{update.NewValuesBlock()}
// 	g := guid.RealmSpecific(guid.Item, s.WS.RealmID, it.ID)
// 	i.SetGUIDValue(update.ObjectGUID, g)
// 	i.SetUint32Value(update.ObjectEntry, i.ItemID)
// 	i.SetTypeMask(s.Version(), guid.TypeMaskObject|guid.TypeMaskItem)
// 	i.SetFloat32Value(update.ObjectScaleX, 1.0)
// 	i.SetGUIDValue(it.)
// 	return i
// }

// func (s *Session) InitInventoryManager() {
// 	var inv []wdb.Item
// 	s.WS.DB.Where("owner = ?", s.GUID().Counter()).Find(&inv)

// 	changes := make(map[update.Global]interface{})
// 	arrayData := &update.ArrayData{
// 		Cols: []string{"Creator", "Entry", "Enchantments", "Properties"},
// 	}

// 	for _, v := range Ü

// }
