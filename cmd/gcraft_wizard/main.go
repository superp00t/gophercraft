package main

import (
	"github.com/superp00t/etc"
	"github.com/superp00t/etc/yo"
	"github.com/superp00t/gophercraft/gcore/config"
	"github.com/superp00t/gophercraft/vsn"
	
)

func main() {
	vsn.PrintBanner()
	wizStart()
}

func getLocal() etc.Path {
	loc := etc.LocalDirectory()
	return loc
}

func getAuth() *config.Auth {
	authLoc := getLocal().Concat("gcraft_auth")

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
