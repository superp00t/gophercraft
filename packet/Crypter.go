package packet

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"net"
	"sync"

	"github.com/superp00t/etc"

	"github.com/superp00t/gophercraft/arc4"
)

func NewCipher(version uint32, sessionKey []byte, server bool) (arc4.Cipher, error) {
	var c arc4.Cipher

	switch version {
	case 5875:
		c = &arc4.Cipher5875{}
	case 12340:
		c = &arc4.Cipher12340{}
	}

	if err := c.Init(server, sessionKey); err != nil {
		return nil, err
	}

	return c, nil
}

// Crypter provides a buffered communication pipe with which to send and receive partially encrypted packets through the game protocol.
type Crypter struct {
	Conn       net.Conn
	Reader     *bufio.Reader
	SessionKey []byte
	Cipher     arc4.Cipher
	Server     bool
	write      sync.Mutex
	closed     bool
}

func NewCrypter(version uint32, c net.Conn, sessionKey []byte, server bool) *Crypter {
	cr := new(Crypter)
	cr.Conn = c
	cr.Reader = bufio.NewReaderSize(c, 65535)
	cr.SessionKey = sessionKey
	cr.Server = server
	var err error
	cr.Cipher, err = NewCipher(version, sessionKey, server)
	if err != nil {
		panic(err)
	}

	return cr
}

type Frame struct {
	Type WorldType
	Data []byte
}

func (wp *WorldPacket) Frame() Frame {
	return Frame{wp.Type, wp.Buffer.Bytes()}
}

func (cl *Crypter) SendFrame(f Frame) error {
	cl.write.Lock()
	defer cl.write.Unlock()
	offset := 6
	if cl.Server {
		offset = 4
	}
	data := etc.NewBuffer()
	data.WriteBigUint16(uint16(len(f.Data)) + uint16(offset-2))
	if cl.Server {
		data.WriteUint16(uint16(f.Type))
	} else {
		data.WriteUint32(uint32(f.Type))
	}
	data.Write(f.Data)
	dat := data.Bytes()

	cl.Cipher.Encrypt(dat[:offset])

	_, err := cl.Conn.Write(dat)
	return err
}

func (cl *Crypter) ReadFrame() (Frame, error) {
	offset := 4
	header := make([]byte, 6)
	if cl.Server == true {
		offset = 6
	}
	i, err := cl.Reader.Read(header[:offset])
	if err != nil {
		return Frame{}, err
	}

	if i != offset {
		return Frame{}, fmt.Errorf("packet: read only %d/%d", i, offset)
	}

	cl.Cipher.Decrypt(header[:offset])

	ln := int(binary.BigEndian.Uint16(header[0:2]))

	var opcode uint32

	if cl.Server {
		opcode = binary.LittleEndian.Uint32(header[2:offset])
	} else {
		opcode = uint32(binary.LittleEndian.Uint16(header[2:offset]))
	}

	frameContent := make([]byte, ln-(offset-2))

	_, err = cl.Reader.Read(frameContent)
	if err != nil {
		return Frame{}, err
	}

	return Frame{
		WorldType(opcode),
		frameContent,
	}, nil
}
