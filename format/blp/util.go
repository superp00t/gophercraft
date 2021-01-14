//Package blp implements a decoder for BLP2 texture files.
//This package is intended for use with the standard Go image packages.

package blp

import (
	"bytes"
	"image"
	"io"
)

// Decode a BLP stream from a file-like interface into an *image.NRGBA (satisfies image.Image)
func Decode(file io.ReadSeeker) (*image.NRGBA, error) {
	dec, err := newDecoder(file)
	if err != nil {
		return nil, err
	}

	// lower-quality mipmaps are unused.
	img, err := dec.decode(0)
	if err != nil {
		return nil, err
	}

	return img, nil
}

// Decode a BLP data buffer into an *image.NRGBA (satisfies image.Image)
func DecodeBytes(blpData []byte) (*image.NRGBA, error) {
	return Decode(bytes.NewReader(blpData))
}
