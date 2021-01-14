package realm

import (
	"reflect"

	"strconv"
	"strings"

	"github.com/schollz/progressbar"
	"github.com/superp00t/gophercraft/format/dbc"
	"github.com/superp00t/gophercraft/guid"
	"github.com/superp00t/gophercraft/realm/wdb"
)

// It seems unlikely that the files will ever reach this super-high threshold.
// Thus, it makes an ideal starting point for custom templates.
func getHighCounter() uint32 {
	// begins at 0x1000000 (16777216)
	return uint32(0) | (1 << 24)
}

// LoadObjectTemplates (item/creature/gameobject templates)
func (ws *Server) LoadObjectTemplates(path string, typeID guid.TypeID, templates interface{}) error {
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

	ws.DB.ClearData(slice.Type())

	bar := progressbar.NewOptions(slice.Len(), progressbar.OptionSetDescription("Loading "+slice.Type().String()+"..."))

	for x := 0; x < slice.Len(); x++ {
		object := slice.Index(x)
		id := wdb.GetID(object)
		str := strings.Split(id, ":")
		var entry uint32
		if len(str) >= 2 {
			num, err := strconv.ParseUint(str[1], 10, 32)
			if err == nil {
				// If this ID contains a number, use this as the entry code.
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
		ws.DB.StoreData(object.Addr())
		bar.Add(1)
	}

	bar.Finish()

	_, err := ws.DB.Insert(&newIDs)
	return err
}

func (ws *Server) LoadStaticFields(path string, items interface{}) {
	ws.PackLoader.ReadAll(path, items)

	ws.DB.ClearData(reflect.ValueOf(items).Elem().Type())

	fields := reflect.ValueOf(items)

	bar := progressbar.NewOptions(fields.Elem().Len(), progressbar.OptionSetDescription("Loading "+fields.Elem().Type().String()+"..."))

	for x := 0; x < fields.Elem().Len(); x++ {
		ws.DB.StoreData(fields.Elem().Index(x).Addr())
		bar.Add(1)
	}

	bar.Finish()
}

func (ws *Server) LoadDatapacks() error {
	var portLocations []wdb.PortLocation
	ws.LoadStaticFields("DB/PortLocation.txt", &portLocations)

	var objs []wdb.GameObjectTemplate
	ws.LoadObjectTemplates("DB/GameObjectTemplate.txt", guid.TypeGameObject, &objs)

	var items []wdb.ItemTemplate
	ws.LoadObjectTemplates("DB/ItemTemplate.txt", guid.TypeItem, &items)

	var texts []wdb.NPCText
	ws.LoadObjectTemplates("DB/NPCText.txt", guid.TypeNPCText, &texts)

	var creatures []wdb.CreatureTemplate
	ws.LoadObjectTemplates("DB/CreatureTemplate.txt", guid.TypeUnit, &creatures)

	var loctext []wdb.LocString
	ws.LoadStaticFields("DB/LocString.txt", &loctext)

	var levelxp []wdb.LevelExperience
	ws.PackLoader.ReadAll("DB/LevelExperience.txt", &levelxp)

	ws.LevelExperience = make(wdb.LevelExperience)
	for _, lexp := range levelxp {
		for level, exp := range lexp {
			ws.LevelExperience[level] = exp
		}
	}

	ws.PackLoader.ReadAll("DB/PlayerCreateInfo.txt", &ws.PlayerCreateInfo)
	ws.PackLoader.ReadAll("DB/PlayerCreateAbility.txt", &ws.PlayerCreateAbilities)
	ws.PackLoader.ReadAll("DB/PlayerCreateItem.txt", &ws.PlayerCreateItems)
	ws.PackLoader.ReadAll("DB/PlayerCreateActionButton.txt", &ws.PlayerCreateActionButtons)

	var maps []wdb.Map
	ws.LoadStaticFields("DB/Map.txt", &maps)

	var zones []dbc.Ent_AreaTable
	ws.LoadStaticFields("DB/DBC_AreaTable.txt", &zones)

	var startOutfits []dbc.Ent_CharStartOutfit
	ws.LoadStaticFields("DB/DBC_CharStartOutfit.txt", &startOutfits)

	var races []dbc.Ent_ChrRaces
	ws.LoadStaticFields("DB/DBC_ChrRaces.txt", &races)

	var classes []dbc.Ent_ChrClasses
	ws.LoadStaticFields("DB/DBC_ChrClasses.txt", &classes)

	var creatureFamilies []dbc.Ent_CreatureFamily
	ws.LoadStaticFields("DB/DBC_CreatureFamily.txt", &creatureFamilies)

	var emotes []dbc.Ent_EmotesText
	ws.LoadStaticFields("DB/DBC_EmotesText.txt", &emotes)

	var areaTriggers []dbc.Ent_AreaTrigger
	ws.LoadStaticFields("DB/DBC_AreaTrigger.txt", &areaTriggers)

	return nil
}
