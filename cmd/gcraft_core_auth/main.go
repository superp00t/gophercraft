package main

import (
	"crypto/tls"
	"net"
	"net/http"

	"github.com/superp00t/gophercraft/gcore/config"

	"github.com/superp00t/etc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/superp00t/etc/yo"
	"github.com/superp00t/gophercraft/auth"
	"github.com/superp00t/gophercraft/auth/sys"
	"github.com/superp00t/gophercraft/bnet"
	"github.com/superp00t/gophercraft/gcore"
	_ "github.com/superp00t/gophercraft/gcore/dbsupport"
)

var core *gcore.Core

func getConfig() *config.Auth {
	cpath := yo.StringG("c")
	if cpath == "" {
		cpath = etc.LocalDirectory().Concat("gcraft_auth").Render()
	}

	yo.Ok("reading", cpath)

	if etc.ParseSystemPath(cpath).IsExtant() == false {
		yo.Println("No config file found at", cpath)
		yo.Confirm("Create one?")

		err := config.GenerateDefaultAuth(cpath)
		if err != nil {
			yo.Fatal(err)
		}

		yo.Ok("Default configuration file created at", cpath)
	}

	cfg, err := config.LoadAuth(cpath)
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
		// uncommenting this breaks it, for some reason:
		// CurvePreferences:         []tls.CurveID{tls.CurveP521, tls.CurveP384, tls.CurveP256},
		// PreferServerCipherSuites: true,
		// CipherSuites: []uint16{
		// 	tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
		// 	tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
		// 	tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
		// 	tls.TLS_RSA_WITH_AES_256_CBC_SHA,
		// },
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
	yo.Stringf("c", "config", "the location of your config file", "")

	yo.Main("Gophercraft Core Authentication Server", func(c []string) {
		gcore.PrintLicense()

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
	})

	yo.Init()
}
