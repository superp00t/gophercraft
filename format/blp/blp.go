package blp

/* package blp decodes BLP images.
   ported from the great https://github.com/Kruithne/js-blp/ to Go */

import (
	"fmt"
	"image"
	"math"

	"github.com/superp00t/etc"
)

const (
	DXT1 = 0x1
	DXT3 = 0x2
	DXT5 = 0x4
)

type BLP struct {
	Version       int
	Type          uint32
	Compression   uint8
	AlphaDepth    uint8
	AlphaType     uint8
	HasMips       uint8
	Width, Height uint32
	MipmapOffset  [16]uint32
	MipmapLengths [16]uint32
	Palette       [256][]byte

	buf                       *etc.Buffer
	MapCount                  int
	Scale                     int
	ScaledWidth, ScaledHeight int
	ScaledLength              int
	RawData                   []byte
}

type BLPMap struct {
	B   *BLP
	buf []byte
}

func (b *BLP) Mipmap(i int) image.Image {
	pix := b.getPixels(i)
	min := image.Point{0, 0}
	max := image.Point{b.ScaledWidth, b.ScaledHeight}
	img := image.NewRGBA(image.Rectangle{min, max})
	img.Pix = pix
	return img
}

func DecodeBLP(data []byte) (*BLP, error) {
	b := &BLP{}
	d := etc.MkBuffer(data)

	mStr := string(d.ReadBytes(4))
	switch mStr {
	case "BLP2":
		b.Version = 2
	case "BLP1":
		return nil, fmt.Errorf("Version 1 not supported")
	default:
		return nil, fmt.Errorf("Unrecognized format")
	}

	b.Type = d.ReadUint32()

	b.Compression = d.ReadByte()
	b.AlphaDepth = d.ReadByte()
	b.AlphaType = d.ReadByte()
	b.HasMips = d.ReadByte()
	b.Width = d.ReadUint32()
	b.Height = d.ReadUint32()

	for i := 0; i < 16; i++ {
		b.MipmapOffset[i] = d.ReadUint32()
	}

	for i := 0; i < 16; i++ {
		b.MipmapLengths[i] = d.ReadUint32()
	}

	for i := 0; i < 256; i++ {
		b.Palette[i] = d.ReadBytes(4)
	}

	for _, v := range b.MipmapOffset {
		if v != 0 {
			b.MapCount++
		}
	}

	b.buf = d

	return b, nil
}

func max(x, y int) int {
	if x > y {
		return x
	}

	return y
}

func min(x, y int) int {
	if x < y {
		return x
	}

	return y
}

func pow(x, y int) int {
	return int(math.Pow(float64(x), float64(y)))
}

func (b *BLP) getPixels(mipmap int) []byte {
	mipmap = max(0, min(mipmap, b.MapCount-1))

	b.Scale = pow(2, mipmap)
	b.ScaledWidth = int(b.Width) / b.Scale
	b.ScaledHeight = int(b.Height) / b.Scale
	b.ScaledLength = b.ScaledWidth * b.ScaledHeight

	b.buf.SeekR(int64(b.MipmapOffset[mipmap]))
	b.RawData = b.buf.ReadBytes(int(b.MipmapLengths[mipmap]))
	switch b.Compression {
	case 1:
		return b.GetUncompressed()
	case 2:
		return b.GetCompressed()
	case 3:
		return MarshalBRGA(b.RawData)
	default:
		return nil
	}
}

func (b *BLP) GetUncompressed() []byte {
	d := etc.NewBuffer()
	for i := 0; i < b.ScaledLength; i++ {
		color := b.Palette[b.RawData[i]]
		d.WriteByte(color[2])
		d.WriteByte(color[1])
		d.WriteByte(color[0])
		d.WriteByte(b.GetAlpha(i))
	}
	d.Write(make([]byte, (b.ScaledLength*4)-d.Len()))
	return d.Bytes()
}

func (b *BLP) GetAlpha(i int) uint8 {
	switch b.AlphaDepth {
	case 1:
		byt := b.RawData[b.ScaledLength+(i/8)]
		t := (byt & (0x01 << uint32(i%8))) == 0
		if t {
			return 0x00
		} else {
			return 0xFF
		}
	case 4:
		byt := b.RawData[b.ScaledLength+(i/2)]
		t := (i%2 == 0)
		if t {
			return (byt & 0x0F) << 4
		} else {
			return byt & 0xF0
		}
	case 8:
		return b.RawData[b.ScaledLength+i]
	default:
		return 0xFF
	}
}

func (blp *BLP) GetCompressed() []byte {
	var flags uint8
	isntd1 := blp.AlphaDepth > 1
	if isntd1 {
		isd5 := blp.AlphaType == 7
		if isd5 {
			flags = DXT5
		} else {
			flags = DXT3
		}
	} else {
		flags = DXT1
	}

	target := make([]byte, 256)
	data := make([]byte, blp.ScaledLength*4)
	pos := 0
	blockBytes := 0

	if (flags & DXT1) != 0 {
		blockBytes = 8
	} else {
		blockBytes = 16
	}

	for y := 0; y < blp.ScaledHeight; y += 4 {
		for x := 0; x < blp.ScaledWidth; x += 4 {
			blockPos := 0
			if len(blp.RawData) == pos {
				continue
			}

			colorIndex := pos
			if (flags & (DXT3 | DXT5)) != 0 {
				colorIndex += 8
			}

			isd1 := (flags & DXT1) != 0
			colors := make([]byte, 16)
			a := unpackColor(blp.RawData, colorIndex, 0, colors, 0)
			b := unpackColor(blp.RawData, colorIndex, 2, colors, 4)

			dT := isd1 && a <= b
			for i := 0; i < 3; i++ {
				c := uint32(colors[i])
				d := uint32(colors[i+4])
				if dT {
					colors[i+8] = byte((c + d) / 2)
					colors[i+12] = 0
				} else {
					colors[i+8] = byte((2*c + d) / 3)
					colors[i+12] = byte((c + 2*d) / 3)
				}
			}

			colors[8+3] = 255
			if dT {
				colors[12+3] = 0
			} else {
				colors[12+3] = 255
			}

			index := make([]byte, 16)
			for i := 0; i < 4; i++ {
				packed := uint64(blp.RawData[colorIndex+4+i])
				index[i*4] = uint8(packed & 0x3)
				index[1+i*4] = uint8((packed >> 2) & 0x3)
				index[2+i*4] = uint8((packed >> 4) & 0x3)
				index[3+i*4] = uint8((packed >> 6) & 0x3)
			}

			for i := 0; i < 16; i++ {
				ofs := int(index[i]) * 4
				target[4*i] = colors[ofs]
				target[4*i+1] = colors[ofs+1]
				target[4*i+2] = colors[ofs+2]
				target[4*i+3] = colors[ofs+3]
			}

			if (flags & DXT3) != 0 {
				for i := 0; i < 8; i++ {
					quant := blp.RawData[pos+i]
					low := quant & 0x0F
					high := quant & 0xF0

					target[8*i+3] = (low | (low << 4))
					target[8*i+7] = (high | (high >> 4))
				}
			} else {
				if (flags & DXT5) != 0 {
					a0 := blp.RawData[pos]
					a1 := blp.RawData[pos+1]
					colours := make([]byte, 8)

					colours[0] = a0
					colours[1] = a1

					h := 7
					if a0 <= a1 {
						h = 5

					}

					for i := 1; i < h; i++ {
						_a0 := int(a0)
						_a1 := int(a1)
						o := (((h-i)*_a0 + i*_a1) / h) | 0
						colours[i+1] = byte(o)
					}

					if a0 <= a1 {
						colours[6] = 0
						colours[7] = 0xFF
					}

					indices := make([]byte, 256)
					blkPos := 2
					indicesPos := 0

					for i := 0; i < 2; i++ {
						var value uint64 = 0
						for j := uint64(0); j < 3; j++ {
							byt := uint64(blp.RawData[pos+blkPos])
							blkPos++
							value |= (byt << 8 * j)
						}

						for j := uint64(0); j < 8; j++ {
							indices[indicesPos] = uint8((value >> 3 * j) & 0x07)
							indicesPos++
						}
					}

					for i := 0; i < 16; i++ {
						target[4*i+3] = colours[indices[i]]

					}
				}
			}
			for pY := 0; pY < 4; pY++ {
				for pX := 0; pX < 4; pX++ {
					sX := x + pX
					sY := y + pY

					if sX < blp.ScaledWidth && sY < blp.ScaledHeight {
						pixel := 4 * (blp.ScaledWidth*sY + sX)
						for i := 0; i < 4; i++ {
							data[pixel+i] = target[blockPos+i]
						}
					}
					blockPos += 4
				}
			}
			pos += blockBytes
		}
	}

	return data
}

func unpackColor(block []byte, index, ofs int, color []byte, colorofs int) uint32 {
	value := uint32(block[index+ofs]) | (uint32(block[index+1+ofs]) << 8)

	r := (value >> 11) & 0x1F
	g := (value >> 5) & 0x3F
	b := value & 0x1F

	color[colorofs] = uint8((r << 3) | (r >> 2))
	color[colorofs+1] = uint8((g << 2) | (g >> 4))
	color[colorofs+2] = uint8((b << 3) | (b >> 2))
	color[colorofs+3] = 255

	return value
}

func MarshalBRGA(input []byte) []byte {
	buf := make([]byte, len(input))
	count := len(input) / 4
	for i := 0; i < count; i++ {
		ofs := i * 4
		copy(buf[ofs:ofs+4], []byte{
			input[ofs+2],
			input[ofs+1],
			input[ofs],
			input[ofs+3],
		})
	}
	return buf
}
