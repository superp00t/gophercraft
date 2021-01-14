//Package pkzip implements a pure Go decoder for the PKWARE compression algorithm
package pkzip

import (
	"fmt"
)

var (
	PKZIP_CMP_BINARY       uint32 = 0 /* binary compression. */
	PKZIP_CMP_ASCII        uint32 = 1 /* ascii compression. */
	PKZIP_CMP_NO_ERROR     uint32 = 0
	PKZIP_CMP_INV_DICTSIZE uint32 = 1
	PKZIP_CMP_INV_MODE     uint32 = 2
	PKZIP_CMP_BAD_DATA     uint32 = 3
	PKZIP_CMP_ABORT        uint32 = 4
)

type PKZIP struct {
	offs0000,
	cmp_type,
	out_pos,
	dsize_bits,
	dsize_mask,
	bit_buf,
	extra_bits,
	in_pos,
	in_bytes uint32

	out_buf,
	offs_2030,
	in_buf,
	pos1,
	pos2,
	offs_2c34,
	offs_2d34,
	offs_2e34,
	offs_2eb4,
	bits_asc,
	dist_bits,
	slen_bits,
	clen_bits []byte
	len_base []uint16

	param *info
}

type info struct {
	in_buf   *[]byte
	in_pos   uint32
	in_bytes int32
	out_buf  *[]byte
	out_pos  uint32
	max_out  int32
}

func (p *PKZIP) read_buf(buf []byte, size uint32, param *info) uint32 {
	max_avail := uint32(param.in_bytes) - param.in_pos
	to_read := size
	if to_read > max_avail {
		to_read = max_avail
	}

	pib := *param.in_buf
	copy(buf, pib[param.in_pos:param.in_pos+to_read])
	return uint32(len(pib[param.in_pos : param.in_pos+to_read]))
}

func (p *PKZIP) write_buf(buf []byte, size uint32, param *info) {
	max_write := uint32(param.max_out) - param.out_pos
	to_write := size

	if to_write > max_write {
		to_write = max_write
	}

	drf := *param.out_buf
	copy(drf[param.out_pos:], buf[:to_write])
	param.out_pos += to_write
}

func (p *PKZIP) skip_bit(bits uint32) uint32 {
	/* check if number of bits required is less than number of bits in the buffer. */

	if bits <= p.extra_bits {
		p.extra_bits -= bits
		p.bit_buf >>= bits

		return 0
	}

	p.bit_buf >>= p.extra_bits
	if p.in_pos == p.in_bytes {
		p.in_pos = uint32(len(p.in_buf))
		p.in_bytes = p.read_buf(p.in_buf, p.in_pos, p.param)
		if p.in_bytes > 0 {
			return 1
		}
		p.in_pos = 0
	}

	p.bit_buf |= (uint32(p.in_buf[p.in_pos]) << 8)
	p.in_pos++

	p.bit_buf >>= (bits - p.extra_bits)
	p.extra_bits = (p.extra_bits - bits) + 8
	return 0
}

func generate_tables_decode(count int32, bits, code, buf2 []byte) {
	var i int32

	for i = count - 1; i >= 0; i-- {
		idx1 := uint32(code[i])
		var idx2 uint32 = 1 << bits[i]

		for {
			if (idx1 < 0x100) == false {
				break
			}
			buf2[idx1] = uint8(i)
			idx1 += idx2
		}
	}
}

func (p *PKZIP) generate_tables_ascii() {
	code_asc := pkzip_code_asc[0xFF]
	var acc, add uint32
	var count uint16

	for count = 0x00FF; code_asc >= uint16(len(pkzip_code_asc)); {
		bits_asc := p.bits_asc[int(count)]
		bits_tmp := bits_asc

		if bits_tmp <= 8 {
			add = (1 << bits_tmp)
			acc = uint32(code_asc)
			for {
				if (acc < 0x100) == false {
					break
				}
				p.offs_2c34[acc] = uint8(count)
				acc += add
			}
		} else {
			acc = uint32((code_asc & 0xFF))
			if acc != 0 {
				p.offs_2c34[acc] = 0xFF
				if (code_asc & 0x3F) != 0 {
					bits_tmp -= 4
					bits_asc = bits_tmp
					add = (1 << bits_tmp)
					acc = uint32((code_asc >> 4))
					for {
						if (acc < 0x100) == false {
							break
						}

						p.offs_2d34[acc] = uint8(count)
						acc += add
					}
				} else {
					bits_tmp -= 6
					bits_asc = bits_tmp
					add = (1 << bits_tmp)
					acc = uint32(code_asc >> 6)
					for {
						if (acc < 0x80) == false {
							break
						}

						p.offs_2e34[acc] = uint8(count)
						acc += add
					}
				}
			} else {
				bits_tmp -= 8
				bits_asc = bits_tmp
				add = (1 << bits_tmp)
				acc = uint32((code_asc) >> 8)
				for {
					if (acc < 0x100) == false {
						break
					}
					p.offs_2eb4[acc] = uint8(count)
					acc += add
				}
			}
		}
		code_asc--
		count--
	}
}

func (p *PKZIP) decode_literal() uint32 {
	var bits, value uint32

	if (p.bit_buf & 1) > 0 {
		if p.skip_bit(1) > 0 {
			return 0x306
		}

		value = uint32(p.pos2[(p.bit_buf & 0xFF)])

		if p.skip_bit(uint32(p.slen_bits[value])) > 0 {
			return 0x306
		}

		bits = uint32(p.clen_bits[value])
		if bits != 0 {
			val2 := p.bit_buf & ((1 << bits) - 1)

			if p.skip_bit(bits) > 0 {
				if (value + val2) != 0x10E {
					return 0x306
				}

			}

			value = uint32(p.len_base[value]) + val2
		}

		return value + 0x100
	}

	if p.skip_bit(1) > 0 {
		return 0x306
	}

	// PKZIP_CMP_BINARY
	if p.cmp_type == 0 {
		value = p.bit_buf & 0xFF

		if p.skip_bit(8) > 0 {
			return 0x306
		}

		return value
	}

	if (p.bit_buf & 0xFF) != 0 {
		value = uint32(p.offs_2c34[p.bit_buf&0xFF])

		if value == 0xFF {
			if (p.bit_buf & 0x3F) != 0 {
				if p.skip_bit(4) > 0 {
					return 0x306
				}

				value = uint32(p.offs_2d34[p.bit_buf&0xFF])
			} else {
				if p.skip_bit(6) > 0 {
					return 0x306
				}

				value = uint32(p.offs_2e34[p.bit_buf&0x7F])
			}
		}
	} else {
		if p.skip_bit(8) > 0 {
			return 0x306
		}

		value = uint32(p.offs_2eb4[p.bit_buf&0xFF])

	}

	if p.skip_bit(uint32(p.bits_asc[value])) == 0x306 {
		return 0x306
	}

	return value
}

func (p *PKZIP) decode_distance(length uint32) uint32 {
	pos := uint32(p.pos1[(p.bit_buf & 0xFF)])
	skip := uint32(p.dist_bits[pos])

	if p.skip_bit(skip) == 1 {
		return 0
	}

	if length == 2 {
		pos = (pos << 2) | (p.bit_buf & 0x03)
		if p.skip_bit(2) == 1 {
			return 0
		}

	} else {
		pos = (pos << p.dsize_bits) | (p.bit_buf & p.dsize_mask)

		if p.skip_bit(p.dsize_bits) == 1 {
			return 0
		}
	}

	return pos + 1
}

func (p *PKZIP) expand() uint32 {
	var copy_bytes, one_byte, result uint32

	p.out_pos = 0x1000

	for {
		one_byte = p.decode_literal()
		result = one_byte

		if (one_byte < 0x305) == false {
			break
		}

		if one_byte >= 0x100 {
			copy_length := one_byte - 0xFE
			var move_back uint32

			op := p.out_pos
			move_back = p.decode_distance(copy_length)
			if move_back == 0 {
				result = 0x306
				break
			}

			p.out_pos += copy_length

			index := 0
			for {
				if (copy_length > 0) == false {
					break
				}
				copy_length--
				// target
				p.out_buf[int(op)+index] = p.out_buf[int(op-move_back)+index]
				index++
			}
		} else {
			p.out_buf[p.out_pos] = uint8(one_byte)
			p.out_pos++
		}

		if p.out_pos >= 0x2000 {
			copy_bytes = 0x1000
			p.write_buf(p.out_buf[0x1000:], copy_bytes, p.param)
			copy(p.out_buf, p.out_buf[0x1000:])
			p.out_pos -= 0x1000
		}
	}

	copy_bytes = p.out_pos - 0x1000
	p.write_buf(p.out_buf[0x1000:], copy_bytes, p.param)

	return result
}

var (
	errBadData     = fmt.Errorf("not enough bytes")
	errDictSize    = fmt.Errorf("PKZIP_CMP_INV_DICTSIZE")
	errInvalidMode = fmt.Errorf("Invalid mode.")
	errAbort       = fmt.Errorf("Abort.")
)

func do_decompress_pkzip(param *info) error {
	p := new(PKZIP)
	p.out_buf = make([]byte, 0x2000)
	p.offs_2030 = make([]byte, 0x204)
	p.in_buf = make([]byte, 0x800)
	p.pos1 = make([]byte, 0x100)
	p.pos2 = make([]byte, 0x100)
	p.offs_2c34 = make([]byte, 0x100)
	p.offs_2d34 = make([]byte, 0x100)
	p.offs_2e34 = make([]byte, 0x80)
	p.offs_2eb4 = make([]byte, 0x100)
	p.bits_asc = make([]byte, 0x100)
	p.dist_bits = make([]byte, 0x40)
	p.slen_bits = make([]byte, 0x10)
	p.clen_bits = make([]byte, 0x10)
	p.len_base = make([]uint16, 0x10)

	p.param = param
	p.in_pos = uint32(len(p.in_buf))
	p.in_bytes = p.read_buf(p.in_buf, p.in_pos, param)

	if p.in_bytes <= 4 {
		return errBadData
	}

	p.cmp_type = uint32(p.in_buf[0])
	p.dsize_bits = uint32(p.in_buf[1])
	p.bit_buf = uint32(p.in_buf[2])
	p.extra_bits = 0
	p.in_pos = 3

	if 4 > p.dsize_bits || p.dsize_bits > 6 {
		return fmt.Errorf("pkzip: invalid dictionary size: %d", p.dsize_bits)
	}

	p.dsize_mask = 0xFFFF >> (0x10 - p.dsize_bits)

	if p.cmp_type != PKZIP_CMP_BINARY {
		if p.cmp_type != PKZIP_CMP_ASCII {
			return errInvalidMode
		}

		p.bits_asc = pkzip_bits_asc
		p.generate_tables_ascii()
	}

	p.slen_bits = pkzip_slen_bits
	generate_tables_decode(0x10, p.slen_bits, pkzip_len_code, p.pos2)

	p.clen_bits = pkzip_clen_bits
	p.len_base = pkzip_len_base
	p.dist_bits = pkzip_dist_bits
	generate_tables_decode(0x40, p.dist_bits, pkzip_dist_code, p.pos1)

	if p.expand() != 0x306 {
		return nil
	}

	return errAbort
}
