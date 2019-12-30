package main

import (
	"net"
)

type rpcSrv struct {
	conns chan net.Conn
	addr  net.Addr
}

func (r *rpcSrv) Accept() (net.Conn, error) {
	return <-r.conns, nil
}

func (r *rpcSrv) Addr() net.Addr {
	return r.addr
}

func (r *rpcSrv) Close() error {
	panic("cannot close")
}
