package datapack

import (
	"testing"

	"github.com/superp00t/etc"
	"github.com/superp00t/etc/yo"
	"github.com/superp00t/gophercraft/worldserver/wdb"
)

type celebRecord struct {
	Name     string
	KnownFor []string
}

func TestLoader(t *testing.T) {
	ld, err := Open(etc.Import("github.com/superp00t/gophercraft/datapack/testdata").Render())
	if err != nil {
		t.Fatal(err)
	}

	var pl []wdb.PortLocation
	ld.ReadAll("DB/PortLocation.csv", &pl)

	yo.Spew(pl)
	ld.Close()
}
