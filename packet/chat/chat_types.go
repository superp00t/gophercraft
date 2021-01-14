package chat

import (
	"fmt"

	"github.com/superp00t/gophercraft/vsn"
)

type MsgType uint8

const (
	MsgAddon = iota
	MsgSay
	MsgParty
	MsgRaid
	MsgGuild
	MsgOfficer
	MsgYell
	MsgWhisper
	MsgWhisperForeign
	MsgWhisperInform
	MsgEmote
	MsgTextEmote
	MsgSystem
	MsgCreatureSay
	MsgCreatureYell
	MsgCreatureWhisper
	MsgCreatureEmote
	MsgChannel
	MsgChannelJoin
	MsgChannelLeave
	MsgChannelList
	MsgChannelNotice
	MsgChannelNoticeUser
	MsgAFK
	MsgDND
	MsgIgnored
	MsgSkill
	MsgLoot
	MsgMoney
	MsgOpening
	MsgBGSystemNeutral
	MsgBGSystemBlueTeam
	MsgBGSystemRedTeam
	MsgRaidLeader
	MsgRaidWarning
	MsgRaidBossWhisper
	MsgRaidBossEmote
	MsgBattleground
	MsgBattlegroundLeader
	MsgCreatureParty
	MsgTradeSkills
	MsgPetInfo
	MsgCombatMiscInfo
	MsgCombatXPGain
	MsgCombatHonorGain
	MsgCombatFactionGain
	MsgFiltered
	MsgRestricted
	MsgBNet
	MsgAchievement
	MsgGuildAchievement
	MsgArenaPoints
	MsgPartyLeader
)

var MsgTypeDescriptors = map[vsn.Build]map[MsgType]uint32{
	vsn.Alpha: {
		MsgSay:               0x00,
		MsgParty:             0x01,
		MsgGuild:             0x02,
		MsgOfficer:           0x03,
		MsgYell:              0x04,
		MsgWhisper:           0x05,
		MsgWhisperInform:     0x06,
		MsgEmote:             0x07,
		MsgTextEmote:         0x08,
		MsgSystem:            0x09,
		MsgCreatureSay:       0x0A,
		MsgCreatureYell:      0x0B,
		MsgCreatureEmote:     0x0C,
		MsgChannel:           0x0D,
		MsgChannelJoin:       0x0E,
		MsgChannelLeave:      0xF,
		MsgChannelList:       0x10,
		MsgChannelNotice:     0x11,
		MsgChannelNoticeUser: 0x12,
		MsgAFK:               0x13,
		MsgDND:               0x14,
		MsgIgnored:           0x16,
		MsgSkill:             0x17,
		MsgLoot:              0x18,
	},

	vsn.V1_12_1: {
		MsgAddon:              0xFFFFFFFF,
		MsgSay:                0x00,
		MsgParty:              0x01,
		MsgRaid:               0x02,
		MsgGuild:              0x03,
		MsgOfficer:            0x04,
		MsgYell:               0x05,
		MsgWhisper:            0x06,
		MsgWhisperInform:      0x07,
		MsgEmote:              0x08,
		MsgTextEmote:          0x09,
		MsgSystem:             0x0A,
		MsgCreatureSay:        0x0B,
		MsgCreatureYell:       0x0C,
		MsgCreatureEmote:      0x0D,
		MsgChannel:            0x0E,
		MsgChannelJoin:        0x0F,
		MsgChannelLeave:       0x10,
		MsgChannelList:        0x11,
		MsgChannelNotice:      0x12,
		MsgChannelNoticeUser:  0x13,
		MsgAFK:                0x14,
		MsgDND:                0x15,
		MsgIgnored:            0x16,
		MsgSkill:              0x17,
		MsgLoot:               0x18,
		MsgCreatureWhisper:    0x1A,
		MsgBGSystemNeutral:    0x52,
		MsgBGSystemBlueTeam:   0x53,
		MsgBGSystemRedTeam:    0x54,
		MsgRaidLeader:         0x57,
		MsgRaidWarning:        0x58,
		MsgRaidBossWhisper:    0x59,
		MsgRaidBossEmote:      0x5A,
		MsgBattleground:       0x5C,
		MsgBattlegroundLeader: 0x5D,
	},

	vsn.V2_4_3: {
		MsgAddon:              0xFFFFFFFF,
		MsgSystem:             0x00,
		MsgSay:                0x01,
		MsgParty:              0x02,
		MsgRaid:               0x03,
		MsgGuild:              0x04,
		MsgOfficer:            0x05,
		MsgYell:               0x06,
		MsgWhisper:            0x07,
		MsgWhisperForeign:     0x08,
		MsgWhisperInform:      0x09,
		MsgEmote:              0x0A,
		MsgTextEmote:          0x0B,
		MsgCreatureSay:        0x0C,
		MsgCreatureParty:      0x0D,
		MsgCreatureYell:       0x0E,
		MsgCreatureWhisper:    0x0F,
		MsgCreatureEmote:      0x10,
		MsgChannel:            0x11,
		MsgChannelJoin:        0x12,
		MsgChannelLeave:       0x13,
		MsgChannelList:        0x14,
		MsgChannelNotice:      0x15,
		MsgChannelNoticeUser:  0x16,
		MsgAFK:                0x17,
		MsgDND:                0x18,
		MsgIgnored:            0x19,
		MsgSkill:              0x1A,
		MsgLoot:               0x1B,
		MsgMoney:              0x1C,
		MsgOpening:            0x1D,
		MsgTradeSkills:        0x1E,
		MsgPetInfo:            0x1F,
		MsgCombatMiscInfo:     0x20,
		MsgCombatXPGain:       0x21,
		MsgCombatHonorGain:    0x22,
		MsgCombatFactionGain:  0x23,
		MsgBGSystemNeutral:    0x24,
		MsgBGSystemBlueTeam:   0x25,
		MsgBGSystemRedTeam:    0x26,
		MsgRaidLeader:         0x27,
		MsgRaidWarning:        0x28,
		MsgRaidBossEmote:      0x29,
		MsgRaidBossWhisper:    0x2A,
		MsgFiltered:           0x2B,
		MsgBattleground:       0x2C,
		MsgBattlegroundLeader: 0x2D,
		MsgRestricted:         0x2E,
	},

	vsn.V3_3_5a: {
		MsgAddon:              0xFFFFFFFF,
		MsgSystem:             0x00,
		MsgSay:                0x01,
		MsgParty:              0x02,
		MsgRaid:               0x03,
		MsgGuild:              0x04,
		MsgOfficer:            0x05,
		MsgYell:               0x06,
		MsgWhisper:            0x07,
		MsgWhisperForeign:     0x08,
		MsgWhisperInform:      0x09,
		MsgEmote:              0x0A,
		MsgTextEmote:          0x0B,
		MsgCreatureSay:        0x0C,
		MsgCreatureParty:      0x0D,
		MsgCreatureYell:       0x0E,
		MsgCreatureWhisper:    0x0F,
		MsgCreatureEmote:      0x10,
		MsgChannel:            0x11,
		MsgChannelJoin:        0x12,
		MsgChannelLeave:       0x13,
		MsgChannelList:        0x14,
		MsgChannelNotice:      0x15,
		MsgChannelNoticeUser:  0x16,
		MsgAFK:                0x17,
		MsgDND:                0x18,
		MsgIgnored:            0x19,
		MsgSkill:              0x1A,
		MsgLoot:               0x1B,
		MsgMoney:              0x1C,
		MsgOpening:            0x1D,
		MsgTradeSkills:        0x1E,
		MsgPetInfo:            0x1F,
		MsgCombatMiscInfo:     0x20,
		MsgCombatXPGain:       0x21,
		MsgCombatHonorGain:    0x22,
		MsgCombatFactionGain:  0x23,
		MsgBGSystemNeutral:    0x24,
		MsgBGSystemBlueTeam:   0x25,
		MsgBGSystemRedTeam:    0x26,
		MsgRaidLeader:         0x27,
		MsgRaidWarning:        0x28,
		MsgRaidBossEmote:      0x29,
		MsgRaidBossWhisper:    0x2A,
		MsgFiltered:           0x2B,
		MsgBattleground:       0x2C,
		MsgBattlegroundLeader: 0x2D,
		MsgRestricted:         0x2E,
		MsgBNet:               0x2F,
		MsgAchievement:        0x30,
		MsgGuildAchievement:   0x31,
		MsgArenaPoints:        0x32,
		MsgPartyLeader:        0x33,
	},
}

func ConvertMsgType(build vsn.Build, value MsgType) (uint32, error) {
	desc, ok := MsgTypeDescriptors[build]
	if !ok {
		return 0, fmt.Errorf("chat: no MsgType descriptor for %s", build)
	}

	code, ok := desc[value]
	if !ok {
		return 0, fmt.Errorf("chat: no MsgType value found for %v", value)
	}

	return code, nil
}

func ResolveMsgType(build vsn.Build, code uint32) (MsgType, error) {
	desc, ok := MsgTypeDescriptors[build]
	if !ok {
		return 0, fmt.Errorf("chat: no MsgType descriptor for %s", build)
	}

	for k, v := range desc {
		if v == code {
			return k, nil
		}
	}

	return 0, fmt.Errorf("chat: no MsgType found for 0x%04X in descriptor %s", code, build)
}
