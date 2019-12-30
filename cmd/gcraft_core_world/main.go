package main

import (
	"log"

	"os"

	"github.com/superp00t/etc"
	"github.com/superp00t/etc/yo"
	"github.com/superp00t/gophercraft/gcore"
	"github.com/superp00t/gophercraft/gcore/config"

	_ "github.com/go-sql-driver/mysql"

	"github.com/superp00t/gophercraft/worldserver"
)

func main() {
	yo.Stringf("c", "config", "your realm configuration directory", etc.LocalDirectory().Concat("gcraft_world_1").Render())

	yo.Main("Gophercraft World Server", func(a []string) {
		gcore.PrintLicense()

		cpath := yo.StringG("c")
		if cpath == "" {
			cpath = etc.LocalDirectory().Concat("gcraft_world_1").Render()
		}

		if etc.ParseSystemPath(cpath).IsExtant() == false {
			yo.Println("No config file found at", cpath)
			os.Exit(0)
		}

		cfg, err := config.LoadWorld(cpath)
		if err != nil {
			yo.Fatal(err)
		}

		log.Fatal(worldserver.Start(cfg))
	})

	yo.Init()
}
