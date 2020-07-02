package packet

import (
	"encoding/binary"
	"fmt"
	"io"

	"github.com/superp00t/gophercraft/vsn"
)

const (
	GroupNormal = 0
	GroupRaid   = 1
)

const (
	MemberOffline = 0x0000
	MemberOnline  = 0x0001 // Lua_UnitIsConnected
	MemberPVP     = 0x0002 // Lua_UnitIsPVP
	MemberDead    = 0x0004 // Lua_UnitIsDead
	MemberGhost   = 0x0008 // Lua_UnitIsGhost
	MemberPVPFFA  = 0x0010 // Lua_UnitIsPVPFreeForAll
	MemberZoneOut = 0x0020 // Lua_GetPlayerMapPosition
	MemberAFK     = 0x0040 // Lua_UnitIsAFK
	MemberDND     = 0x0080 // Lua_UnitIsDND
)

const (
	GroupUpdateNone            = 0x00000000 // nothing
	GroupUpdateStatus          = 0x00000001 // uint16, flags
	GroupUpdateCurrentHealth   = 0x00000002 // uint32
	GroupUpdateMaxHealth       = 0x00000004 // uint32
	GroupUpdatePowerType       = 0x00000008 // uint8
	GroupUpdateCurrentPower    = 0x00000010 // uint16
	GroupUpdateMaxPower        = 0x00000020 // uint16
	GroupUpdateLevel           = 0x00000040 // uint16
	GroupUpdateZone            = 0x00000080 // uint16
	GroupUpdatePosition        = 0x00000100 // uint16, uint16
	GroupUpdateAuras           = 0x00000200 // uint64 mask, for each bit set uint32 spellid + uint8 unk
	GroupUpdatePetGUID         = 0x00000400 // uint64 pet guid
	GroupUpdatePetName         = 0x00000800 // pet name, nullptr terminated string
	GroupUpdatePetModelID      = 0x00001000 // uint16, model id
	GroupUpdatePetCurrentHP    = 0x00002000 // uint32 pet cur health
	GroupUpdatePetMaxHP        = 0x00004000 // uint32 pet max health
	GroupUpdatePetPowerType    = 0x00008000 // uint8 pet power type
	GroupUpdatePetCurrentPower = 0x00010000 // uint16 pet cur power
	GroupUpdatePetMaxPower     = 0x00020000 // uint16 pet max power
	GroupUpdatePetAuras        = 0x00040000 // uint64 mask, for each bit set uint32 spellid + uint8 unk, pet auras...
	GroupUpdateVehicleSeat     = 0x00080000 // uint32 vehicle_seat_id (index from VehicleSeat.dbc)
	GroupUpdatePet             = 0x0007FC00 // all pet flags
	GroupUpdateFull            = 0x0007FFFF // all known flags
)

type PartyOperation uint32

const (
	PartyInvite PartyOperation = 0
	PartyLeave  PartyOperation = 2
	PartySwap   PartyOperation = 4
)

type PartyResult uint32

const (
	PartyOK PartyResult = iota
	PartyBadPlayerName
	PartyTargetNotInGroup
	PartyGroupFull
	PartyAlreadyInGroup
	PartyNotInGroup
	PartyNotLeader
	PartyWrongFaction
	PartyIgnoringYou
)

var PartyResultDescriptors = map[vsn.Build]map[PartyResult]uint32{
	5875: {
		PartyOK:               0,
		PartyBadPlayerName:    1,
		PartyTargetNotInGroup: 2,
		PartyGroupFull:        3,
		PartyAlreadyInGroup:   4,
		PartyNotInGroup:       5,
		PartyNotLeader:        6,
		PartyWrongFaction:     7,
		PartyIgnoringYou:      8,
	},
}

func EncodePartyResult(build vsn.Build, out io.Writer, pr PartyResult) error {
	desc, ok := PartyResultDescriptors[build]
	if !ok {
		return fmt.Errorf("packet: no PartyResult descriptor for %s", build)
	}

	op, ok := desc[pr]
	if !ok {
		return fmt.Errorf("packet: no PartyResult code found for %v in %s", pr, build)
	}

	var u32 [4]byte
	binary.LittleEndian.PutUint32(u32[:], op)

	_, err := out.Write(u32[:])
	return err
}
