package blp

import (
	"fmt"
	"image"
	"io"
	"math"
)

const (
	// Similar to GIF. Uses 8-bit values as index to 256-color palette, ideal for low-res bitmaps.
	TextureRaw = 1
	// Uses S3 Texture Compression algorithm (S3TC/DXT/DXTC)
	TextureS3 = 2
	// Data is stored as a plain ARGB stream. Used when every bit of quality is required.
	TextureUncompressed = 3
)

type scaleData struct {
	Mipmap       int
	Scale        uint32
	ScaledWidth  uint32
	ScaledHeight uint32
	ScaledLength uint32
}

// Precalculated scaling information
func (decoder *decoder) getScale(mip int) *scaleData {
	sd := &scaleData{}
	sd.Mipmap = mip
	sd.Scale = uint32(math.Pow(2, float64(mip)))
	sd.ScaledWidth = decoder.Header.Width / sd.Scale
	sd.ScaledHeight = decoder.Header.Height / sd.Scale
	sd.ScaledLength = sd.ScaledWidth * sd.ScaledHeight
	return sd
}

func (decoder *decoder) decode(mip int) (*image.NRGBA, error) {
	if mip < 0 || mip >= 16 {
		return nil, fmt.Errorf("blp: invalid mipmap %d", mip)
	}

	mipmap := decoder.MipmapOffset[mip]
	if mipmap == 0 {
		return nil, fmt.Errorf("blp: file includes no mipmap %d", mip)
	}

	// Go to where the BLP header says a mipmap is
	_, err := decoder.ReadSeeker.Seek(int64(mipmap), 0)
	if err != nil {
		return nil, err
	}

	var rgba []byte

	sd := decoder.getScale(mip)

	switch decoder.Compression {
	case TextureRaw:
		rgba, err = decoder.decodeRaw(sd)
	case TextureS3:
		rgba, err = decoder.decodeS3(sd)
	case TextureUncompressed:
		rgba, err = decoder.decodeUncompressed(sd)
	default:
		return nil, fmt.Errorf("blp: unsupported compression type %d", decoder.Compression)
	}

	if err != nil {
		return nil, err
	}

	rgbaImage := &image.NRGBA{
		Pix:    rgba,
		Stride: 4 * int(sd.ScaledWidth),
		Rect:   image.Rect(0, 0, int(sd.ScaledWidth), int(sd.ScaledHeight)),
	}

	return rgbaImage, nil
}

func (d *decoder) getAlpha(sd *scaleData, data []byte, pixel uint32) uint8 {
	switch d.Header.AlphaDepth {
	case 1:
		byte := uint32(data[sd.ScaledLength+(pixel/8)])
		t := (byte & (0x01 << (pixel % 8))) == 0
		if t {
			return 0x00
		} else {
			return 0xFF
		}
	case 4:
		byte := uint32(data[sd.ScaledLength+(pixel/2)])
		even := (pixel%2 == 0)
		if even {
			return uint8((byte & 0x0F) << 4)
		} else {
			return uint8(byte & 0xF0)
		}
	case 8:
		return data[sd.ScaledLength+pixel]
	default:
		return 0xFF
	}
}

func (decoder *decoder) decodeRaw(sd *scaleData) ([]byte, error) {
	// RGBA data
	rgba := make([]byte, sd.ScaledLength*4)
	// Indices to color palette.
	pixelRefs := make([]byte, decoder.Header.MipmapLengths[sd.Mipmap])
	if len(pixelRefs) == 0 {
		return nil, fmt.Errorf("blp: no pixel data in texture")
	}

	_, err := io.ReadFull(decoder, pixelRefs)
	if err != nil {
		return nil, err
	}

	for iPixel := uint32(0); iPixel < sd.ScaledLength; iPixel++ {
		pixelRef := pixelRefs[iPixel]
		// The Header's palette contains BGRA colors. To get the correct color for this pixel reference byte, convert its values to RGBA
		var rgbaPixel [4]byte
		var bgraPixel = decoder.Header.ColorPalette[pixelRef]
		// [R]GBA = BG[R]A
		rgbaPixel[0] = bgraPixel[2]
		// R[G]BA = B[G]RA
		rgbaPixel[1] = bgraPixel[1]
		// RG[B]A = [B]GRA
		rgbaPixel[2] = bgraPixel[0]

		rgbaPixel[3] = decoder.getAlpha(sd, pixelRefs, iPixel)

		// Copy pixel to RGBA buffer.
		copy(rgba[iPixel*4:(iPixel*4)+4], rgbaPixel[:])
	}

	return rgba, nil
}

func (decoder *decoder) decodeUncompressed(sd *scaleData) ([]byte, error) {
	// Output buffer
	rgba := make([]byte, sd.ScaledLength*4)

	// Input buffer. The general layout of data is the same, we just need to swap the channel bytes.
	bgra := make([]byte, sd.ScaledLength*4)

	_, err := io.ReadFull(decoder, bgra)
	if err != nil {
		return nil, err
	}

	var rgbaPixel [4]byte

	for iPixel := uint32(0); iPixel < sd.ScaledLength; iPixel++ {
		startIndex := iPixel * 4
		endIndex := (iPixel * 4) + 4

		bgraPixel := bgra[startIndex:endIndex]

		// [R]GBA = BG[R]A
		rgbaPixel[0] = bgraPixel[2]
		// R[G]BA = B[G]RA
		rgbaPixel[1] = bgraPixel[1]
		// RG[B]A = [B]GRA
		rgbaPixel[2] = bgraPixel[0]
		// RGB[A] = BGR[A]
		rgbaPixel[3] = bgraPixel[3]

		copy(rgba[startIndex:endIndex], rgbaPixel[:])
	}

	return rgba, nil
}
