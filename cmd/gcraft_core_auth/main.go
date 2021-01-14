package main

import (
	"crypto/tls"
	"net"
	"net/http"
	"os"

	"github.com/superp00t/gophercraft/gcore/config"
	"github.com/superp00t/gophercraft/vsn"

	"github.com/superp00t/etc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/superp00t/etc/yo"
	"github.com/superp00t/gophercraft/auth"
	"github.com/superp00t/gophercraft/bnet"
	"github.com/superp00t/gophercraft/gcore"
	_ "github.com/superp00t/gophercraft/gcore/dbsupport"
	"github.com/superp00t/gophercraft/gcore/sys"
)

var authLoc etc.Path
var core *gcore.Core

func getConfig() *config.Auth {
	if authLoc.IsExtant() == false {
		yo.Println("No config file found at", authLoc.Render())
		yo.Println("You can create one using gcraft_wizard.")
		os.Exit(0)
	}

	cfg, err := config.LoadAuth(authLoc.Render())
	if err != nil {
		yo.Fatal(err)
	}

	fp, _ := sys.GetCertFileFingerprint(cfg.Path.Concat("cert.pem").Render())

	yo.Ok("This server's fingerprint is", fp)

	if cfg.HostExternal == "" {
		cfg.HostExternal = "localhost"
	}

	return cfg
}

type authServer struct {
	*gcore.Core
	conf      *config.Auth
	rpcServer *rpcSrv
}

func (s *authServer) tlsConfig() *tls.Config {
	cfg := &tls.Config{
		Certificates: []tls.Certificate{s.conf.Certificate},
		MinVersion:   tls.VersionTLS12,
		ClientAuth:   tls.RequireAnyClientCert,
	}
	return cfg
}

func (s *authServer) HandleSpecialConn(protocol string, conn net.Conn) {
	switch protocol {
	case "grpc":
		s.rpcServer.conns <- conn
	}
}

func main() {
	authLoc = etc.LocalDirectory().Concat("Gophercraft").Concat("Auth")
	if len(os.Args) > 1 {
		authLoc = etc.ParseSystemPath(os.Args[1])
	}

	vsn.PrintBanner()

	cfg := getConfig()

	yo.Println("Starting Gophercraft Core Auth Server...")

	var err error
	core, err = gcore.NewCore(cfg)
	if err != nil {
		yo.Fatal(err)
	}

	yo.Ok("Database opened without issue.")

	addr, err := net.ResolveTCPAddr("tcp", cfg.AuthListen)
	if err != nil {
		yo.Fatal(err)
	}

	backend := &authServer{}
	backend.conf = cfg
	backend.Core = core

	backend.rpcServer = &rpcSrv{
		make(chan net.Conn),
		addr,
	}

	go func() {
		yo.Println("Starting HTTP server at", cfg.HTTPInternal)
		mux := core.WebAPI()
		yo.Fatal(http.ListenAndServe(cfg.HTTPInternal, mux))
	}()

	go func() {
		// It should probably be disabled by default like this.
		if cfg.AlphaRealmlistListen != "" {
			yo.Println("Starting Alpha realmlist at", cfg.AlphaRealmlistListen)

			alpha, err := net.Listen("tcp", cfg.AlphaRealmlistListen)
			if err != nil {
				yo.Fatal(err)
			}

			for {
				conn, err := alpha.Accept()
				if err != nil {
					yo.Fatal(err)
				}

				go backend.serveAlpha(conn)
			}
		}
	}()

	go func() {
		yo.Println("Starting bnet server at", cfg.BnetListen)
		lst, err := bnet.Listen(cfg.BnetListen, cfg.BnetRESTListen, cfg.HostExternal)
		if err != nil {
			yo.Fatal(err)
		}

		lst.Backend = core

		yo.Fatal(lst.Serve())
	}()

	go func() {
		config := backend.tlsConfig()

		grpcServer := grpc.NewServer(grpc.Creds(credentials.NewTLS(config)))
		sys.RegisterAuthServiceServer(grpcServer, backend.Core)
		grpcServer.Serve(backend.rpcServer)
	}()

	server := &auth.Server{
		Listen:  cfg.AuthListen,
		Backend: backend,
	}

	yo.Println("Starting Realmlist/GRPC server at", cfg.AuthListen)
	yo.Fatal(server.Run())

}
