package blp

import (
	"encoding/binary"
	"fmt"
	"io"
)

var BLP2 [4]byte

func init() {
	copy(BLP2[:], []byte("BLP2"))
}

type decoder struct {
	io.ReadSeeker
	Header
}

func newDecoder(reader io.ReadSeeker) (*decoder, error) {
	decoder := &decoder{ReadSeeker: reader}

	// Use reflection to read to the BLP header.
	err := binary.Read(reader, binary.LittleEndian, &decoder.Header)
	if err != nil {
		return nil, err
	}

	if decoder.Header.Version != BLP2 {
		return nil, fmt.Errorf("blp: not a BLP file: %s", decoder.Header.Version[:])
	}

	return decoder, nil
}
