package packet

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"net"
	"sync"

	"github.com/superp00t/etc"
	"github.com/superp00t/etc/yo"

	"github.com/superp00t/gophercraft/crypto"
	"github.com/superp00t/gophercraft/vsn"
)

// A Frame's length cannot exceed this size
const MaxLength = 32766

// Connection provides a buffered communication pipe with which to send and receive partially encrypted packets through the game protocol.
// Starting in 8.0, packets are totally encrypted with AES-128 with the exception of length headers.
type Connection struct {
	Build      vsn.Build
	Conn       net.Conn
	Reader     *bufio.Reader
	SessionKey []byte
	Cipher     crypto.Cipher
	Server     bool
	write      sync.Mutex
	closed     bool
}

func NewConnection(version vsn.Build, c net.Conn, sessionKey []byte, server bool) *Connection {
	cr := new(Connection)
	cr.Build = version
	cr.Conn = c
	cr.Reader = bufio.NewReaderSize(c, 65535)
	cr.SessionKey = sessionKey
	cr.Server = server
	var err error
	cr.Cipher, err = crypto.NewCipher(version, sessionKey, server)
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

func (cl *Connection) SendFrame(f Frame) error {
	if cl.Build.AddedIn(vsn.NewCryptSystem) {
		cl.write.Lock()
		defer cl.write.Unlock()
		header := make([]byte, 16)
		// Todo: automatically compress large packets
		binary.LittleEndian.PutUint32(header, uint32(int32(len(f.Data)+2)))

		wt, err := ConvertWorldTypeToUint(cl.Build, f.Type)
		if err != nil {
			return err
		}

		var opcode [2]byte
		binary.LittleEndian.PutUint16(header, uint16(wt))
		data := append(opcode[:], f.Data...)
		if err := cl.Cipher.Encrypt(data, header[4:]); err != nil {
			return err
		}

		if _, err := cl.Conn.Write(header); err != nil {
			return err
		}
		if _, err := cl.Conn.Write(data); err != nil {
			return err
		}
		return nil
	}

	if len(f.Data) > MaxLength {
		return fmt.Errorf("packet: Frame length exceeds maximum %d/%d", len(f.Data), MaxLength)
	}

	cl.write.Lock()
	defer cl.write.Unlock()
	offset := 6
	if cl.Server {
		offset = 4
	}

	// Alpha
	if cl.Build.RemovedIn(vsn.V1_12_1) {
		offset += 2
	}

	u32type, err := ConvertWorldTypeToUint(cl.Build, f.Type)
	if err != nil {
		yo.Warn(err)
		return err
	}

	data := etc.NewBuffer()
	data.WriteBigUint16(uint16(len(f.Data)) + uint16(offset-2))
	if cl.Server {
		data.WriteUint16(uint16(u32type))
	} else {
		data.WriteUint32(u32type)
	}

	// Alpha
	if cl.Build.RemovedIn(vsn.V1_12_1) {
		data.WriteByte(0)
		data.WriteByte(0)
	}

	data.Write(f.Data)
	dat := data.Bytes()

	cl.Cipher.Encrypt(dat[:offset], nil)

	_, err = cl.Conn.Write(dat)
	return err
}

func (cl *Connection) ReadFrame() (Frame, error) {
	if cl.Build.AddedIn(vsn.NewCryptSystem) {
		var header [16]byte
		if _, err := cl.Reader.Read(header[:]); err != nil {
			return Frame{}, err
		}
		size := int32(binary.LittleEndian.Uint32(header[0:4]))
		tag := header[4:]
		if size > 1000000 {
			return Frame{}, fmt.Errorf("frame exceeded 100kb %d", size)
		}
		if size < 2 {
			return Frame{}, fmt.Errorf("frame has no content %d", size)
		}

		sealedFrame := make([]byte, size)
		if _, err := cl.Reader.Read(sealedFrame); err != nil {
			return Frame{}, err
		}

		if err := cl.Cipher.Decrypt(sealedFrame, tag); err != nil {
			return Frame{}, err
		}

		// Todo: handling decompression/multiple packets
		var err error
		f := Frame{}
		f.Type, err = LookupWorldType(cl.Build, uint32(binary.LittleEndian.Uint16(sealedFrame[0:2])))
		if err != nil {
			return f, err
		}

		f.Data = sealedFrame[2:]
		return f, nil
	}

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

	cl.Cipher.Decrypt(header[:offset], nil)

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

	wType, err := LookupWorldType(cl.Build, opcode)
	if err != nil {
		return Frame{}, err
	}

	return Frame{
		wType,
		frameContent,
	}, nil
}

func (c *Connection) Close() error {
	return c.Conn.Close()
}
