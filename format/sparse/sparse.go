package sparse

import (
	"github.com/superp00t/etc"
)

func Decompress(input []byte) ([]byte, error) {
	outputStream := etc.NewBuffer()
	src := etc.FromBytes(input)
	src.ReadUint32()

	for src.Available() > 0 {
		next := src.ReadByte()

		if next&0x80 != 0 {
			chunkSize := (next & 0x7F) + 1
			chunk := src.ReadBytes(int(chunkSize))
			outputStream.Write(chunk)
		} else {
			chunkSize := (next & 0x7f) + 3
			for x := uint8(0); x < chunkSize; x++ {
				outputStream.WriteByte(0)
			}
		}
	}

	return outputStream.Bytes(), nil
}
