package packet

import (
	"fmt"

	"github.com/superp00t/etc"
)

var RealmList_C = []byte{
	uint8(REALM_LIST),
	0x00,
	0x00,
	0x00,
	0x00,
}

type RealmList_S struct {
	Cmd    AuthType
	Realms []RealmListing
}

type RealmType uint8

const (
	Normal  RealmType = 0
	PVP     RealmType = 1
	Normal2 RealmType = 4
	RP      RealmType = 6
	PVPRP   RealmType = 8
	FFARP   RealmType = 16
)

func ConvertRealmType(input string) RealmType {
	switch input {
	case "normal", "":
		return Normal
	case "pvp":
		return PVP
	case "pvp-rp":
		return PVPRP
	case "rp":
		return RP
	case "ffa-rp":
		return FFARP
	}

	return Normal
}

type RealmListing struct {
	Type       RealmType //
	Locked     bool
	Flags      uint8
	Name       string
	Address    string
	Population float32
	Characters uint8
	Timezone   uint8
	ID         uint8
}

func (rlst *RealmList_S) Encode(version uint32) []byte {
	listBuffer := etc.NewBuffer()
	listBuffer.WriteUint32(0)
	listBuffer.WriteByte(uint8(len(rlst.Realms)))

	for _, v := range rlst.Realms {
		if version == 5875 {
			listBuffer.WriteUint32(uint32(v.Type))
		} else {
			listBuffer.WriteByte(uint8(v.Type))
			listBuffer.WriteBool(v.Locked)
		}

		listBuffer.WriteByte(v.Flags)
		listBuffer.WriteCString(v.Name)
		listBuffer.WriteCString(v.Address)
		listBuffer.WriteFloat32(v.Population)
		listBuffer.WriteByte(v.Characters)
		listBuffer.WriteByte(v.Timezone)
		listBuffer.WriteByte(v.ID)
	}

	if version == 5875 {
		listBuffer.WriteUint16(0x0002)
	} else {
		listBuffer.WriteByte(0x10)
		listBuffer.WriteByte(0x00)
	}

	head := etc.NewBuffer()
	head.WriteByte(uint8(rlst.Cmd))
	head.WriteUint16(uint16(listBuffer.Len()))
	head.Write(listBuffer.Bytes())
	return head.Bytes()
}

func UnmarshalRealmList_S(build uint32, input []byte) (*RealmList_S, error) {
	header := etc.FromBytes(input)

	rls := &RealmList_S{}
	rls.Cmd = AuthType(header.ReadByte())
	size := header.ReadUint16()

	in := etc.FromBytes(header.ReadBytes(int(size)))

	if build == 5875 {
		in.ReadUint32()
	}

	numRealms := in.ReadByte()

	if rls.Cmd != REALM_LIST {
		return nil, fmt.Errorf("packet: request type is %s, not REALM_LIST", rls.Cmd)
	}

	for x := uint8(0); x < numRealms; x++ {
		rlst := RealmListing{}
		if build == 5875 {
			rlst.Type = RealmType(in.ReadUint32())
		} else {
			rlst.Type = RealmType(in.ReadByte())
			rlst.Locked = in.ReadBool()
		}

		rlst.Flags = in.ReadByte()
		rlst.Name = in.ReadCString()
		rlst.Address = in.ReadCString()
		rlst.Population = in.ReadFloat32()
		rlst.Characters = in.ReadByte()
		rlst.Timezone = in.ReadByte()
		rlst.ID = in.ReadByte()
		rls.Realms = append(rls.Realms, rlst)
	}

	return rls, nil
}
