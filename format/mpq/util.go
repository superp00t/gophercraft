package mpq

import "io"
import "encoding/binary"

func lu16(i io.Reader) uint16 {
	buf := lb(i, 2)
	return binary.LittleEndian.Uint16(buf)
}

func lu32(i io.Reader) uint32 {
	buf := lb(i, 4)
	return binary.LittleEndian.Uint32(buf)
}

func lu64(i io.Reader) uint64 {
	buf := lb(i, 8)
	return binary.LittleEndian.Uint64(buf)
}

func lb(i io.Reader, l int) []byte {
	buf := make([]byte, l)
	_, err := io.ReadAtLeast(i, buf, l)
	if err != nil {
		panic(err)
	}
	return buf
}

func _BSWAP_ARRAY32_UNSIGNED(ptr []uint32) {
	for i, v := range ptr {
		ptr[i] = swapbytes32(v)
	}
}

func swapbytes8(d uint8) uint8 {
	return (d << 4) | (d >> 4)
}

func swapbytes16(d uint16) uint16 {
	return (d << 8) | (d >> 8)
}

func swapbytes32(d uint32) uint32 {
	return (d << 16) | (d >> 16)
}

func swapbytes64(d uint64) uint64 {
	return (d << 32) | (d >> 32)
}
