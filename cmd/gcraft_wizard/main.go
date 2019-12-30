package main

import (
	"fmt"

	"github.com/superp00t/etc"
	"github.com/superp00t/etc/yo"
	"github.com/superp00t/gophercraft/datapack"
	"github.com/superp00t/gophercraft/format/mpq"
	"github.com/superp00t/gophercraft/gcore/config"
	"github.com/superp00t/gophercraft/vsn"
)

var (
	pack    *datapack.Pack
	pool    *mpq.Pool
	version vsn.Build = 0
)

func main() {
	yo.Stringf("w", "gamepath", "game path", "")
	yo.Stringf("o", "outpath", "datapack output path", "")
	yo.Int64f("r", "realm", "the realm ID for generating configuration folders", 1)
	// yo.AddSubroutine("csv", nil, "debug csv", csvfn)
	yo.Main("server setup", _main)
	yo.Init()
}

func getAuth() *config.Auth {
	authLoc := etc.LocalDirectory().Concat("gcraft_auth")

	if !authLoc.IsExtant() {
		if err := config.GenerateDefaultAuth(authLoc.Render()); err != nil {
			yo.Fatal(err)
		}
	}

	a, err := config.LoadAuth(authLoc.Render())
	if err != nil {
		yo.Fatal(err)
	}

	return a
}

func getRealmID() uint64 {
	id := uint64(yo.Int64G("r"))
	return id
}

func _main(a []string) {
	auth := getAuth()

	gamePath := yo.StringG("w")

	if gamePath == "" {
		yo.Println("gcraft_wizard -w <game Data folder> [-o <world server conf folder>]")
		return
	}

	var err error
	version, err = vsn.DetectGame(gamePath)
	if err != nil {
		yo.Fatal(err)
	}

	worldfolder := yo.StringG("o")

	realmID := getRealmID()

	if worldfolder == "" {
		worldfolder = etc.LocalDirectory().Concat(fmt.Sprintf("gcraft_world_%d", realmID)).Render()
	}

	wf := etc.ParseSystemPath(worldfolder)

	if wf.IsExtant() {
		yo.Println("folder already exists. regenerating datapack.")
	} else {
		if err := auth.GenerateDefaultWorld(uint32(version), realmID, wf.Render()); err != nil {
			yo.Fatal(err)
		}
	}

	generateDatapack(gamePath, wf.Concat("datapacks", "!base.zip").Render())
}
