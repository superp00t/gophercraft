package bnet

import (
	"bufio"
	"crypto/rand"
	"crypto/tls"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/golang/protobuf/proto"
	"github.com/gorilla/mux"
	"github.com/superp00t/etc"
	"github.com/superp00t/etc/yo"
	p "github.com/superp00t/gophercraft/bnet/bgs/protocol"
	"github.com/superp00t/gophercraft/bnet/realmlist"
	"github.com/superp00t/gophercraft/gcore"
)

func GenerateTicket() string {
	var random [16]byte
	io.ReadFull(rand.Reader, random[:])
	return "TC-" + strings.ToUpper(hex.EncodeToString(random[:]))
}

type Backend interface {
	GetAccount(user string) (*gcore.Account, []gcore.GameAccount, error)
	StoreKey(user, locale, platform string, K []byte)
	StoreLoginTicket(user, ticket string, expiry time.Time)
	GetTicket(ticket string) (string, time.Time)
	AccountID(user string) uint64
	InfoHandler() http.Handler
	ListRealms() []gcore.Realm
}

type Listener struct {
	Backend      Backend
	RESTAddress  string
	HostExternal string
	l            net.Listener
}

func Bind(address string) (net.Listener, error) {
	cer, err := tls.X509KeyPair(
		[]byte(Cert),
		[]byte(Key),
	)
	if err != nil {
		return nil, err
	}

	config := &tls.Config{Certificates: []tls.Certificate{cer}}
	ln, err := tls.Listen("tcp", address, config)
	return ln, err
}

func Listen(address, restAddress, externalHost string) (*Listener, error) {
	ln, err := Bind(address)
	if err != nil {
		return nil, err
	}

	list := &Listener{
		RESTAddress:  restAddress,
		HostExternal: externalHost,
		l:            ln,
	}
	return list, nil
}

type response struct {
	header *p.Header
	data   []byte
}

type Conn struct {
	server           *Listener
	c                net.Conn
	r                *bufio.Reader
	authed           bool
	tokenCtr         uint32
	pendingRequests  map[uint32]chan response
	pendingRequestsL sync.Mutex

	locale          string
	platform        string
	user            string
	ticket          string
	version         uint32
	gameAccountName string

	versionInfo *realmlist.ClientVersion

	SecretData [64]byte
}

func (c *Conn) Close() error {
	return c.c.Close()
}

func (c *Conn) Read(b []byte) (int, error) {
	return c.c.Read(b[:])
}

func (l *Listener) Accept() (*Conn, error) {
	cn, err := l.l.Accept()
	if err != nil {
		return nil, err
	}
	return &Conn{
		c:               cn,
		r:               bufio.NewReaderSize(cn, 0xFFFF),
		server:          l,
		pendingRequests: make(map[uint32]chan response),
	}, nil
}

func (c *Conn) Handle() {
	spew.Config.DisableMethods = true
	for {
		header, data, err := c.ReadHeader()
		if err != nil {
			yo.Warn("reading header: ", err)
			c.c.Close()
			return
		}
		c.Dispatch(header, data)
	}
}

func (c *Conn) ReadHeader() (*p.Header, []byte, error) {
	bs, err := c.ReadMsgBytes()
	if err != nil {
		return nil, nil, err
	}

	bn := new(p.Header)

	err = proto.Unmarshal(bs, bn)
	if err != nil {
		return nil, nil, err
	}

	if bn.Size == nil {
		return nil, nil, fmt.Errorf("no size header")
	}

	data := make([]byte, *bn.Size)
	_, err = c.Read(data)
	if err != nil {
		return nil, nil, err
	}

	return bn, data, nil
}

func (c *Conn) ReadMsgBytes() ([]byte, error) {
	var msgLength uint16

	err := binary.Read(c, binary.BigEndian, &msgLength)
	if err != nil {
		return nil, err
	}
	fmt.Println("Getting ", msgLength, "bytes")

	msgData := make([]byte, msgLength)
	// i, err := io.ReadFull(c, msgData)
	// if err != nil {
	// 	return nil, err
	// }

	i, err := c.Read(msgData)
	if err != nil {
		return nil, err
	}

	fmt.Println("Recvd", msgLength)

	if i != int(msgLength) {
		yo.Spew(msgData)
		return nil, fmt.Errorf("Count of received bytes (%d) was not equal to length prefix (%d)", i, msgLength)
	}

	return msgData, nil
}

func (c *Conn) Dispatch(header *p.Header, data []byte) {
	sid := header.GetServiceHash()

	if header.GetServiceId() == ResponseService {
		c.pendingRequestsL.Lock()
		if ch := c.pendingRequests[header.GetToken()]; ch != nil {
			delete(c.pendingRequests, header.GetToken())
			c.pendingRequestsL.Unlock()
			ch <- response{header, data}
			return
		} else {
			yo.Warn("Unknown token", header.GetToken())
		}
		c.pendingRequestsL.Unlock()
		return
	}

	switch sid {
	case AccountServiceHash:
		yo.Ok("Dispatching account service", header.GetMethodId())
		DispatchAccountService(c, c.server, header.GetToken(), header.GetMethodId(), data)
	case ConnectionServiceHash:
		yo.Ok("connection method", header.GetMethodId())
		DispatchConnectionService(c, c.server, header.GetToken(), header.GetMethodId(), data)
	case AuthenticationServiceHash:
		yo.Ok("authentication method", header.GetMethodId())
		DispatchAuthenticationService(c, c.server, header.GetToken(), header.GetMethodId(), data)
	case GameUtilitiesServiceHash:
		yo.Ok("Game utilities method", header.GetMethodId())
		DispatchGameUtilitiesService(c, c.server, header.GetToken(), header.GetMethodId(), data)
	default:
		yo.Warnf("unknown service 0x%08X %d\n", header.GetServiceHash(), header.GetServiceId())
	}
}

func (c *Conn) Request(servicehash, method uint32, data proto.Message) (*p.Header, []byte, error) {
	ch := make(chan response)
	c.pendingRequestsL.Lock()
	token := c.tokenCtr
	c.tokenCtr++
	c.pendingRequests[token] = ch
	c.pendingRequestsL.Unlock()

	var err error
	var content []byte

	sid := uint32(0)

	size := uint32(0)
	if data != nil {
		content, err = proto.Marshal(data)
		if err != nil {
			return nil, nil, err
		}
		size = uint32(len(content))
	}

	h := &p.Header{
		ServiceId:   &sid,
		ServiceHash: &servicehash,
		MethodId:    &method,
		Token:       &token,
	}

	if size > 0 {
		h.Size = &size
	}

	yo.Ok("sending")
	yo.Spew(h)
	yo.Spew(content)

	header, err := proto.Marshal(h)
	if err != nil {
		return nil, nil, err
	}

	e := etc.NewBuffer()
	e.WriteBigUint16(uint16(len(header)))
	e.Write(header)
	e.Write(content)

	_, err = c.c.Write(e.Bytes())
	if err != nil {
		return nil, nil, err
	}

	select {
	case <-time.After(18 * time.Second):
		return nil, nil, fmt.Errorf("bnet: request timed out")
	case resp := <-ch:
		return resp.header, resp.data, nil
	}
}

func (c *Conn) SendResponse(token uint32, v proto.Message) {
	c.SendResponseMessage(false, ERROR_OK, token, v)
}

func (c *Conn) SendResponseOK(token uint32, v proto.Message) {
	c.SendResponseMessage(true, ERROR_OK, token, v)
}

func (c *Conn) SendResponseMessage(useStatus bool, status Status, token uint32, v proto.Message) {
	serial, err := proto.Marshal(v)
	if err != nil {
		panic(err)
	}

	contentSize := uint32(len(serial))

	h := &p.Header{
		ServiceId: &ResponseService,
		Token:     &token,
		Size:      &contentSize,
	}

	if useStatus {
		h.Status = u32p(uint32(status))
	}

	headerBytes, err := proto.Marshal(h)
	if err != nil {
		panic(err)
	}

	e := etc.NewBuffer()
	e.WriteBigUint16(uint16(len(headerBytes)))
	e.Write(headerBytes)
	e.Write(serial)

	_, err = c.c.Write(e.Bytes())
	if err != nil {
		yo.Warn(err)
	}
	yo.Ok("Sent response")
}

var (
	ResponseService = uint32(0xFE)
)

func (c *Conn) SendResponseCode(token uint32, status Status) {
	st := uint32(status)
	header := p.Header{}
	header.ServiceId = &ResponseService
	header.Token = &token
	header.Status = &st

	h, err := proto.Marshal(&header)
	if err != nil {
		yo.Fatal(err)
	}

	content := etc.NewBuffer()
	content.WriteBigUint16(uint16(len(h)))
	content.Write(h)

	_, err = c.c.Write(content.Bytes())
	if err != nil {
		yo.Warn(err)
	}
}

func extractL(input []byte) (int, uint16) {
	e := etc.FromBytes(input)
	lngth := int(e.ReadBigUint16())
	op := e.ReadUint16()

	return lngth, op
}

func (lst *Listener) Serve() error {
	ch := make(chan error)

	go func() {
		ln, err := Bind(lst.RESTAddress)
		if err != nil {
			ch <- err
			return
		}

		r := mux.NewRouter()
		r.Handle("/", lst.Backend.InfoHandler())

		bs := r.PathPrefix("/bnetserver/").Subrouter()
		bs.HandleFunc("/login/", lst.HandleLoginGet).Methods("GET")
		bs.HandleFunc("/login/", lst.HandleLoginPost).Methods("POST")
		ch <- http.Serve(ln, r)
	}()

	go func() {
		for {
			cn, err := lst.Accept()
			if err != nil {
				ch <- err
				return
			}

			yo.Ok("Accepted conn")

			go cn.Handle()
		}
	}()

	return <-ch
}
