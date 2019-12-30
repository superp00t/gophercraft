package main

import (
	"log"

	"github.com/davecgh/go-spew/spew"
	"github.com/superp00t/etc/yo"
	"github.com/superp00t/gophercraft/client"
	"github.com/superp00t/gophercraft/packet"
)

func main() {
	yo.Stringf("u", "username", "username", "username")
	yo.Stringf("p", "password", "password", "password")
	yo.Stringf("a", "authserver", "the auth server", "localhost:3724")
	yo.Stringf("r", "realm", "realm", "demo")
	yo.Stringf("n", "playername", "your character's name", "demo")
	yo.Int64f("b", "build", "the build which you wish to simulate", 5875)
	yo.Main("the bot", _main)
	yo.Init()
}

func _main(args []string) {
	spew.Config.SortKeys = true
	spew.Config.SpewKeys = true

	c, err := client.New(&client.Config{
		Username:   yo.StringG("u"),
		Password:   yo.StringG("p"),
		AuthServer: yo.StringG("a"),
		Playername: yo.StringG("n"),
		Build:      uint32(yo.Int64G("b")),
	})

	if err != nil {
		yo.Fatal(err)
	}

	yo.Spew(c.RealmList)

	for _, v := range c.RealmList.Realms {
		if v.Name == yo.StringG("r") {
			connect(c, v)
			return
		}
	}
}

func connect(c *client.Client, v packet.RealmListing) {
	err := c.WorldConnect(v.Address)
	if err != nil {
		log.Fatal(err)
	}
}
