package blp

type Header struct {
	Version       [4]byte
	Type          uint32
	Compression   uint8
	AlphaDepth    uint8
	AlphaType     uint8
	HasMips       uint8
	Width         uint32
	Height        uint32
	MipmapOffset  [16]uint32
	MipmapLengths [16]uint32
	ColorPalette  [256][4]byte
}
