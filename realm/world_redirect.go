package realm

import (
	"net"

	"github.com/superp00t/etc"

	"github.com/superp00t/etc/yo"
)

// Needed for the Alpha protocol.
func (ws *Server) serveRedirect() {
	const redirectAddress = "0.0.0.0:9090"

	srv, err := net.Listen("tcp", redirectAddress)
	if err != nil {
		yo.Fatal(err)
	}

	yo.Ok("Serving Alpha redirection server at", redirectAddress)

	for {
		conn, err := srv.Accept()
		if err != nil {
			yo.Fatal(err)
		}

		ws.sendRedirectAddress(conn)
	}
}

func (ws *Server) sendRedirectAddress(conn net.Conn) {
	redirectAddress := ws.Config.PublicAddress
	yo.Ok("Sending redirection server", redirectAddress)

	e := etc.NewBuffer()
	e.WriteCString(redirectAddress)

	conn.Write(e.Bytes())
	conn.Close()
}
