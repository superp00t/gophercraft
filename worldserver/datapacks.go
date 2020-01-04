package worldserver

import (
	"fmt"
	"reflect"

	"github.com/superp00t/etc/yo"
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

	yo.Ok("Loading", csvPath)

	slice := reflect.New(reflect.SliceOf(typeOf))

	ws.PackLoader.ReadAll(csvPath, slice.Interface())

	yo.Ok("Loaded", csvPath, "successfully,", slice.Elem().Len(), "records read")

	_, err = ws.DB.Insert(slice.Interface())
	return err
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
	} {
		if err := ws.LoadCSV(st.Path, st.Prototype); err != nil {
			return err
		}
	}

	return nil
}
