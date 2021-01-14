//Package chunked implements RIFF-like encoding used in the WDT and ADT terrain formats
package chunked

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"unicode/utf8"
)

type ID [4]byte

var Zero ID

func bswap(in, out *[4]byte) {
	out[0] = in[3]
	out[1] = in[2]
	out[2] = in[1]
	out[3] = in[0]
}

func (id ID) String() string {
	if bytes.Equal(id[:], Zero[:]) {
		return "Zero"
	}

	var idb [4]byte = id
	var prints [4]byte
	bswap(&idb, &prints)

	str := string(prints[:])
	if !utf8.ValidString(str) {
		return "Invalid chunk ID: " + hex.EncodeToString(id[:])
	}

	return string(prints[:])
}

func CnkID(s string) ID {
	if len(s) != 4 {
		panic(s)
	}

	var inputBytes [4]byte
	copy(inputBytes[:], []byte(s))

	var outputBytes [4]byte
	bswap(&inputBytes, &outputBytes)

	return ID(outputBytes)
}

type Reader struct {
	Reader io.Reader
}

func (c *Reader) ReadChunk() (ID, []byte, error) {
	var id ID
	idBytes, err := c.Reader.Read(id[:])
	if err != nil {
		return id, nil, err
	}

	if idBytes != 4 {
		return Zero, nil, io.EOF
	}

	if id == Zero {
		return Zero, nil, io.EOF
	}

	var size uint32
	err = binary.Read(c.Reader, binary.LittleEndian, &size)
	if err != nil {
		return id, nil, err
	}

	if size > 0xFFFFF {
		return id, nil, fmt.Errorf("chunked: chunk %s is way too big: %d bytes", id, size)
	}

	bytes := make([]byte, size)

	i, err := io.ReadFull(c.Reader, bytes)
	if err != nil {
		return Zero, nil, err
	}

	if uint32(i) != size {
		return id, nil, fmt.Errorf("chunked: stream did not return all %d bytes referenced in this chunk %s (only %d)", size, id, i)
	}

	return id, bytes, nil
}
