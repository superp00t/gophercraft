package worldserver

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/superp00t/etc"
	"github.com/superp00t/etc/yo"
	"github.com/superp00t/gophercraft/packet"
)

func (s *Session) SendChat(ch *packet.ChatMessage) {
	p := ch.Packet()
	s.SendAsync(p)
}

func (s *Session) SystemChat(data string) {
	lines := strings.Split(data, "\n")

	for _, ln := range lines {
		s.SendChat(&packet.ChatMessage{
			Type:     packet.CHAT_MSG_SYSTEM,
			Language: packet.LANG_UNIVERSAL,
			Body:     ln,
		})
	}
}

func (s *Session) Sysf(data string, args ...interface{}) {
	s.SystemChat(fmt.Sprintf(data, args...))
}

// prints something in green text with [SERVER] prepended
// for use in global announcements
func (s *Session) Annf(data string, args ...interface{}) {
	s.SystemChat("|cFF00FF00[SERVER] " + fmt.Sprintf(data, args...) + "|r")
}

func (s *Session) Warnf(data string, args ...interface{}) {
	s.SystemChat(fmt.Sprintf("|cFFFFFF00%s|r", fmt.Sprintf(data, args...)))
}

func (s *Session) PlayerName() string {
	return s.Char.Name
}

func (s *Session) Tag() uint8 {
	if s.Tier == Admin {
		return packet.CHAT_TAG_GM
	}

	return packet.CHAT_TAG_NONE
}

func (s *Session) IsGM() bool {
	return s.Tier > GameMaster
}

func (s *Session) HandleChat(b []byte) {
	e := etc.FromBytes(b)
	t := e.ReadUint32()
	// TODO: implement language checks
	lang := e.ReadUint32()

	if t >= packet.CHAT_MSG_MAX {
		return
	}

	switch t {
	// TODO: implement rudimentary rate limiting
	case packet.CHAT_MSG_SAY, packet.CHAT_MSG_YELL, packet.CHAT_MSG_EMOTE:
		lang = packet.LANG_UNIVERSAL
		body := e.ReadCString()

		if len(body) > 255 {
			return
		}

		if body == "" {
			return
		}

		if !utf8.ValidString(body) {
			return
		}

		if strings.HasPrefix(body, ".") {
			s.HandleCommand(body)
			return
		}

		pck := packet.ChatMessage{
			Type:       uint8(t),
			Language:   lang,
			SenderName: s.PlayerName(),
			SenderGUID: s.GUID(),
			Body:       body,
			Tag:        s.Tag(),
		}

		s.SendAreaAll(pck.Packet())
	}
}

func (s *Session) HandleCommand(c string) {
	yo.Ok("command received", c)

	cmd, args, err := parseCmd(c)
	if err != nil {
		yo.Warn(err)
		return
	}

	for _, v := range CmdHandlers {
		if v.Signature == cmd {
			if v.Required <= s.Tier {
				inv := &C{
					s,
					cmd,
					args,
				}

				v.Function(inv)
				return
			} else {
				s.Annf("Sorry, you lack the required permissions to invoke this command. Contact an admin if you believe this is in error.")
			}
			break
		}
	}

	s.Annf("Unknown command: %s", cmd)
}

func parseCmd(s string) (string, []string, error) {
	e := etc.FromString(s)

	if rn, _, _ := e.ReadRune(); rn != '.' {
		return "", nil, fmt.Errorf("not a command")
	}

	name := etc.NewBuffer()

	for {
		if e.Available() == 0 {
			break
		}

		rn, _, err := e.ReadRune()
		if err != nil {
			return "", nil, err
		}

		if rn == 0 {
			break
		}

		if rn == ' ' {
			break
		}

		name.WriteRune(rn)
	}

	if e.Available() == 0 {
		return name.ToString(), nil, nil
	}

	var args []string

argScan:
	for {
		argBuf := etc.NewBuffer()

		for x := 0; ; x++ {
			rn, _, _ := e.ReadRune()
			if rn == 0 {
				args = append(args, argBuf.ToString())
				goto endScan
			}

			if rn == ' ' && x == 0 {
				continue
			}

			if rn == ' ' {
				args = append(args, argBuf.ToString())
				argBuf = etc.NewBuffer()
				continue argScan
			}

			argBuf.WriteRune(rn)
		}
	}
endScan:

	return name.ToString(), args, nil
}
