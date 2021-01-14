package main

import (
	"fmt"
	"net"
	"strings"

	"github.com/superp00t/etc"
)

func (as *authServer) serveAlpha(conn net.Conn) {
	rlm := as.Core.ListRealms()
	p := etc.NewBuffer()
	p.WriteByte(uint8(len(rlm)))

	for _, listing := range rlm {
		if listing.Version.RemovedIn(5875) {
			p.WriteCString(fmt.Sprintf("|cFF00FF00%s|r", listing.Name))
			var redirectServer = "0.0.0.0:9090"
			worldServer := strings.Split(listing.Address, ":")
			if len(worldServer) == 2 {
				redirectServer = worldServer[0] + ":9090"
			}
			p.WriteCString(redirectServer)
			p.WriteUint32(0)
		}
	}

	conn.Write(p.Bytes())
	conn.Close()
}
