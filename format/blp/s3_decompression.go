package blp

import (
	"encoding/binary"
	"fmt"
	"io"
)

const (
	DXT1 = 1
	DXT3 = 3
	DXT5 = 5
)

// Ported from SereniaBLPLib
// https://github.com/WoW-Tools/SereniaBLPLib/blob/master/SereniaBLPLib/DXTDecompression.cs

func unpack565(block []byte, blockIndex int, packed_offset int, colour []byte, colour_offset int) uint32 {
	// Build packed value
	value := uint32(binary.LittleEndian.Uint16(block[blockIndex+packed_offset : blockIndex+packed_offset+2]))

	red := (value >> 11) & 0x1F
	green := ((value >> 5) & 0x3F)
	blue := (value & 0x1F)

	// Scale up to 8 Bit
	colour[0+colour_offset] = (byte)((red << 3) | (red >> 2))
	colour[1+colour_offset] = (byte)((green << 2) | (green >> 4))
	colour[2+colour_offset] = (byte)((blue << 3) | (blue >> 2))
	colour[3+colour_offset] = 255

	return value
}

func (decoder *decoder) decompressColor(rgba []byte, block []byte, blockIndex int, isDxt1 bool) {
	// Unpack Endpoints
	codes := make([]byte, 16)

	a := unpack565(block, blockIndex, 0, codes, 0)
	b := unpack565(block, blockIndex, 2, codes, 4)

	// generate Midpoints
	for i := 0; i < 3; i++ {
		c := uint32(codes[i])
		d := uint32(codes[4+i])

		if isDxt1 && a <= b {
			codes[8+i] = byte((c + d) / 2)
			codes[12+i] = 0
		} else {
			codes[8+i] = byte((2*c + d) / 3)
			codes[12+i] = byte((c + 2*d) / 3)
		}
	}

	// Fill in alpha for intermediate values
	codes[8+3] = 255
	iAlpha := uint8(255)
	if isDxt1 && a <= b {
		iAlpha = 0
	}
	codes[12+3] = iAlpha

	// unpack the indices
	indices := make([]byte, 16)
	for i := 0; i < 4; i++ {
		packed := uint32(block[blockIndex+4+i])

		indices[0+i*4] = (byte)(packed & 0x3)
		indices[1+i*4] = (byte)((packed >> 2) & 0x3)
		indices[2+i*4] = (byte)((packed >> 4) & 0x3)
		indices[3+i*4] = (byte)((packed >> 6) & 0x3)
	}

	// store out the colours
	for i := 0; i < 16; i++ {
		offset := 4 * int(indices[i])

		copy(rgba[4*i:4*i+4], codes[offset:offset+4])
	}
}

func (decoder *decoder) decompress(rgba, block []byte, blockIndex int, dxtType uint8) {
	// get the block locations
	alphaIndex := blockIndex
	colorBlockIndex := blockIndex

	if dxtType == DXT3 || dxtType == DXT5 {
		colorBlockIndex += 8
	}

	// decompress color
	decoder.decompressColor(rgba, block, colorBlockIndex, dxtType == DXT1)

	// decompress alpha separately if necessary
	if dxtType == DXT3 {
		decompressAlphaDxt3(rgba, block, alphaIndex)
	} else if dxtType == DXT5 {
		decompressAlphaDxt5(rgba, block, alphaIndex)
	}
}

func decompressAlphaDxt3(rgba, block []byte, alphaIndex int) {
	// Unpack the alpha values pairwise
	// 16 4-bit values (8 bytes)

	for i := 0; i < 8; i++ {
		// // Quantise down to 4 bits
		quant := block[alphaIndex+i]

		lo := quant & 0x0F
		hi := quant & 0xF0

		lo = (lo | (lo << 4))
		hi = (hi | (hi >> 4))

		loIndex := 8*i + 3
		hiIndex := 8*i + 7

		// Convert back up to bytes
		rgba[loIndex] = byte(lo)
		rgba[hiIndex] = byte(hi)
	}
}

func decompressAlphaDxt5(rgba, block []byte, blockIndex int) {
	// Get the two alpha values
	alpha0 := int(block[blockIndex+0])
	alpha1 := int(block[blockIndex+1])

	// compare the values to build the codebook
	var codes [8]byte
	codes[0] = uint8(alpha0)
	codes[1] = uint8(alpha1)
	if alpha0 <= alpha1 {
		// Use 5-Alpha Codebook
		for i := 1; i < 5; i++ {
			codes[1+i] = (byte)(((5-i)*alpha0 + i*alpha1) / 5)
		}
		codes[6] = 0
		codes[7] = 255
	} else {
		// Use 7-Alpha Codebook
		for i := 1; i < 7; i++ {
			codes[i+1] = (byte)(((7-i)*alpha0 + i*alpha1) / 7)
		}
	}

	// decode indices
	var indices [16]byte

	blockSrc_pos := 2
	indices_pos := 0
	for i := 0; i < 2; i++ {
		// grab 3 bytes
		value := uint32(0)
		for j := 0; j < 3; j++ {
			_byte := uint32(block[blockIndex+blockSrc_pos])
			blockSrc_pos++
			value |= (_byte << 8 * uint32(j))
		}

		// unpack 8 3-bit values from it
		for j := 0; j < 8; j++ {
			index := (value >> 3 * uint32(j)) & 0x07
			indices[indices_pos] = byte(index)
			indices_pos++
		}
	}

	// write out the indexed codebook values
	for i := 0; i < 16; i++ {
		rgba[4*i+3] = codes[indices[i]]
	}
}

func (decoder *decoder) decodeS3(sd *scaleData) ([]byte, error) {
	// Determine algorithm to use
	switch decoder.Header.AlphaType {
	case 0:
		return decoder.decodeDXT(sd, DXT1)
	case 1:
		return decoder.decodeDXT(sd, DXT3)
	case 7:
		return decoder.decodeDXT(sd, DXT5)
	default:
		return nil, fmt.Errorf("blp: unknown S3 compression type: %d", decoder.Header.AlphaType)
	}
}

func (decoder *decoder) decodeDXT(sd *scaleData, dxtType uint8) ([]byte, error) {
	rgba := make([]byte, sd.ScaledLength*4)
	input := make([]byte, decoder.Header.MipmapLengths[sd.Mipmap])

	var targetRGBA [4 * 16]byte

	_, err := io.ReadFull(decoder, input)
	if err != nil {
		return nil, err
	}

	sourceBlock_pos := 0
	bytesPerBlock := 8
	if dxtType == DXT3 || dxtType == DXT5 {
		bytesPerBlock = 16
	}

	for y := uint32(0); y < sd.ScaledHeight; y += 4 {
		for x := uint32(0); x < sd.ScaledWidth; x += 4 {
			if sourceBlock_pos == len(input) {
				continue
			}

			targetRGBA_pos := 0
			decoder.decompress(targetRGBA[:], input, sourceBlock_pos, dxtType)

			// Write the decompressed pixels to the correct image locations
			for py := 0; py < 4; py++ {
				for px := 0; px < 4; px++ {
					sx := int(x) + px
					sy := int(y) + py
					if sx < int(sd.ScaledWidth) && sy < int(sd.ScaledHeight) {
						targetPixel := 4 * ((int(sd.ScaledWidth) * sy) + sx)
						// targetPixel := 4 * (int(sd.ScaledWidth)*sy + sx)

						rgba[targetPixel+0] = targetRGBA[targetRGBA_pos+0]
						rgba[targetPixel+1] = targetRGBA[targetRGBA_pos+1]
						rgba[targetPixel+2] = targetRGBA[targetRGBA_pos+2]
						rgba[targetPixel+3] = targetRGBA[targetRGBA_pos+3]

						targetRGBA_pos += 4
					} else {
						// Ignore that pixel
						targetRGBA_pos += 4
					}
				}
			}
			sourceBlock_pos += bytesPerBlock
		}
	}

	return rgba, nil
}
