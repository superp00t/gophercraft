package auth

import (
	"fmt"

	"github.com/superp00t/etc"
	"github.com/superp00t/gophercraft/gcore"
	"github.com/superp00t/gophercraft/gcore/config"
	"github.com/superp00t/gophercraft/vsn"
)

var RealmList_C = []byte{
	uint8(REALM_LIST),
	0x00,
	0x00,
	0x00,
	0x00,
}

type RealmListing struct {
	Type       config.RealmType //
	Locked     bool
	Flags      uint8
	Name       string
	Address    string
	Population float32
	Characters uint8
	Timezone   uint8
	ID         uint8
}

type RealmList_S struct {
	Realms []RealmListing
}

func MakeRealmlist(rlst []gcore.Realm) *RealmList_S {
	realmState := &RealmList_S{}

	for _, realm := range rlst {
		realmListing := RealmListing{}
		realmListing.Type = realm.Type
		realmListing.Locked = realm.Locked
		if realm.Offline() {
			realmListing.Flags |= 0x02
		}
		realmListing.Name = realm.Name
		realmListing.Address = realm.Address
		realmListing.Population = float32(realm.ActivePlayers/8000) * 3.0
		realmListing.Characters = 0 // todo: get character count from database map
		realmListing.Timezone = uint8(realm.Timezone)
		realmListing.ID = uint8(realm.ID)
		realmState.Realms = append(realmState.Realms, realmListing)
	}

	return realmState
}

func (rlst *RealmList_S) Encode(build vsn.Build) []byte {
	listBuffer := etc.NewBuffer()
	listBuffer.WriteUint32(0)

	if build.AddedIn(vsn.V2_4_3) {
		listBuffer.WriteUint16(uint16(len(rlst.Realms)))
	} else {
		listBuffer.WriteByte(uint8(len(rlst.Realms)))
	}

	for _, v := range rlst.Realms {
		if build.AddedIn(vsn.V2_4_3) {
			listBuffer.WriteByte(uint8(v.Type))
			listBuffer.WriteBool(v.Locked)
		} else {
			listBuffer.WriteUint32(uint32(v.Type))
		}

		listBuffer.WriteByte(v.Flags)
		listBuffer.WriteCString(v.Name)
		listBuffer.WriteCString(v.Address)
		listBuffer.WriteFloat32(v.Population)
		listBuffer.WriteByte(v.Characters) // TODO: character count has to be included
		listBuffer.WriteByte(uint8(v.Timezone))
		listBuffer.WriteByte(uint8(v.ID))
	}

	if build.AddedIn(vsn.V2_4_3) {
		listBuffer.WriteUint16(0x10)
	} else {
		listBuffer.WriteUint16(0x02)
	}

	head := etc.NewBuffer()
	head.WriteByte(uint8(REALM_LIST))
	head.WriteUint16(uint16(listBuffer.Len()))
	head.Write(listBuffer.Bytes())
	return head.Bytes()
}

func UnmarshalRealmList_S(build vsn.Build, input []byte) (*RealmList_S, error) {
	header := etc.FromBytes(input)

	rls := &RealmList_S{}
	cmd := AuthType(header.ReadByte())
	if cmd != REALM_LIST {
		return nil, fmt.Errorf("wrong type %s", cmd)
	}
	size := header.ReadUint16()

	in := etc.FromBytes(header.ReadBytes(int(size)))

	if build.RemovedIn(vsn.V2_4_3) {
		in.ReadUint32()
	}

	numRealms := in.ReadByte()

	for x := uint8(0); x < numRealms; x++ {
		rlst := RealmListing{}
		if build.RemovedIn(vsn.V2_4_3) {
			rlst.Type = config.RealmType(in.ReadUint32())
		} else {
			rlst.Type = config.RealmType(in.ReadByte())
			rlst.Locked = in.ReadBool()
		}

		rlst.Flags = in.ReadByte()
		rlst.Name = in.ReadCString()
		rlst.Address = in.ReadCString()
		rlst.Population = in.ReadFloat32()
		rlst.Characters = in.ReadByte()
		rlst.Timezone = in.ReadByte()
		rlst.ID = uint8(in.ReadByte())
		rls.Realms = append(rls.Realms, rlst)
	}

	return rls, nil
}
