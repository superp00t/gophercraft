package wdb

import (
	"github.com/go-xorm/xorm"
	"github.com/superp00t/gophercraft/format/dbc"
	_ "github.com/superp00t/gophercraft/gcore/dbsupport"
)

type Core struct {
	*xorm.Engine
}

func NewCore(driver, source string) (*Core, error) {
	var err error
	cn := new(Core)
	cn.Engine, err = xorm.NewEngine(driver, source)
	if err != nil {
		return nil, err
	}

	err = cn.Engine.Sync2(
		new(PortLocation),
		new(Character),
		new(Item),
		new(dbc.Ent_CharStartOutfit),
		new(dbc.Ent_EmotesText),
		new(dbc.Ent_Emotes),
		new(dbc.Ent_ChrRaces),
		new(dbc.Ent_ChrClasses),
	)

	if err != nil {
		return nil, err
	}

	_, err = cn.Count(new(Character))
	if err != nil {
		return nil, err
	}

	return cn, nil
}
