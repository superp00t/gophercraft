package chat

import (
	"github.com/superp00t/etc"
	"github.com/superp00t/gophercraft/guid"
	"github.com/superp00t/gophercraft/packet"
)

const (
	MSG_ADDON               = 0xFFFFFFFF
	MSG_SAY                 = 0x00
	MSG_PARTY               = 0x01
	MSG_RAID                = 0x02
	MSG_GUILD               = 0x03
	MSG_OFFICER             = 0x04
	MSG_YELL                = 0x05
	MSG_WHISPER             = 0x06
	MSG_WHISPER_INFORM      = 0x07
	MSG_EMOTE               = 0x08
	MSG_TEXT_EMOTE          = 0x09
	MSG_SYSTEM              = 0x0A
	MSG_MONSTER_SAY         = 0x0B
	MSG_MONSTER_YELL        = 0x0C
	MSG_MONSTER_EMOTE       = 0x0D
	MSG_CHANNEL             = 0x0E
	MSG_CHANNEL_JOIN        = 0x0F
	MSG_CHANNEL_LEAVE       = 0x10
	MSG_CHANNEL_LIST        = 0x11
	MSG_CHANNEL_NOTICE      = 0x12
	MSG_CHANNEL_NOTICE_USER = 0x13
	MSG_AFK                 = 0x14
	MSG_DND                 = 0x15
	MSG_IGNORED             = 0x16
	MSG_SKILL               = 0x17
	MSG_LOOT                = 0x18
	MSG_MONSTER_WHISPER     = 0x1A
	MSG_BG_SYSTEM_NEUTRAL   = 0x52
	MSG_BG_SYSTEM_ALLIANCE  = 0x53
	MSG_BG_SYSTEM_HORDE     = 0x54
	MSG_RAID_LEADER         = 0x57
	MSG_RAID_WARNING        = 0x58
	MSG_RAID_BOSS_WHISPER   = 0x59
	MSG_RAID_BOSS_EMOTE     = 0x5A
	MSG_BATTLEGROUND        = 0x5C
	MSG_BATTLEGROUND_LEADER = 0x5D
	MSG_MAX                 = 0x5E

	LANG_UNIVERSAL   = 0
	LANG_ORCISH      = 1
	LANG_DARNASSIAN  = 2
	LANG_TAURAHE     = 3
	LANG_DWARVISH    = 6
	LANG_COMMON      = 7
	LANG_DEMONIC     = 8
	LANG_TITAN       = 9
	LANG_THALASSIAN  = 10
	LANG_DRACONIC    = 11
	LANG_KALIMAG     = 12
	LANG_GNOMISH     = 13
	LANG_TROLL       = 14
	LANG_GUTTERSPEAK = 33
	LANG_ADDON       = 0xFFFFFFFF

	TAG_NONE = 0
	TAG_AFK  = 1
	TAG_DND  = 2
	TAG_GM   = 3
)

type Message struct {
	Type        uint8
	Language    uint32
	ChannelName string
	PlayerRank  uint32
	SenderName  string
	SenderGUID  guid.GUID
	TargetGUID  guid.GUID
	Body        string
	Tag         uint8
}

func UnmarshalMessage(input []byte) *Message {
	in := etc.FromBytes(input)

	cm := new(Message)
	cm.Type = in.ReadByte()
	cm.Language = in.ReadUint32()

	switch cm.Type {
	case MSG_MONSTER_WHISPER, MSG_RAID_BOSS_WHISPER, MSG_RAID_BOSS_EMOTE, MSG_MONSTER_EMOTE:
		cm.SenderName = DecodeUintString(in)
		cm.SenderGUID = guid.Classic(in.ReadUint64())
	case MSG_SAY, MSG_PARTY, MSG_YELL:
		cm.SenderGUID = guid.Classic(in.ReadUint64())
		in.ReadUint64()
	case MSG_MONSTER_SAY, MSG_MONSTER_YELL:
		cm.SenderGUID = guid.Classic(in.ReadUint64())
		cm.SenderName = DecodeUintString(in)
		cm.TargetGUID = guid.Classic(in.ReadUint64())
	case MSG_CHANNEL:
		cm.ChannelName = DecodeUintString(in)
		cm.PlayerRank = in.ReadUint32()
		cm.SenderGUID = guid.Classic(in.ReadUint64())
	default:
		cm.SenderGUID = guid.Classic(in.ReadUint64())
	}

	cm.Body = DecodeUintString(in)
	return cm
}

func DecodeUintString(in *etc.Buffer) string {
	snd := in.ReadUint32()
	return in.ReadFixedString(int(snd))
}

func EncodeUintString(out *etc.Buffer, str string) {
	out.WriteUint32(uint32(len(str) + 1))
	out.Write([]byte(str))
	out.WriteByte(0)
}

func (cm *Message) Packet() *packet.WorldPacket {
	p := packet.NewWorldPacket(packet.SMSG_MESSAGECHAT)
	p.WriteByte(cm.Type)
	p.WriteUint32(cm.Language)

	switch cm.Type {
	case MSG_MONSTER_WHISPER, MSG_RAID_BOSS_WHISPER, MSG_RAID_BOSS_EMOTE, MSG_MONSTER_EMOTE:
		EncodeUintString(p.Buffer, cm.SenderName)
		p.WriteUint64(cm.SenderGUID.Classic())
	case MSG_SAY, MSG_PARTY, MSG_YELL:
		p.WriteUint64(cm.SenderGUID.Classic())
		p.WriteUint64(cm.SenderGUID.Classic())
	case MSG_MONSTER_SAY, MSG_MONSTER_YELL:
		p.WriteUint64(cm.SenderGUID.Classic())
		EncodeUintString(p.Buffer, cm.SenderName)
		p.WriteUint64(cm.TargetGUID.Classic())
	case MSG_CHANNEL:
		EncodeUintString(p.Buffer, cm.ChannelName)
		p.WriteUint32(cm.PlayerRank)
		p.WriteUint64(cm.SenderGUID.Classic())
	default:
		p.WriteUint64(cm.SenderGUID.Classic())
	}

	EncodeUintString(p.Buffer, cm.Body)
	p.Buffer.WriteByte(cm.Tag)
	return p
}
