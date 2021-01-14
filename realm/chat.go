package realm

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/superp00t/gophercraft/guid"
	"github.com/superp00t/gophercraft/packet"
	"github.com/superp00t/gophercraft/vsn"

	"github.com/superp00t/gophercraft/gcore/sys"

	"github.com/superp00t/etc/yo"
	"github.com/superp00t/gophercraft/packet/chat"
)

func (s *Session) SendChat(ch *chat.Message) {
	p := ch.Packet(s.Build())
	s.SendAsync(p)
}

func (s *Session) SystemChat(data string) {
	lines := strings.Split(data, "\n")

	for _, ln := range lines {
		s.sysChatLine(ln)
	}
}

func (s *Session) sysChatLine(ln string) {
	s.SendChat(&chat.Message{
		Type:     chat.MsgSystem,
		Language: chat.LANG_UNIVERSAL,
		Body:     ln,
	})
}

func (s *Session) Sysf(data string, args ...interface{}) {
	s.SystemChat(fmt.Sprintf(data, args...))
}

// prints something in green text with [SERVER] prepended
// for use in global announcements
func (s *Session) Annf(data string, args ...interface{}) {
	s.SystemChat("|cFF00FF00[SERVER] " + fmt.Sprintf(data, args...) + "|r")
}

func (s *Session) MOTD(fstr string, args ...interface{}) {
	if s.Build().AddedIn(vsn.V2_4_3) {
		str := fmt.Sprintf(fstr, args...)
		mtd := packet.NewWorldPacket(packet.SMSG_MOTD)
		elements := strings.Split(str, "\n")
		mtd.WriteUint32(uint32(len(elements)))

		for _, el := range elements {
			mtd.WriteCString(el)
		}

		s.SendAsync(mtd)
		return
	}

	s.ColorPrintf("FF50c41a", fstr, args...)
}

func (s *Session) ColorPrintf(color string, data string, args ...interface{}) {
	printed := fmt.Sprintf(data, args...)
	if len(printed) < 255 {
		// We can send this as one packet.
		s.sysChatLine("|c" + color + printed + "|r")
		return
	}

	lines := strings.Split(printed, "\n")

	for _, ln := range lines {
		s.sysChatLine("|c" + color + ln + "|r")
	}
}

func (s *Session) Warnf(data string, args ...interface{}) {
	s.ColorPrintf("FFFFFF00", data, args...)
}

func (s *Session) printfObjMgr(data string, args ...interface{}) {
	s.Sysf("|cFFD97438[Object Manager]|r %s", fmt.Sprintf(data, args...))
}

func (s *Session) NoSuchPlayer(playerName string) {
	s.Warnf("The player '%s' could not be found.", playerName)
}

func (s *Session) PlayerName() string {
	return s.Char.Name
}

func (s *Session) Tag() uint8 {
	if s.Tier == sys.Tier_Admin {
		return chat.TAG_GM
	}

	return chat.TAG_NONE
}

func (s *Session) IsGM() bool {
	return s.Tier >= sys.Tier_GameMaster
}

func (s *Session) IsAdmin() bool {
	return s.Tier >= sys.Tier_Admin
}

func (s *Session) HandleChat(b []byte) {
	cm, err := chat.UnmarshalClientMessage(s.Build(), b)
	if err != nil {
		yo.Warn(err)
		return
	}

	if len(cm.Body) > 255 {
		return
	}

	if cm.Body == "" {
		return
	}

	if !utf8.ValidString(cm.Body) {
		return
	}

	if strings.HasPrefix(cm.Body, ".") && len(cm.Body) > 1 {
		s.HandleCommand(cm.Body)
		return
	}

	ss, _ := s.WS.ThinkOn(ChatEvent, cm.Type, cm)
	if ss {
		return
	}

	switch cm.Type {
	case chat.MsgWhisper:
		s.HandleWhisper(cm)
	case chat.MsgSay:
		s.HandleSay(cm)
	case chat.MsgParty:
		s.HandlePartyMessage(cm)
	}
}

func (s *Session) SendChatPlayerNotFound(name string) {
	p := packet.NewWorldPacket(packet.SMSG_CHAT_PLAYER_NOT_FOUND)
	p.WriteCString(name)
	s.SendAsync(p)
}

func (s *Session) SendChatPlayerIsEnemy(name string) {
	p := packet.NewWorldPacket(packet.SMSG_CHAT_WRONG_FACTION)
	p.WriteCString(name)
	s.SendAsync(p)
}

func (s *Session) SendChatPlayerIsIgnoringYou(name string) {
	p := packet.NewWorldPacket(packet.SMSG_CHAT_IGNORED_ACCOUNT_MUTED)
	p.WriteCString(name)
	s.SendAsync(p)
}

func (s *Session) IsEnemy(wo WorldObject) bool {
	return false
}

func (s *Session) IsIgnoring(player guid.GUID) bool {
	return false
}

func (s *Session) HandleWhisper(whisper *chat.Message) {
	target := whisper.Name
	targetSession, err := s.WS.GetSessionByPlayerName(target)
	if err != nil {
		s.SendChatPlayerNotFound(target)
		return
	}

	if s.Config().Bool("Chat.LanguageBarrier") {
		if s.IsEnemy(targetSession) {
			s.SendChatPlayerIsEnemy(target)
			return
		}
	}

	if targetSession.IsIgnoring(s.GUID()) {
		s.SendChatPlayerIsIgnoringYou(target)
		return
	}

	s.SendChat(&chat.Message{
		Type:       chat.MsgWhisperInform,
		SenderGUID: s.GUID(),
		TargetGUID: targetSession.GUID(),
		Body:       whisper.Body,
	})

	targetSession.SendChat(&chat.Message{
		Type:       chat.MsgWhisper,
		SenderGUID: s.GUID(),
		Body:       whisper.Body,
	})
}

func (s *Session) HandleSay(say *chat.Message) {
	var lang uint32 = chat.LANG_UNIVERSAL
	// TODO: check appropriate speech babblification in future
	// if s.Config().Bool("PVP.AtWar") {

	// }

	// TODO: use the existing structure

	pck := chat.Message{
		Type:       chat.MsgSay,
		Language:   lang,
		Name:       s.PlayerName(),
		SenderGUID: s.GUID(),
		Body:       say.Body,
		Tag:        s.Tag(),
	}

	s.SendAreaAll(pck.Packet(s.Build()))
}
