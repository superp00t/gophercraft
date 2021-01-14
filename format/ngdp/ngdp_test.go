package ngdp

import (
	"testing"

	"github.com/superp00t/etc/yo"
)

func TestNGDP(t *testing.T) {
	ag := DefaultAgent()
	ag.HostServer = "http://us.patch.battle.net:1119"

	online, err := ag.OpenOnline("wow")
	if err != nil {
		t.Fatal(err)
	}
	yo.Spew(online.BuildConfig)
}
