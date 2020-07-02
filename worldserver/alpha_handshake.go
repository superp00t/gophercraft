package worldserver

import (
	"net"

	"github.com/superp00t/etc/yo"
	"github.com/superp00t/gophercraft/packet"
)

func (ws *WorldServer) handleAlpha(c net.Conn) {
	conn := packet.NewConnection(ws.Config.Version, c, nil, true)

	for {
		frame, err := conn.ReadFrame()
		if err != nil {
			yo.Warn(err)
			conn.Conn.Close()
			return
		}

		yo.Spew(frame)

		switch frame.Type {
		case packet.CMSG_PLAYED_TIME:
			response := packet.NewWorldPacket(packet.SMSG_PLAYED_TIME)
			response.WriteInt32(0)
			response.WriteInt32(0)
			conn.SendFrame(packet.Frame{
				Type: response.Type,
				Data: response.Bytes(),
			})
		case packet.CMSG_AUTH_SESSION:
			yo.Puke(frame.Data)
		}
	}
}
