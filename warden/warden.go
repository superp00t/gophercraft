package warden

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/superp00t/etc"

	"github.com/davecgh/go-spew/spew"
	"github.com/superp00t/etc/yo"
	"github.com/superp00t/gophercraft/arc4"
	"github.com/superp00t/gophercraft/packet"
)

type WardenModule struct {
	Hash              []byte
	Module            []byte
	ModuleKey         []byte
	Seed              []byte
	ServerKeySeed     []byte
	ClientKeySeed     []byte
	ClientKeySeedHash []byte
}

type Warden struct {
	InputCrypto, OutputCrypto *arc4.ARC4
	Module                    *WardenModule

	PreviousTimestamp, CurrentTimestamp, CheckTimer int64
}

func GetIdByAddr(addr uint32) uint16 {
	for _, v := range ChecksDB {
		if v.Address == int(addr) {
			return v.ID
		}
	}

	return 0
}

type PCheck struct {
	Type uint8
	ID   uint16
}

func UnmarshalWardenServerChecks(input []byte, xorByte uint8) ([]PCheck, error) {
	e := etc.FromBytes(input)

	if e.ReadByte() != packet.WARDEN_SMSG_CHEAT_CHECKS_REQUEST {
		return nil, fmt.Errorf("Invalid packet")
	}

	buf := new(bytes.Buffer)
	for {
		b := e.ReadByte()
		if b == 0x00 {
			break
		}
		buf.WriteByte(b)
	}

	yo.Println("Driver headers: ", spew.Sdump(buf.Bytes()))

	if e.ReadByte() == (packet.TIMING_CHECK ^ xorByte) {
		yo.Warn("TIMING_CHECK detected")
	}

	var ppc []PCheck
	for {
		tk := e.ReadByte()
		if tk == xorByte {
			break
		}

		t := tk ^ xorByte
		switch t {
		case packet.LUA_STR_CHECK:
			e.ReadByte()
			id := ppc[len(ppc)-1].ID + 1
			ppc = append(ppc, PCheck{t, id})
		case packet.PAGE_CHECK_B:
			e.ReadBytes(24)
			addr := e.ReadUint32()
			e.ReadByte()
			id := GetIdByAddr(addr)
			ppc = append(ppc, PCheck{t, id})
		case packet.MEM_CHECK:
			e.ReadByte()
			addr := e.ReadUint32()
			id := GetIdByAddr(addr)
			ppc = append(ppc, PCheck{t, id})
		case packet.DRIVER_CHECK:
			data := e.ReadBytes(24)
			e.ReadByte()
			dataStr := strings.ToUpper(hex.EncodeToString(data))
			for _, v := range ChecksDB {
				if v.Data == dataStr {
					ppc = append(ppc, PCheck{t, v.ID})
					break
				}
			}
		case packet.MODULE_CHECK:
			seed := e.ReadBytes(4)
			dig := e.ReadBytes(20)
			for _, v := range ChecksDB {
				if v.Type == packet.MODULE_CHECK {
					hm := hmac.New(sha1.New, seed)
					hm.Write([]byte(v.Str))
					d := hm.Sum(nil)

					if bytes.Equal(d, dig) {
						ppc = append(ppc, PCheck{t, v.ID})
						break
					}
				}
			}
		}
	}

	return ppc, nil
}
