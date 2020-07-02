package packet

import (
	"encoding/binary"
	"fmt"
	"io"

	"github.com/superp00t/etc"
	"github.com/superp00t/gophercraft/guid"
	"github.com/superp00t/gophercraft/vsn"
)

//go:generate gcraft_stringer -type=Race
type Race uint8

const (
	RACE_NONE               Race = 0
	RACE_HUMAN              Race = 1
	RACE_ORC                Race = 2
	RACE_DWARF              Race = 3
	RACE_NIGHTELF           Race = 4
	RACE_UNDEAD_PLAYER      Race = 5
	RACE_TAUREN             Race = 6
	RACE_GNOME              Race = 7
	RACE_TROLL              Race = 8
	RACE_GOBLIN             Race = 9
	RACE_BLOODELF           Race = 10
	RACE_DRAENEI            Race = 11
	RACE_FEL_ORC            Race = 12
	RACE_NAGA               Race = 13
	RACE_BROKEN             Race = 14
	RACE_SKELETON           Race = 15
	RACE_VRYKUL             Race = 16
	RACE_TUSKARR            Race = 17
	RACE_FOREST_TROLL       Race = 18
	RACE_TAUNKA             Race = 19
	RACE_NORTHREND_SKELETON Race = 20
	RACE_ICE_TROLL          Race = 21
	RACE_WORGEN             Race = 22
	// RACE_GILNEAN        Race    = 23
	RACE_PANDAREN_NEUTRAL    Race = 24
	RACE_PANDAREN_ALLIANCE   Race = 25
	RACE_PANDAREN_HORDE      Race = 26
	RACE_NIGHTBORNE          Race = 27
	RACE_HIGHMOUNTAIN_TAUREN Race = 28
	RACE_VOID_ELF            Race = 29
	RACE_LIGHTFORGED_DRAENEI Race = 30
	RACE_ZANDALARI_TROLL     Race = 31
	RACE_KUL_TIRAN           Race = 32
	RACE_THIN_HUMAN          Race = 33
	RACE_DARK_IRON_DWARF     Race = 34
	//RACE_VULPERA            = 35,
	RACE_MAGHAR_ORC Race = 36

	GENDER_MALE   = 0
	GENDER_FEMALE = 1
	GENDER_NONE   = 2

	MANA      = 0
	RAGE      = 1
	FOCUS     = 2
	ENERGY    = 3
	HAPPINESS = 4
)

//go:generate gcraft_stringer -type=Class
type Class uint8

const (
	CLASS_NONE         Class = 0
	CLASS_WARRIOR      Class = 1
	CLASS_PALADIN      Class = 2
	CLASS_HUNTER       Class = 3
	CLASS_ROGUE        Class = 4
	CLASS_PRIEST       Class = 5
	CLASS_DEATH_KNIGHT Class = 6
	CLASS_SHAMAN       Class = 7
	CLASS_MAGE         Class = 8
	CLASS_WARLOCK      Class = 9
	CLASS_MONK         Class = 10
	CLASS_DRUID        Class = 11
	CLASS_DEMON_HUNTER Class = 12
)

type Character struct {
	GUID       guid.GUID
	Name       string
	Race       Race
	Class      Class
	Gender     uint8
	Skin       uint8
	Face       uint8
	HairStyle  uint8
	HairColor  uint8
	FacialHair uint8
	Level      uint8
	Zone       uint32
	Map        uint32
	X, Y, Z    float32
	Guild      uint32
	Flags      CharacterFlags

	Customization uint32
	FirstLogin    uint8

	PetModel, PetLevel, PetFamily uint32

	Equipment []Item
}

type Item struct {
	Model       uint32
	Type        uint8
	Enchantment uint32
}

type CharacterList struct {
	Characters []Character
}

func (c *CharacterList) Packet(version vsn.Build) *WorldPacket {
	p := NewWorldPacket(SMSG_CHAR_ENUM)
	p.WriteByte(uint8(len(c.Characters)))
	for _, v := range c.Characters {
		v.GUID.EncodeUnpacked(version, p)
		p.WriteCString(v.Name)
		p.WriteByte(uint8(v.Race))
		p.WriteByte(uint8(v.Class))
		p.WriteByte(v.Gender)
		p.Write([]byte{v.Skin, v.Face, v.HairStyle, v.HairColor})
		p.WriteByte(v.FacialHair)
		p.WriteByte(v.Level)
		p.WriteUint32(v.Zone)
		p.WriteUint32(v.Map)
		p.WriteFloat32(v.X)
		p.WriteFloat32(v.Y)
		p.WriteFloat32(v.Z)
		p.WriteUint32(v.Guild)
		if err := EncodeCharacterFlags(version, p, v.Flags); err != nil {
			panic(err)
		}

		if version == 12340 {
			p.WriteUint32(v.Customization)
		}
		p.WriteByte(v.FirstLogin)
		p.WriteUint32(v.PetModel)
		p.WriteUint32(v.PetLevel)
		p.WriteUint32(v.PetFamily)

		for i := 0; i < EquipLen(version); i++ {
			var item Item

			if i < len(v.Equipment) {
				item = v.Equipment[i]
			}

			p.WriteUint32(item.Model)
			p.WriteByte(item.Type)

			if version >= 12340 {
				p.WriteUint32(item.Enchantment)
			}
		}

		if version == 5875 {
			// Bags
			p.WriteUint32(0)
			p.WriteByte(0)
		}
	}
	return p
}

func UnmarshalCharacterList(build vsn.Build, input []byte) (*CharacterList, error) {
	pkt := etc.FromBytes(input)
	count := int(pkt.ReadByte())
	var chh CharacterList
	for x := 0; x < count; x++ {
		ch := Character{}
		var err error
		ch.GUID, err = guid.DecodePacked(build, pkt)
		if err != nil {
			return nil, err
		}
		ch.Name = pkt.ReadCString()
		ch.Race = Race(pkt.ReadByte())
		ch.Class = Class(pkt.ReadByte())
		ch.Gender = pkt.ReadByte()
		ch.Skin = pkt.ReadByte()
		ch.Face = pkt.ReadByte()
		ch.HairStyle = pkt.ReadByte()
		ch.HairColor = pkt.ReadByte()
		ch.FacialHair = pkt.ReadByte()
		ch.Level = pkt.ReadByte()
		ch.Zone = pkt.ReadUint32()
		ch.Map = pkt.ReadUint32()
		ch.X = pkt.ReadFloat32()
		ch.Y = pkt.ReadFloat32()
		ch.Z = pkt.ReadFloat32()
		ch.Guild = pkt.ReadUint32()
		ch.Flags, err = DecodeCharacterFlags(build, pkt)
		if err != nil {
			return nil, err
		}
		if build >= 12340 {
			ch.Customization = pkt.ReadUint32()
		}
		ch.FirstLogin = pkt.ReadByte()
		ch.PetModel = pkt.ReadUint32()
		ch.PetLevel = pkt.ReadUint32()
		ch.PetFamily = pkt.ReadUint32()

		// Get equipment
		for j := 0; j < EquipLen(build); j++ {
			model := pkt.ReadUint32()
			typ := pkt.ReadByte()
			item := Item{
				Model: model,
				Type:  typ,
			}
			if build >= 12340 {
				item.Enchantment = pkt.ReadUint32()
			}
			ch.Equipment = append(ch.Equipment, item)
		}

		if build == 5875 {
			//bags
			pkt.ReadUint32()
			pkt.ReadByte()
		}

		chh.Characters = append(chh.Characters, ch)
	}
	return &chh, nil
}

func EquipLen(build vsn.Build) int {
	switch build {
	case 5875:
		return 19
	default:
		return 23
	}
}

//go:generate gcraft_stringer -type=CharacterOp
type CharacterOp uint8

const (
	CHAR_CREATE_IN_PROGRESS CharacterOp = iota
	CHAR_CREATE_SUCCESS
	CHAR_CREATE_ERROR
	CHAR_CREATE_FAILED
	CHAR_CREATE_NAME_IN_USE
	CHAR_CREATE_DISABLED
	CHAR_CREATE_PVP_TEAMS_VIOLATION
	CHAR_CREATE_SERVER_LIMIT
	CHAR_CREATE_ACCOUNT_LIMIT
	CHAR_CREATE_SERVER_QUEUE
	CHAR_CREATE_ONLY_EXISTING
	CHAR_CREATE_EXPANSION
	CHAR_CREATE_EXPANSION_CLASS
	CHAR_CREATE_LEVEL_REQUIREMENT
	CHAR_CREATE_UNIQUE_CLASS_LIMIT
	CHAR_CREATE_CHARACTER_IN_GUILD
	CHAR_CREATE_RESTRICTED_RACECLASS
	CHAR_CREATE_CHARACTER_CHOOSE_RACE
	CHAR_CREATE_CHARACTER_ARENA_LEADER
	CHAR_CREATE_CHARACTER_DELETE_MAIL
	CHAR_CREATE_CHARACTER_SWAP_FACTION
	CHAR_CREATE_CHARACTER_RACE_ONLY
	CHAR_CREATE_CHARACTER_GOLD_LIMIT
	CHAR_CREATE_FORCE_LOGIN
	CHAR_NAME_SUCCESS
	CHAR_NAME_FAILURE
	CHAR_NAME_NO_NAME
	CHAR_NAME_TOO_SHORT
	CHAR_NAME_TOO_LONG
	CHAR_NAME_INVALID_CHARACTER
	CHAR_NAME_MIXED_LANGUAGES
	CHAR_NAME_PROFANE
	CHAR_NAME_RESERVED
	CHAR_NAME_INVALID_APOSTROPHE
	CHAR_NAME_MULTIPLE_APOSTROPHES
	CHAR_NAME_THREE_CONSECUTIVE
	CHAR_NAME_INVALID_SPACE
	CHAR_NAME_CONSECUTIVE_SPACES
	CHAR_NAME_RUSSIAN_CONSECUTIVE_SILENT_CHARACTERS
	CHAR_NAME_RUSSIAN_SILENT_CHARACTER_AT_BEGINNING_OR_END
	CHAR_NAME_DECLENSION_DOESNT_MATCH_BASE_NAME
	CHAR_DELETE_IN_PROGRESS
	CHAR_DELETE_SUCCESS
	CHAR_DELETE_FAILED
	CHAR_DELETE_FAILED_LOCKED_FOR_TRANSFER
	CHAR_DELETE_FAILED_GUILD_LEADER
	CHAR_DELETE_FAILED_ARENA_CAPTAIN
)

type CharacterOpDescriptor map[CharacterOp]uint8

var CharacterOpDescriptors = map[vsn.Build]CharacterOpDescriptor{
	5875: {
		CHAR_CREATE_IN_PROGRESS:         0x2D,
		CHAR_CREATE_SUCCESS:             0x2E,
		CHAR_CREATE_ERROR:               0x2F,
		CHAR_CREATE_FAILED:              0x30,
		CHAR_CREATE_NAME_IN_USE:         0x31,
		CHAR_CREATE_DISABLED:            0x3A,
		CHAR_CREATE_PVP_TEAMS_VIOLATION: 0x33,
		CHAR_CREATE_SERVER_LIMIT:        0x34,
		CHAR_CREATE_ACCOUNT_LIMIT:       0x35,
		CHAR_CREATE_SERVER_QUEUE:        0x30, /// UNSURE
		CHAR_CREATE_ONLY_EXISTING:       0x30, /// UNSURE

		CHAR_DELETE_IN_PROGRESS:                0x38,
		CHAR_DELETE_SUCCESS:                    0x39,
		CHAR_DELETE_FAILED:                     0x3A,
		CHAR_DELETE_FAILED_LOCKED_FOR_TRANSFER: 0x3A, /// UNSURE
		CHAR_DELETE_FAILED_GUILD_LEADER:        0x3A, /// UNSURE

		/*CHAR_LOGIN_IN_PROGRESS                                 : 0x3B,
		  CHAR_LOGIN_SUCCESS                                     : 0x3C,
		  CHAR_LOGIN_NO_WORLD                                    : 0x3D,
		  CHAR_LOGIN_DUPLICATE_CHARACTER                         : 0x3E,
		  CHAR_LOGIN_NO_INSTANCES                                : 0x3F,
		  CHAR_LOGIN_FAILED                                      : 0x40,
		  CHAR_LOGIN_DISABLED                                    : 0x41,
		  CHAR_LOGIN_NO_CHARACTER                                : 0x42,
		  CHAR_LOGIN_LOCKED_FOR_TRANSFER                         : 0x40, /// UNSURE
		  CHAR_LOGIN_LOCKED_BY_BILLING                           : 0x40, /// UNSURE*/

		CHAR_NAME_SUCCESS:                                      0x50,
		CHAR_NAME_FAILURE:                                      0x4F,
		CHAR_NAME_NO_NAME:                                      0x43,
		CHAR_NAME_TOO_SHORT:                                    0x44,
		CHAR_NAME_TOO_LONG:                                     0x45,
		CHAR_NAME_INVALID_CHARACTER:                            0x46,
		CHAR_NAME_MIXED_LANGUAGES:                              0x47,
		CHAR_NAME_PROFANE:                                      0x48,
		CHAR_NAME_RESERVED:                                     0x49,
		CHAR_NAME_INVALID_APOSTROPHE:                           0x4A,
		CHAR_NAME_MULTIPLE_APOSTROPHES:                         0x4B,
		CHAR_NAME_THREE_CONSECUTIVE:                            0x4C,
		CHAR_NAME_INVALID_SPACE:                                0x4D,
		CHAR_NAME_CONSECUTIVE_SPACES:                           0x4E,
		CHAR_NAME_RUSSIAN_CONSECUTIVE_SILENT_CHARACTERS:        0x4E, /// UNSURE
		CHAR_NAME_RUSSIAN_SILENT_CHARACTER_AT_BEGINNING_OR_END: 0x4E, /// UNSURE
		CHAR_NAME_DECLENSION_DOESNT_MATCH_BASE_NAME:            0x4E, /// UNSURE
	},

	12340: {
		CHAR_CREATE_IN_PROGRESS:                                0x2E,
		CHAR_CREATE_SUCCESS:                                    0x2F,
		CHAR_CREATE_ERROR:                                      0x30,
		CHAR_CREATE_FAILED:                                     0x31,
		CHAR_CREATE_NAME_IN_USE:                                0x32,
		CHAR_CREATE_DISABLED:                                   0x33,
		CHAR_CREATE_PVP_TEAMS_VIOLATION:                        0x34,
		CHAR_CREATE_SERVER_LIMIT:                               0x35,
		CHAR_CREATE_ACCOUNT_LIMIT:                              0x36,
		CHAR_CREATE_SERVER_QUEUE:                               0x37,
		CHAR_CREATE_ONLY_EXISTING:                              0x38,
		CHAR_CREATE_EXPANSION:                                  0x39,
		CHAR_CREATE_EXPANSION_CLASS:                            0x3A,
		CHAR_CREATE_LEVEL_REQUIREMENT:                          0x3B,
		CHAR_CREATE_UNIQUE_CLASS_LIMIT:                         0x3C,
		CHAR_CREATE_CHARACTER_IN_GUILD:                         0x3D,
		CHAR_CREATE_RESTRICTED_RACECLASS:                       0x3E,
		CHAR_CREATE_CHARACTER_CHOOSE_RACE:                      0x3F,
		CHAR_CREATE_CHARACTER_ARENA_LEADER:                     0x40,
		CHAR_CREATE_CHARACTER_DELETE_MAIL:                      0x41,
		CHAR_CREATE_CHARACTER_SWAP_FACTION:                     0x42,
		CHAR_CREATE_CHARACTER_RACE_ONLY:                        0x43,
		CHAR_CREATE_CHARACTER_GOLD_LIMIT:                       0x44,
		CHAR_CREATE_FORCE_LOGIN:                                0x45,
		CHAR_NAME_SUCCESS:                                      0x57,
		CHAR_NAME_FAILURE:                                      0x58,
		CHAR_NAME_NO_NAME:                                      0x59,
		CHAR_NAME_TOO_SHORT:                                    0x5A,
		CHAR_NAME_TOO_LONG:                                     0x5B,
		CHAR_NAME_INVALID_CHARACTER:                            0x5C,
		CHAR_NAME_MIXED_LANGUAGES:                              0x5D,
		CHAR_NAME_PROFANE:                                      0x5E,
		CHAR_NAME_RESERVED:                                     0x5F,
		CHAR_NAME_INVALID_APOSTROPHE:                           0x60,
		CHAR_NAME_MULTIPLE_APOSTROPHES:                         0x61,
		CHAR_NAME_THREE_CONSECUTIVE:                            0x62,
		CHAR_NAME_INVALID_SPACE:                                0x63,
		CHAR_NAME_CONSECUTIVE_SPACES:                           0x64,
		CHAR_NAME_RUSSIAN_CONSECUTIVE_SILENT_CHARACTERS:        0x65,
		CHAR_NAME_RUSSIAN_SILENT_CHARACTER_AT_BEGINNING_OR_END: 0x66,
		CHAR_NAME_DECLENSION_DOESNT_MATCH_BASE_NAME:            0x67,

		CHAR_DELETE_IN_PROGRESS:                70,
		CHAR_DELETE_SUCCESS:                    71,
		CHAR_DELETE_FAILED:                     72,
		CHAR_DELETE_FAILED_LOCKED_FOR_TRANSFER: 73,
		CHAR_DELETE_FAILED_GUILD_LEADER:        74,
		CHAR_DELETE_FAILED_ARENA_CAPTAIN:       75,
	},
}

func DecodeCharacterOp(version vsn.Build, in *etc.Buffer) (CharacterOp, error) {
	desc, ok := CharacterOpDescriptors[version]
	if !ok {
		return 0, fmt.Errorf("packet: no CharacterOpDescriptor for version %d", version)
	}

	ib := in.ReadByte()

	for op, value := range desc {
		if value == ib {
			return op, nil
		}
	}

	return 0, fmt.Errorf("packet: no character op found for byte 0x%02X in version %d", ib, version)
}

func EncodeCharacterOp(version vsn.Build, out *etc.Buffer, value CharacterOp) error {
	desc, ok := CharacterOpDescriptors[version]
	if !ok {
		return fmt.Errorf("packet: no CharacterOpDescriptor for version %d", version)
	}

	opcode, ok := desc[value]
	if !ok {
		return fmt.Errorf("packet: no opcode for %s in the descriptor", value)
	}

	out.WriteByte(opcode)

	return nil
}

type CharLoginResult uint8

const (
	CharNoWorld CharLoginResult = iota
	CharLoginDuplicateCharacter
	CharLoginNoInstances
	CharLoginDisabled
	CharLoginNoCharacter
	CharLoginLockedForTransfer
	CharLoginLockedByBilling
	CharLoginFailed
)

var CharLoginResultDescriptors = map[vsn.Build]map[CharLoginResult]uint8{
	5875: {
		CharNoWorld:                 0x01,
		CharLoginDuplicateCharacter: 0x02,
		CharLoginNoInstances:        0x03,
		CharLoginDisabled:           0x04,
		CharLoginNoCharacter:        0x05,
		CharLoginLockedForTransfer:  0x06,
		CharLoginLockedByBilling:    0x07,
		CharLoginFailed:             0x08,
	},
}

type CharacterFlags uint32

const (
	CharacterLockedForTransfer CharacterFlags = 1 << iota
	CharacterHideHelm
	CharacterHideCloak
	CharacterGhost
	CharacterRename
	CharacterLockedByBilling
	CharacterDeclined
)

var CharacterFlagDescriptors = map[vsn.Build]map[CharacterFlags]uint32{
	5875: {
		CharacterLockedForTransfer: 0x00000004,
		CharacterHideHelm:          0x00000400,
		CharacterHideCloak:         0x00000800,
		CharacterGhost:             0x00002000,
		CharacterRename:            0x00004000,
		CharacterLockedByBilling:   0x01000000,
		CharacterDeclined:          0x02000000,
	},
}

func EncodeCharacterFlags(build vsn.Build, out io.Writer, flags CharacterFlags) error {
	desc, ok := CharacterFlagDescriptors[build]
	if !ok {
		return fmt.Errorf("packet: unknown build in character flags %s", build)
	}

	var uflags uint32

	for flag, code := range desc {
		if flags&flag != 0 {
			uflags |= code
		}
	}

	var bytes [4]byte
	binary.LittleEndian.PutUint32(bytes[:], uflags)
	_, err := out.Write(bytes[:])
	return err
}

func DecodeCharacterFlags(build vsn.Build, in io.Reader) (CharacterFlags, error) {
	desc, ok := CharacterFlagDescriptors[build]
	if !ok {
		return 0, fmt.Errorf("packet: unknown build in character flags %s", build)
	}

	var bytes [4]byte
	_, err := in.Read(bytes[:])
	if err != nil {
		return 0, err
	}

	var uflags uint32 = binary.LittleEndian.Uint32(bytes[:])
	var outflags CharacterFlags
	for flag, code := range desc {
		if uflags&code != 0 {
			outflags |= flag
		}
	}

	return outflags, nil
}
