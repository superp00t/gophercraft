/* go.Zamara Library
 * Copyright (c) 2012, Erik Davidson
 * All rights reserved.
 *
 * Redistribution and use in source and binary forms, with or without
 * modification, are permitted provided that the following conditions
 * are met:
 *
 * 1. Redistributions of source code must retain the above copyright notice,
 *    this list of conditions and the following disclaimer.
 *
 * 2. Redistributions in binary form must reproduce the above copyright notice,
 *    this list of conditions and the following disclaimer in the documentation
 *    and/or other materials provided with the distribution.
 *
 * THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
 * AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
 * IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE
 * ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE
 * LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR
 * CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF
 * SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS
 * INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN
 * CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE)
 * ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE
 * POSSIBILITY OF SUCH DAMAGE.
 */

package mpq

import (
	"encoding/binary"
	"strings"
)

var blockEncryptionTable []uint32

func init() {
	generateEncryptionTable()
}

func generateEncryptionTable() {
	blockEncryptionTable = make([]uint32, 1280)

	var seed uint32 = 0x00100001

	for mainIdx := 0; mainIdx < 256; mainIdx++ {
		currentIdx := mainIdx

		for innerIdx := 0; innerIdx < 5; innerIdx++ {
			seed = (seed*125 + 3) % 0x2AAAAB
			temp1 := (seed & 0xFFFF) << 0x10

			seed = (seed*125 + 3) % 0x2AAAAB
			temp2 := (seed & 0xFFFF)

			blockEncryptionTable[currentIdx] = (temp1 | temp2)

			currentIdx += 0x100
		}
	}
}

func hashString(input string, offset uint16) (hash uint32) {
	var seed1 uint32 = 0x7FED7FED
	var seed2 uint32 = 0xEEEEEEEE

	str := strings.ToUpper(input)

	for _, curChar := range str {
		value := blockEncryptionTable[offset+uint16(curChar)]
		seed1 = (value ^ (seed1 + seed2)) & 0xFFFFFFFF
		seed2 = (uint32(curChar) + seed1 + seed2 + (seed2 << 5) + 3) & 0xFFFFFFFF
	}

	return seed1
}

func decrypt(seed1 uint32, table []byte) (err error) {
	var seed2 uint32 = 0xEEEEEEEE

	size := len(table)
	pos := 0
	for ; size >= 4; size -= 4 {
		seed2 += blockEncryptionTable[0x400+(seed1&0xFF)]
		curEntry := binary.LittleEndian.Uint32(table[pos : pos+4])
		entry := curEntry ^ (seed1 + seed2)
		seed1 = ((^seed1 << 0x15) + 0x11111111) | (seed1 >> 0x0B)
		seed2 = uint32(entry) + seed2 + (seed2 << 5) + 3

		binary.LittleEndian.PutUint32(table[pos:pos+4], entry)
		pos += 4
	}
	return
}
