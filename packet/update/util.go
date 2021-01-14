package update

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"
)

// Bitmask stores a bit-packed list of boolean values.
type Bitmask []uint32

func (b Bitmask) String() string {
	str := fmt.Sprintf("(len: %d)", len(b))

	for x := uint32(0); x < uint32(len(b)*32); x++ {
		if b.Enabled(x) {
			str += fmt.Sprintf(" 0x%04X", x)
		}
	}

	return str
}

func (mask Bitmask) Len() int {
	return len(mask)
}

func NewBitmask() *Bitmask {
	var offsets Bitmask
	return &offsets
}

func (b Bitmask) Clear() {
	for i := range b {
		b[i] = 0
	}
}

func ReadBitmask(descriptor *Descriptor, reader io.Reader) (*Bitmask, error) {
	// struct Bitmask {
	// uint8_t size;
	// uint32_t enabled_offsets[size];
	// };

	var size [1]byte
	_, err := reader.Read(size[:])
	if err != nil {
		return nil, fmt.Errorf("update: error reading bitmask length: %s", err)
	}

	bmask := make(Bitmask, size[0])

	if size[0] == 0 {
		return &bmask, nil
	}

	for chunk := uint8(0); chunk < size[0]; chunk++ {
		var bits [4]byte
		_, err := reader.Read(bits[:])
		if err != nil {
			return nil, err
		}
		bmask[int(chunk)] = binary.LittleEndian.Uint32(bits[:])
	}

	return &bmask, nil
}

func WriteBitmask(mask *Bitmask, descriptor *Descriptor, writer io.Writer) error {
	if err := writeUint8(writer, uint8(mask.Len())); err != nil {
		return err
	}

	for _, block := range *mask {
		if err := writeUint32(writer, block); err != nil {
			return err
		}
	}

	return nil
}

func (b *Bitmask) Enabled(offset uint32) bool {
	mask := *b

	// fail if out of bounds
	if offset/32 >= uint32(len(mask)) {
		return false
	}

	base := offset / 32
	bitIndex := offset % 32

	// check if the offset is toggled.
	return (mask[base] & (1 << bitIndex)) != 0
}

func (b *Bitmask) Set(offset uint32, value bool) {
	mask := *b
	// Suppose len(mask) = 2
	// and offset = 112

	// 3
	blockOffset := int(offset / 32)
	// 16
	bitOffset := int(offset % 32)

	if len(mask) <= blockOffset+1 {
		// mask len = 2 + (4-2) = 4
		mask = append(mask, make([]uint32, blockOffset+1-len(mask))...)
	}

	if value {
		mask[blockOffset] |= (1 << bitOffset)
	} else {
		mask[blockOffset] &= ^(1 << bitOffset)
	}
	*b = mask
}

type Quaternion struct {
	X, Y, Z, W float32
}

const (
	packCoeffYZ = 1 << 20
	packCoeffX  = 1 << 21
)

func DecodePackedQuaternion(input io.Reader) (Quaternion, error) {
	var q Quaternion

	uraw, err := readUint64(input)
	if err != nil {
		return q, err
	}

	raw := int64(uraw)

	x := float64(raw>>4) / packCoeffX
	y := float64(raw<<22>>43) / packCoeffYZ
	z := float64(raw<<43>>42) / packCoeffYZ
	w := 1 - (x*x + y*y + z*z)

	w = math.Sqrt(w)

	return Quaternion{
		float32(x),
		float32(y),
		float32(z),
		float32(w),
	}, nil
}

func (q Quaternion) EncodePacked(wr io.Writer) error {
	return writeUint64(wr, uint64(q.GetPacked()))
}

func (q Quaternion) GetPacked() int64 {
	var wSign int32
	if q.W >= 0 {
		wSign = 1
	} else {
		wSign = -1
	}

	x := int64(int32(int32(q.X)*packCoeffX*wSign) & ((1 << 22) - 1))
	y := int64(int32(int32(q.X)*packCoeffYZ*wSign) & ((1 << 21) - 1))
	z := int64(int32(int32(q.X)*packCoeffYZ*wSign) & ((1 << 21) - 1))

	raw := z | (y << 21) | (x << 42)
	return raw
}

func readBool(reader io.Reader) (bool, error) {
	var boolean [1]byte

	if _, err := reader.Read(boolean[:]); err != nil && err != io.EOF {
		return false, err
	}

	if boolean[0] > 0x01 {
		return false, fmt.Errorf("update: unexpected non-boolean value 0x%02X", boolean[0])
	}

	return boolean[0] == 1, nil
}

func readUint8(reader io.Reader) (uint8, error) {
	var byte [1]byte

	if _, err := reader.Read(byte[:]); err != nil && err != io.EOF {
		return 0, err
	}

	return byte[0], nil
}

func readUint16(reader io.Reader) (uint16, error) {
	var data [2]byte
	if _, err := reader.Read(data[:]); err != nil && err != io.EOF {
		return 0, err
	}

	return binary.LittleEndian.Uint16(data[:]), nil
}

func readUint32(reader io.Reader) (uint32, error) {
	var data [4]byte
	if _, err := reader.Read(data[:]); err != nil && err != io.EOF {
		return 0, err
	}

	return binary.LittleEndian.Uint32(data[:]), nil
}

func readUint64(reader io.Reader) (uint64, error) {
	var data [8]byte
	if _, err := reader.Read(data[:]); err != nil && err != io.EOF {
		return 0, err
	}

	return binary.LittleEndian.Uint64(data[:]), nil
}

func readFloat32(reader io.Reader) (float32, error) {
	u32, err := readUint32(reader)
	if err != nil && err != io.EOF {
		return 0, err
	}

	return math.Float32frombits(u32), nil
}

func writeUint8(writer io.Writer, value uint8) error {
	_, err := writer.Write([]byte{value})
	return err
}

func writeBool(writer io.Writer, value bool) error {
	var b uint8

	if value {
		b++
	}

	_, err := writer.Write([]byte{b})
	return err
}

func writeUint16(writer io.Writer, value uint16) error {
	var data [2]byte
	binary.LittleEndian.PutUint16(data[:], value)
	_, err := writer.Write(data[:])
	return err
}

func writeUint32(writer io.Writer, value uint32) error {
	var data [4]byte
	binary.LittleEndian.PutUint32(data[:], value)
	_, err := writer.Write(data[:])
	return err
}

func writeUint64(writer io.Writer, value uint64) error {
	var data [8]byte
	binary.LittleEndian.PutUint64(data[:], value)
	_, err := writer.Write(data[:])
	return err
}

func writeFloat32(writer io.Writer, value float32) error {
	return writeUint32(writer, math.Float32bits(value))
}

func readInt32(reader io.Reader) (int32, error) {
	u, err := readUint32(reader)
	return int32(u), err
}

func writeInt32(writer io.Writer, value int32) error {
	return writeUint32(writer, uint32(value))
}

func u32(b []byte) uint32 {
	return binary.LittleEndian.Uint32(b[:])
}

func nxtChunk(chunkOffset, bitOffset *uint32) {
	if *bitOffset > 0 {
		*bitOffset = 0
	}

	*chunkOffset++
}

func nxtByte(chunkOffset, bitOffset *uint32) {
	if *bitOffset%8 != 0 {
		*bitOffset += (8 - *bitOffset%8)
	} else {
		*bitOffset += 8
	}

	if *bitOffset == 32 {
		*bitOffset = 0
		*chunkOffset++
	}
}

func nxtBit(chunkOffset, bitOffset *uint32) {
	*bitOffset++
	if *bitOffset == 32 {
		*chunkOffset++
	}
}
