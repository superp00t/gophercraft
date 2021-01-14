//package chat contains packets relating to the in-game chatbox feature.
package chat

import (
	"fmt"

	"github.com/superp00t/etc"
	"github.com/superp00t/gophercraft/guid"
	"github.com/superp00t/gophercraft/packet"
	"github.com/superp00t/gophercraft/vsn"
)

const (
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
	Type        MsgType
	Language    uint32
	ChannelName string
	PlayerRank  uint32
	Name        string
	SenderGUID  guid.GUID
	TargetGUID  guid.GUID
	TargetName  string
	Body        string
	Tag         uint8
}

func UnmarshalClientMessage(build vsn.Build, input []byte) (*Message, error) {
	in := etc.FromBytes(input)

	code := in.ReadUint32()

	mType, err := ResolveMsgType(build, uint32(code))
	if err != nil {
		return nil, err
	}

	msg := new(Message)
	msg.Type = mType
	msg.Language = in.ReadUint32()

	switch msg.Type {
	case MsgSay, MsgEmote, MsgYell,
		MsgParty, MsgOfficer, MsgRaid,
		MsgRaidLeader, MsgRaidWarning,
		MsgAFK, MsgDND:
		msg.Body = in.ReadCString()
	case MsgWhisper:
		msg.Name = in.ReadCString()
		msg.Body = in.ReadCString()
	case MsgChannel:
		msg.Name = in.ReadCString()
		msg.Body = in.ReadCString()
	default:
		return nil, fmt.Errorf("chat: unrecognized type in client message: %v", msg.Type)
	}

	return msg, nil
}

func UnmarshalMessage(build vsn.Build, input []byte) (*Message, error) {
	in := etc.FromBytes(input)

	code := in.ReadByte()

	var err error
	cm := new(Message)
	cm.Type, err = ResolveMsgType(build, uint32(code))
	if err != nil {
		return nil, err
	}

	cm.Language = in.ReadUint32()

	switch cm.Type {
	case MsgCreatureWhisper, MsgRaidBossWhisper, MsgRaidBossEmote, MsgCreatureEmote:
		cm.Name = DecodeChatString(build, in)
		cm.SenderGUID, err = guid.DecodeUnpacked(build, in)
		if err != nil {
			return nil, err
		}
	case MsgSay, MsgParty, MsgYell:
		cm.SenderGUID = guid.Classic(in.ReadUint64())
		in.ReadUint64()
	case MsgCreatureSay, MsgCreatureYell:
		cm.SenderGUID = guid.Classic(in.ReadUint64())
		cm.Name = DecodeChatString(build, in)
		cm.TargetGUID = guid.Classic(in.ReadUint64())
	case MsgChannel:
		cm.ChannelName = DecodeChatString(build, in)
		cm.PlayerRank = in.ReadUint32()
		cm.SenderGUID = guid.Classic(in.ReadUint64())
	default:
		cm.SenderGUID = guid.Classic(in.ReadUint64())
	}

	cm.Body = DecodeChatString(build, in)
	return cm, nil
}

func DecodeChatString(build vsn.Build, in *etc.Buffer) string {
	if build.RemovedIn(vsn.V1_12_1) {
		return in.ReadCString()
	}
	snd := in.ReadUint32()
	return in.ReadFixedString(int(snd))
}

func EncodeChatString(build vsn.Build, out *etc.Buffer, str string) {
	if build.RemovedIn(vsn.V1_12_1) {
		out.WriteCString(str)
		return
	}
	out.WriteUint32(uint32(len(str) + 1))
	out.Write([]byte(str))
	out.WriteByte(0)
}

func (cm *Message) buildPacketAfter2(build vsn.Build, p *packet.WorldPacket) {
	cm.SenderGUID.EncodeUnpacked(build, p)
	p.WriteUint32(0)

	switch cm.Type {
	case MsgCreatureSay, MsgCreatureParty, MsgCreatureYell, MsgCreatureWhisper, MsgCreatureEmote, MsgRaidBossWhisper, MsgRaidBossEmote, MsgWhisperForeign:
		EncodeChatString(build, p.Buffer, cm.Name)
		cm.TargetGUID.EncodeUnpacked(build, p)
		if cm.Type != MsgWhisperForeign {
			if (cm.TargetGUID != guid.Nil) && !(cm.TargetGUID.HighType() == guid.Player) && !(cm.TargetGUID.HighType() == guid.Pet) {
				EncodeChatString(build, p.Buffer, cm.TargetName)
			}
		}
		EncodeChatString(build, p.Buffer, cm.Body)
		p.WriteByte(cm.Tag)
	case MsgBGSystemNeutral, MsgBGSystemBlueTeam, MsgBGSystemRedTeam:
		cm.TargetGUID.EncodeUnpacked(build, p)
		if cm.TargetGUID != guid.Nil && cm.TargetGUID.HighType() != guid.Player {
			EncodeChatString(build, p.Buffer, cm.TargetName)
		}
		EncodeChatString(build, p.Buffer, cm.Body)
		p.WriteByte(cm.Tag)
	default:
		if cm.Type == MsgChannel {
			p.WriteCString(cm.ChannelName)
		}

		cm.TargetGUID.EncodeUnpacked(build, p)
		EncodeChatString(build, p.Buffer, cm.Body)
		p.WriteByte(cm.Tag)

		// if cm.Tag == TAG_GM {
		// 	EncodeChatString(build, p.Buffer, cm.Name)
		// }
	}
}

func (cm *Message) Packet(build vsn.Build) *packet.WorldPacket {
	p := packet.NewWorldPacket(packet.SMSG_MESSAGECHAT)

	msgCode, err := ConvertMsgType(build, cm.Type)
	if err != nil {
		panic(err)
	}

	p.WriteByte(uint8(msgCode))
	p.WriteUint32(cm.Language)

	// Alpha format is much more basic
	if build.RemovedIn(vsn.V1_12_1) {
		cm.SenderGUID.EncodeUnpacked(build, p)
		EncodeChatString(build, p.Buffer, cm.Body)
		p.WriteByte(cm.Tag)
		return p
	}

	if build.AddedIn(vsn.V2_4_3) {
		cm.buildPacketAfter2(build, p)
		return p
	}

	switch cm.Type {
	case MsgCreatureWhisper, MsgRaidBossWhisper, MsgRaidBossEmote, MsgCreatureEmote:
		EncodeChatString(build, p.Buffer, cm.Name)
		cm.SenderGUID.EncodeUnpacked(build, p)
	case MsgSay, MsgParty, MsgYell:
		cm.SenderGUID.EncodeUnpacked(build, p)
		cm.SenderGUID.EncodeUnpacked(build, p)
	case MsgCreatureSay, MsgCreatureYell:
		cm.SenderGUID.EncodeUnpacked(build, p)
		EncodeChatString(build, p.Buffer, cm.Name)
		cm.TargetGUID.EncodeUnpacked(build, p)
	case MsgChannel:
		EncodeChatString(build, p.Buffer, cm.ChannelName)
		p.WriteUint32(cm.PlayerRank)
		cm.SenderGUID.EncodeUnpacked(build, p)
	default:
		cm.SenderGUID.EncodeUnpacked(build, p)
	}

	EncodeChatString(build, p.Buffer, cm.Body)
	p.Buffer.WriteByte(cm.Tag)
	return p
}
