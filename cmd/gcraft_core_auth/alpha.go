package main

import (
	"fmt"
	"net"
	"strings"

	"github.com/superp00t/etc"
	"github.com/superp00t/gophercraft/vsn"
)

func (as *authServer) serveAlpha(conn net.Conn) {
	rlm := as.Core.ListRealms("", vsn.Alpha)
	p := etc.NewBuffer()
	p.WriteByte(uint8(len(rlm)))

	for _, listing := range rlm {
		p.WriteCString(fmt.Sprintf("|cFF00FFFF%s", listing.Name))
		var redirectServer = "0.0.0.0:9090"
		worldServer := strings.Split(listing.Address, ":")
		if len(worldServer) == 2 {
			redirectServer = worldServer[0] + ":9090"
		}
		fmt.Println("Redirection Server", redirectServer)
		p.WriteCString(redirectServer)
		p.WriteUint32(0)
	}

	conn.Write(p.Bytes())
	conn.Close()
}
