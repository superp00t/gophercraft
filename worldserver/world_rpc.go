package worldserver

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"time"

	"github.com/superp00t/gophercraft/vsn"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/superp00t/etc/yo"
	"github.com/superp00t/gophercraft/gcore/sys"
	"google.golang.org/grpc"
)

func (ws *WorldServer) dialContext(ctx context.Context, address string) (net.Conn, error) {
	cn, err := tls.Dial("tcp", address, ws.tlsConfig)
	if err != nil {
		return nil, err
	}

	certs := cn.ConnectionState().PeerCertificates

	if len(certs) != 1 {
		return nil, fmt.Errorf("peer certificate wrong size")
	}

	fp, err := sys.GetCertFingerprint(certs[0])
	if err != nil {
		panic(err)
	}

	if fp != ws.Config.AuthServerFingerprint {
		return nil, fmt.Errorf("server has invalid fingerprint %s", fp)
	}

	return cn, nil
}

func (ws *WorldServer) connectRPC() {
	ws.tlsConfig = &tls.Config{
		MinVersion:         tls.VersionTLS12,
		Certificates:       []tls.Certificate{ws.Config.Certificate},
		InsecureSkipVerify: true,
	}

	gc, err := grpc.Dial(
		ws.Config.AuthServer,
		grpc.WithInsecure(),
		grpc.WithContextDialer(ws.dialContext),
	)
	if err != nil {
		panic(err)
	}

	cl := sys.NewAuthServiceClient(gc)

	ws.AuthServiceClient = cl

	vi, err := cl.GetVersionData(context.Background(), &empty.Empty{})
	if err != nil {
		yo.Warn(err)
	} else {
		if vi.CoreVersion != vsn.CoreVersion {
			yo.Warn("Your authentication server is using Gophercraft", vi.CoreVersion, ", and this server is using", vsn.CoreVersion, ". This is not necessarily a problem, but may lead to bugs and other unanticipated behavior. You are encouraged to update your Gophercraft installation.")
		}
	}

	for {
		st, err := cl.AnnounceRealm(context.Background(), &sys.AnnounceRealmMsg{
			RealmID:          ws.Config.RealmID,
			Type:             ws.Config.Type,
			RealmName:        ws.Config.RealmName,
			RealmDescription: ws.Config.RealmDescription,
			Build:            uint32(ws.Config.Version),
			Address:          ws.Config.PublicAddress,
			ActivePlayers:    uint32(len(ws.PlayerList)),
		})
		if err != nil {
			yo.Warn(err)
		}
		if st != nil && st.Status != sys.Status_SysOK {
			yo.Warn("Recieved non-ok status", st.Status)
		}
		time.Sleep(8 * time.Second)
	}
}
