package worldserver

import (
	"reflect"

	"strconv"
	"strings"

	"github.com/schollz/progressbar"
	"github.com/superp00t/gophercraft/format/dbc"
	"github.com/superp00t/gophercraft/guid"
	"github.com/superp00t/gophercraft/worldserver/wdb"
)

// It seems unlikely that the main item list will ever reach this super-high threshold.
// Thus, it makes an ideal starting namespace for custom items.
func getHighCounter() uint32 {
	// return uint32(0) | (1 << 31)
	return uint32(0) | (1 << 24)
}

// LoadObjectTemplates (item/creature/gameobject templates)
func (ws *WorldServer) LoadObjectTemplates(path string, typeID guid.TypeID, templates interface{}) error {
	// offset counter to largest in registry. This prevents collisions from occuring.
	var highCounter uint32
	highCounter = getHighCounter()
	var startCounter = highCounter

	// The server keeps a list of ID-entry associations (ObjectTemplateRegistry)
	// This allows you to repeatedly add new custom object templates without always needing to refresh the cache.
	var knownIDs []wdb.ObjectTemplateRegistry
	var newIDs []wdb.ObjectTemplateRegistry
	ws.DB.Where("type = ?", typeID).Find(&knownIDs)
	var cachedIDs = map[string]uint32{}

	for _, id := range knownIDs {
		cachedIDs[id.ID] = id.Entry

		// offset counter to largest in registry. This prevents collisions from occuring.
		if id.Entry >= startCounter {
			startCounter = id.Entry + 1
		}
	}

	ws.PackLoader.ReadAll(path, templates)

	slice := reflect.ValueOf(templates).Elem()

	wdb.ClearData(slice.Type())

	bar := progressbar.NewOptions(slice.Len(), progressbar.OptionSetDescription("Loading "+slice.Type().String()+"..."))

	for x := 0; x < slice.Len(); x++ {
		object := slice.Index(x)
		id := wdb.GetID(object)
		str := strings.Split(id, ":")
		var entry uint32
		if len(str) >= 2 {
			num, err := strconv.ParseUint(str[1], 10, 32)
			if err == nil {
				entry = uint32(num)
			} else {
				entry = startCounter
				startCounter++
				newIDs = append(newIDs, wdb.ObjectTemplateRegistry{
					ID:    id,
					Type:  typeID,
					Entry: entry,
				})
			}
		} else {
			entry = startCounter
			startCounter++
			newIDs = append(newIDs, wdb.ObjectTemplateRegistry{
				ID:    id,
				Type:  typeID,
				Entry: entry,
			})
		}

		wdb.SetEntry(object, entry)
		wdb.StoreData(object.Addr())
		bar.Add(1)
	}

	bar.Finish()

	_, err := ws.DB.Insert(&newIDs)
	return err
}

func (ws *WorldServer) LoadStaticFields(path string, items interface{}) {
	ws.PackLoader.ReadAll(path, items)

	wdb.ClearData(reflect.ValueOf(items).Elem().Type())

	fields := reflect.ValueOf(items)

	bar := progressbar.NewOptions(fields.Elem().Len(), progressbar.OptionSetDescription("Loading "+fields.Elem().Type().String()+"..."))

	for x := 0; x < fields.Elem().Len(); x++ {
		wdb.StoreData(fields.Elem().Index(x).Addr())
		bar.Add(1)
	}

	bar.Finish()
}

func (ws *WorldServer) LoadDatapacks() error {
	var portLocations []wdb.PortLocation
	ws.LoadStaticFields("DB/PortLocation.txt", &portLocations)

	var objs []wdb.GameObjectTemplate
	ws.LoadObjectTemplates("DB/GameObjectTemplate.txt", guid.TypeGameObject, &objs)

	var items []wdb.ItemTemplate
	ws.LoadObjectTemplates("DB/ItemTemplate.txt", guid.TypeItem, &items)

	var startOutfits []dbc.Ent_CharStartOutfit
	ws.LoadStaticFields("DB/DBC_CharStartOutfit.txt", &startOutfits)

	var races []dbc.Ent_ChrRaces
	ws.LoadStaticFields("DB/DBC_ChrRaces.txt", &races)

	var classes []dbc.Ent_ChrClasses
	ws.LoadStaticFields("DB/DBC_ChrClasses.txt", &classes)

	var emotes []dbc.Ent_EmotesText
	ws.LoadStaticFields("DB/DBC_EmotesText.txt", &emotes)

	var areaTriggers []dbc.Ent_AreaTrigger
	ws.LoadStaticFields("DB/DBC_AreaTrigger.txt", &areaTriggers)

	return ws.loadScripts()
}
