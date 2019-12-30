package auth

import (
	"bufio"
	"net"
	"time"
)

// type check
var _ net.Conn = &tmpConn{}

type tmpConn struct {
	conn   net.Conn
	reader *bufio.Reader
}

func wrapConn(c net.Conn) *tmpConn {
	tc := new(tmpConn)
	tc.conn = c
	tc.reader = bufio.NewReader(c)
	return tc
}

func (h *tmpConn) Peek(i int) ([]byte, error) {
	return h.reader.Peek(i)
}

func (h *tmpConn) Read(b []byte) (int, error) {
	return h.reader.Read(b)
}

func (h *tmpConn) Write(b []byte) (int, error) {
	return h.conn.Write(b)
}

func (h *tmpConn) RemoteAddr() net.Addr {
	return h.conn.RemoteAddr()
}

func (h *tmpConn) LocalAddr() net.Addr {
	return h.conn.RemoteAddr()
}

func (h *tmpConn) Close() error {
	return h.conn.Close()
}

func (h *tmpConn) SetDeadline(t time.Time) error {
	return h.conn.SetDeadline(t)
}

func (h *tmpConn) SetReadDeadline(t time.Time) error {
	return h.conn.SetReadDeadline(t)
}

func (h *tmpConn) SetWriteDeadline(t time.Time) error {
	return h.conn.SetWriteDeadline(t)
}
