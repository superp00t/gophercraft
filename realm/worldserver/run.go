//Package worldserver is the main ent
package worldserver

import (
	"fmt"

	"os"

	"github.com/superp00t/etc"
	"github.com/superp00t/etc/yo"
	"github.com/superp00t/gophercraft/gcore/config"
	"github.com/superp00t/gophercraft/gcore/sys"
	"github.com/superp00t/gophercraft/realm"
	"github.com/superp00t/gophercraft/vsn"

	_ "github.com/go-sql-driver/mysql"
	// Imports core plugins, needed to avoid an import cycle.
	_ "github.com/superp00t/gophercraft/realm/plugins/commands"
	_ "github.com/superp00t/gophercraft/realm/plugins/discord"
)

func Run() {
	vsn.PrintBanner()

	var configPath etc.Path

	if len(os.Args) > 1 {
		configPath = etc.ParseSystemPath(os.Args[1])
		if configPath.IsExtant() == false {
			localConfig := etc.LocalDirectory().Concat("Gophercraft").Concat(os.Args[1])
			if localConfig.IsExtant() == false {
				fmt.Println("Fatal error: could not find world config")
				fmt.Println("  at:", configPath)
				fmt.Println("  or:", localConfig)
			} else {
				configPath = localConfig
			}
		}
	} else {
		fmt.Println(os.Args[0], "<world config path>")
		os.Exit(0)
	}

	cfg, err := config.LoadWorld(configPath.Render())
	if err != nil {
		yo.Fatal(err)
	}

	fp, _ := sys.GetCertFileFingerprint(cfg.Path.Concat("cert.pem").Render())

	yo.Ok("This server's fingerprint is", fp)

	yo.Fatal(realm.Start(cfg))
}
