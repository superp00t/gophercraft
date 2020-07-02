package main

import (
	"log"

	"github.com/superp00t/gophercraft/gcore/sys"
	"github.com/superp00t/gophercraft/vsn"

	"os"

	"github.com/superp00t/etc"
	"github.com/superp00t/etc/yo"
	"github.com/superp00t/gophercraft/gcore/config"

	_ "github.com/go-sql-driver/mysql"

	"github.com/superp00t/gophercraft/worldserver"
)

func main() {
	yo.Stringf("c", "config", "your realm configuration directory", etc.LocalDirectory().Concat("gcraft_world_1").Render())

	yo.Main("Gophercraft World Server", func(a []string) {
		vsn.PrintBanner()

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

		fp, _ := sys.GetCertFileFingerprint(cfg.Path.Concat("cert.pem").Render())

		yo.Ok("This server's fingerprint is", fp)

		log.Fatal(worldserver.Start(cfg))
	})

	yo.Init()
}
