package blp

import (
	"encoding/binary"
	"fmt"
	"image"
	"image/draw"
	"io"
)

const (
	colorB = iota
	colorG
	colorR
	colorA
)

type encoder struct {
	Source *image.NRGBA
	Header
	io.Writer
}

func getNRGBA(src image.Image) *image.NRGBA {
	switch img := src.(type) {
	case *image.NRGBA:
		return img
	default:
		b := src.Bounds()
		m := image.NewNRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
		draw.Draw(m, m.Bounds(), src, b.Min, draw.Src)
		return m
	}
}

func Encode(source image.Image, writer io.Writer, textureCompression uint8) error {
	switch textureCompression {
	case 0:
		textureCompression = TextureS3
	}

	encoder := &encoder{Writer: writer}
	encoder.Header.Version = BLP2
	encoder.Header.Type = 1
	encoder.Header.Compression = textureCompression
	encoder.Header.AlphaType = 1

	switch textureCompression {
	case TextureUncompressed:
	case TextureS3:
	default:
		return fmt.Errorf("blp: unimplemented compression type 0x%x", textureCompression)
	}

	encoder.Source = getNRGBA(source)

	// Mipmaps are unsupported at this time. This function can write only a single texture.
	encoder.MipmapOffset[0] = uint32(binary.Size(encoder.Header))

	encoder.Width = uint32(encoder.Source.Bounds().Dx())
	encoder.Height = uint32(encoder.Source.Bounds().Dy())

	var mipData []byte
	var err error

	switch textureCompression {
	case TextureUncompressed:
		mipData, err = encoder.encodeUncompressed()
	default:
		panic("unreachable")
	}

	encoder.Header.MipmapLengths[0] = uint32(len(mipData))

	if err := binary.Write(encoder, binary.LittleEndian, &encoder.Header); err != nil {
		return err
	}

	if _, err = encoder.Write(mipData); err != nil {
		return err
	}

	return nil
}

func (encoder *encoder) encodeUncompressed() ([]byte, error) {
	output := make([]byte, len(encoder.Source.Pix))

	numPixels := (len(encoder.Source.Pix) / 4)

	for pixel := 0; pixel < numPixels; pixel++ {
		pixelBegin := pixel * 4
		pixelEnd := (pixel + 1) * 4
		rgba := encoder.Source.Pix[pixelBegin:pixelEnd]
		bgra := output[pixelBegin:pixelEnd]
		bgra[0] = rgba[2]
		bgra[1] = rgba[1]
		bgra[2] = rgba[0]
		bgra[3] = rgba[3]
	}
	return output, nil
}
