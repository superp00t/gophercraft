package datapack

import (
	"testing"

	"github.com/superp00t/etc"
	"github.com/superp00t/etc/yo"
	"github.com/superp00t/gophercraft/gcore"
)

func TestLoader(t *testing.T) {
	ld, err := Open(etc.Import("github.com/superp00t/gophercraft/datapack/testdata").Render())
	if err != nil {
		t.Fatal(err)
	}

	var pl []gcore.PortLocation
	ld.ReadAll("DB/PortLocation.csv", &pl)

	yo.Spew(pl)
	ld.Close()
}
