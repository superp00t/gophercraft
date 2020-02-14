package worldserver

import (
	"fmt"
	"reflect"

	"strconv"
	"strings"

	"github.com/schollz/progressbar"
	"github.com/superp00t/gophercraft/format/dbc"
	"github.com/superp00t/gophercraft/worldserver/wdb"
)

// everything must go!
// TODO: implement safer alternative
func (ws *WorldServer) deleteAllInTable(table interface{}) error {
	fmt.Println("Dropping table", reflect.TypeOf(table))
	if err := ws.DB.DropTables(table); err != nil {
		return err
	}
	return ws.DB.Sync2(table)
}

func (ws *WorldServer) LoadCSV(csvPath string, prototype interface{}) error {
	err := ws.deleteAllInTable(prototype)
	if err != nil {
		return err
	}

	typeOf := reflect.TypeOf(prototype).Elem()

	slice := reflect.New(reflect.SliceOf(typeOf))

	ws.PackLoader.ReadAll(csvPath, slice.Interface())

	bar := progressbar.NewOptions(slice.Elem().Len(), progressbar.OptionSetDescription("Installing "+typeOf.String()), progressbar.OptionEnableColorCodes(true))
	err = ws.bulkInsert(slice.Interface(), bar)
	fmt.Println()
	return err
}

// toggle the leftmost bit to 1.
// It seems unlikely that the MaNGOS item list will ever reach this super-high threshold.
// Thus, it makes an ideal starting namespace for custom items.
func getHighCounter() uint32 {
	// return uint32(0) | (1 << 31)
	return uint32(0) | (1 << 24)
}

func (ws *WorldServer) LoadItems() error {
	var itt []wdb.ItemTemplate
	ws.DB.Find(&itt)

	ws.deleteAllInTable(new(wdb.ItemTemplate))

	var items []wdb.ItemTemplate
	ws.PackLoader.ReadAll("DB/ItemTemplate.csv", &items)

	var highCounter uint32

	highCounter = getHighCounter()

	var startCounter = highCounter

	for _, e := range itt {
		if e.Entry > highCounter {
			startCounter = e.Entry + 1
		}
	}

	// Loop through items, applying the appropriate Entry code. This will be how the client identifies items, which may be custom.
itemLoop:
	for i := range items {
		v := &items[i]
		nspace := strings.Split(v.ID, ":")
		if len(nspace) != 2 {
			return fmt.Errorf("invalid id: %s", v.ID)
		}

		switch nspace[0] {
		// If "it", it means it's part of the MaNGOS item list. The entry has to be the same as the ID.
		case "it":
			u, err := strconv.ParseUint(nspace[1], 0, 32)
			if err != nil {
				return err
			}

			v.Entry = uint32(u)
		default:
			for _, e := range itt {
				if e.ID == v.ID {
					v.Entry = e.Entry
					continue itemLoop
				}
			}
			v.Entry = startCounter
			startCounter++
		}
	}

	bar := progressbar.NewOptions(len(items), progressbar.OptionSetDescription("Installing Items..."))
	err := ws.bulkInsert(&items, bar)
	fmt.Println()
	return err
}

func (ws *WorldServer) LoadGameObjects() error {
	var got []wdb.GameObjectTemplate
	ws.DB.Find(&got)

	ws.deleteAllInTable(new(wdb.GameObjectTemplate))

	var gobjs []wdb.GameObjectTemplate
	ws.PackLoader.ReadAll("DB/GameObjectTemplate.csv", &gobjs)

	var highCounter uint32
	highCounter = getHighCounter()
	var startCounter = highCounter

	for _, e := range got {
		if e.Entry > highCounter {
			startCounter = e.Entry + 1
		}
	}

gobjLoop:
	for i := range gobjs {
		v := &gobjs[i]
		nspace := strings.Split(v.ID, ":")
		if len(nspace) != 2 {
			return fmt.Errorf("invalid id: %s", v.ID)
		}

		switch nspace[0] {
		// If "go", it means it's part of the MaNGOS gameobject list. The entry has to be the same as the ID.
		case "go":
			u, err := strconv.ParseUint(nspace[1], 0, 32)
			if err != nil {
				return err
			}

			v.Entry = uint32(u)
		default:
			for _, e := range got {
				if e.ID == v.ID {
					v.Entry = e.Entry
					continue gobjLoop
				}
			}
			v.Entry = startCounter
			startCounter++
		}
	}

	bar := progressbar.NewOptions(len(gobjs), progressbar.OptionSetDescription("Installing GameObjects..."))
	err := ws.bulkInsert(&gobjs, bar)

	return err
}

func (ws *WorldServer) bulkInsert(v interface{}, pb *progressbar.ProgressBar) error {
	min := func(x, y int) int {
		if y < x {
			return y
		}
		return x
	}

	if pb != nil {
		pb.RenderBlank()
	}

	sliValue := reflect.ValueOf(v).Elem()

	if sliValue.Kind() != reflect.Slice {
		return fmt.Errorf("invalid kind")
	}

	offset := 512

	if sliValue.Len() <= offset {
		if pb != nil {
			pb.Add(sliValue.Len())
		}
		_, err := ws.DB.Insert(v)
		return err
	}

	// XORM can only put in so many records at once. we need to chunk it up, so that the operation can be successfully completed.
	for x := 0; x < sliValue.Len(); {
		lo := x
		increment := min(offset, sliValue.Len()-x)
		hi := x + increment

		// Equivalent of:
		// sli := new([]RecordType)
		// *sli = records[lo:hi]
		// DB.Insert(sli)
		sli := reflect.New(sliValue.Type())
		sli.Elem().Set(sliValue.Slice(lo, hi))
		// fmt.Println("new slice", sli.Elem().Len())
		_, err := ws.DB.Insert(sli.Interface())
		if err != nil {
			return err
		}
		x += increment
		if pb != nil {
			pb.Add(increment)
		}
	}

	return nil
}

func (ws *WorldServer) LoadDatapacks() error {
	for _, st := range []struct {
		Path      string
		Prototype interface{}
	}{
		{"DB/PortLocation.csv", new(wdb.PortLocation)},
		{"DB/DBC_CharStartOutfit.csv", new(dbc.Ent_CharStartOutfit)},
		{"DB/DBC_ChrRaces.csv", new(dbc.Ent_ChrRaces)},
		{"DB/DBC_ChrClasses.csv", new(dbc.Ent_ChrClasses)},
		{"DB/DBC_EmotesText.csv", new(dbc.Ent_EmotesText)},
		{"DB/DBC_AreaTrigger.csv", new(dbc.Ent_AreaTrigger)},
	} {
		if err := ws.LoadCSV(st.Path, st.Prototype); err != nil {
			return err
		}
	}

	if err := ws.LoadItems(); err != nil {
		return err
	}

	if err := ws.LoadGameObjects(); err != nil {
		return err
	}

	if err := ws.loadScripts(); err != nil {
		return err
	}

	return nil
}
