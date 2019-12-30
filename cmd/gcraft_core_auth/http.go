package main

import (
	"encoding/json"
	"net"
	"net/http"
)

type httpSrv struct {
	conns chan net.Conn
	addr  net.Addr
}

func (s *httpSrv) Accept() (net.Conn, error) {
	return <-s.conns, nil
}

func (s *httpSrv) Addr() net.Addr {
	return s.addr
}

func (s *httpSrv) Close() error {
	panic("cannot close")
}

func respond(rw http.ResponseWriter, v interface{}) {
	json.NewEncoder(rw).Encode(v)
}
