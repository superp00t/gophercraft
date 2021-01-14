package realm

import (
	"context"
	"net"
	"strings"

	"github.com/superp00t/etc"

	"github.com/superp00t/etc/yo"
	"github.com/superp00t/gophercraft/gcore/sys"
	"github.com/superp00t/gophercraft/i18n"
	"github.com/superp00t/gophercraft/packet"
	"github.com/superp00t/gophercraft/vsn"
)

func (ws *Server) handleAlpha(c net.Conn) {
	conn := packet.NewConnection(ws.Config.Version, c, nil, true)

	for {
		frame, err := conn.ReadFrame()
		if err != nil {
			yo.Warn(err)
			conn.Conn.Close()
			return
		}

		switch frame.Type {
		case packet.CMSG_AUTH_SESSION:
			e := etc.FromBytes(frame.Data)
			build := vsn.Build(e.ReadUint32())
			if build != ws.Config.Version {
				wp := packet.NewWorldPacket(packet.SMSG_AUTH_RESPONSE)
				wp.WriteByte(packet.AUTH_VERSION_MISMATCH)
				conn.SendFrame(wp.Frame())
				return
			}

			_ = e.ReadUint32()

			authFile := e.ReadCString()
			authFile = strings.ReplaceAll(authFile, "\r", "")
			authParts := strings.Split(authFile, "\n")
			if len(authParts) != 2 {
				wp := packet.NewWorldPacket(packet.SMSG_AUTH_RESPONSE)
				wp.WriteByte(packet.AUTH_REJECT)
				conn.SendFrame(wp.Frame())
				return
			}

			user := authParts[0]
			passHash := authParts[1]

			yo.Ok(user, passHash)

			resp, err := ws.AuthServiceClient.VerifyWorld(context.Background(), &sys.VerifyWorldQuery{
				RealmID:     ws.Config.RealmID,
				Build:       uint32(build),
				Account:     user,
				GameAccount: "Zero",
				IP:          c.RemoteAddr().String(),
				Digest:      []byte(passHash),
			})

			if err != nil {
				yo.Warn(err)
				wp := packet.NewWorldPacket(packet.SMSG_AUTH_RESPONSE)
				wp.WriteByte(packet.AUTH_INCORRECT_PASSWORD)
				conn.SendFrame(wp.Frame())
				return
			}

			switch resp.Status {
			case sys.Status_SysOK:
				wp := packet.NewWorldPacket(packet.SMSG_AUTH_RESPONSE)
				wp.WriteByte(packet.AUTH_OK)
				conn.SendFrame(wp.Frame())

				// Setup session
				session := &Session{
					WS:          ws,
					Account:     resp.Account,
					GameAccount: resp.GameAccount,
					Connection:  conn,
					Tier:        resp.Tier,
					Locale:      i18n.Locale(resp.Locale),
				}

				session.Init()
				return
			default:
				yo.Warn(resp.Status)
				wp := packet.NewWorldPacket(packet.SMSG_AUTH_RESPONSE)
				wp.WriteByte(packet.AUTH_INCORRECT_PASSWORD)
				conn.SendFrame(wp.Frame())
				return
			}
		}
	}
}
