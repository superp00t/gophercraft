package bnet

import (
	"crypto/tls"

	"github.com/golang/protobuf/proto"
	"github.com/superp00t/etc"
	"github.com/superp00t/etc/yo"
	p "github.com/superp00t/gophercraft/bnet/bgs/protocol"
	cn "github.com/superp00t/gophercraft/bnet/bgs/protocol/connection/v1"
)

func Dial(endpoint string) (*Conn, error) {
	cn := new(Conn)
	var err error
	cn.c, err = tls.Dial("tcp", endpoint, &tls.Config{
		InsecureSkipVerify: true,
	})
	if err != nil {
		return nil, err
	}

	return cn, nil
}

func (c *Conn) Emit(h *p.Header, data proto.Message) {
	body, err := proto.Marshal(data)
	if err != nil {
		panic(err)
	}

	sz := uint32(len(body))
	h.Size = &sz

	head, err := proto.Marshal(h)
	if err != nil {
		yo.Fatal(err)
	}

	yo.Spew(body)
	yo.Spew(head)

	pkt := etc.NewBuffer()
	pkt.WriteBigUint16(uint16(len(head)))
	_, err = c.c.Write(pkt.Bytes())
	if err != nil {
		panic(err)
	}

	_, err = c.c.Write(head)
	if err != nil {
		panic(err)
	}

	_, err = c.c.Write(body)
	if err != nil {
		panic(err)
	}
}

func (c *Conn) Connect() error {
	tr := true
	cr := &cn.ConnectRequest{
		UseBindlessRpc: &tr,
	}

	methodID := uint32(1)
	serviceID := uint32(0)
	token := uint32(0)

	head := &p.Header{
		ServiceId: &serviceID,
		MethodId:  &methodID,
		Token:     &token,
	}

	c.Emit(head, cr)

	resph, data, err := c.ReadHeader()
	if err != nil {
		yo.Warn("Err reading header", err)
		return err
	}

	rsp := new(cn.ConnectResponse)
	err = proto.Unmarshal(data, rsp)
	if err != nil {
		return err
	}

	yo.Spew(resph)
	yo.Spew(rsp)

	return nil
}
