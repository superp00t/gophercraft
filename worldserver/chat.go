package worldserver

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/davecgh/go-spew/spew"
	"github.com/superp00t/gophercraft/gcore/sys"

	"github.com/superp00t/etc"
	"github.com/superp00t/etc/yo"
	"github.com/superp00t/gophercraft/packet/chat"
)

func (s *Session) SendChat(ch *chat.Message) {
	p := ch.Packet()
	s.SendAsync(p)
}

func (s *Session) SystemChat(data string) {
	lines := strings.Split(data, "\n")

	for _, ln := range lines {
		s.SendChat(&chat.Message{
			Type:     chat.MSG_SYSTEM,
			Language: chat.LANG_UNIVERSAL,
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
	return s.Tier > sys.Tier_GameMaster
}

func (s *Session) HandleChat(b []byte) {
	e := etc.FromBytes(b)
	t := e.ReadUint32()
	// TODO: implement language checks
	lang := e.ReadUint32()

	if t >= chat.MSG_MAX {
		return
	}

	switch t {
	// TODO: implement rudimentary rate limiting
	case chat.MSG_SAY, chat.MSG_YELL, chat.MSG_EMOTE:
		lang = chat.LANG_UNIVERSAL
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

		pck := chat.Message{
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

func (s *Session) Invoke(handlers []Command, index int, args []string) {
	fmt.Println(index, len(args))
	if index >= len(args) {
		return
	}

	cmd := strings.ToLower(args[index])

	var foundHandler *Command

	for idx := range handlers {
		v := &handlers[idx]

		if v.Signature == cmd {
			foundHandler = v
			fmt.Println("Found exact match for", v.Signature, "for", cmd, spew.Sdump(foundHandler))
		}
	}

	if foundHandler == nil {
		fmt.Println("Could not find exact instance of", args[index], "searching for prefix")
		for idx := range handlers {
			v := &handlers[idx]

			if strings.HasPrefix(v.Signature, cmd) {
				fmt.Println("found", v.Signature, "as candidate for", args[index])
				foundHandler = v
			}
		}
	}

	if foundHandler == nil {
		s.Warnf("Unknown command: %s", args[index])
		return
	}

	fmt.Println("definitely found", foundHandler.Signature, spew.Sdump(foundHandler))

	switch c := foundHandler.Function.(type) {
	case []Command:
		fmt.Println("invoking subcommands of", args[index])
		s.Invoke(c, index+1, args)
	case func(*C):
		var ags []string
		if index+1 <= len(args) {
			ags = args[index+1:]
		}

		c(&C{
			s,
			ags,
		})
	}
}

func (s *Session) HandleCommand(c string) {
	yo.Ok("command received", c)
	args, err := parseCmd(c)
	if err != nil {
		yo.Warn(err)
		return
	}

	s.Invoke(CmdHandlers, 0, args)
}

func parseCmd(s string) ([]string, error) {
	e := etc.FromString(s)

	if rn, _, _ := e.ReadRune(); rn != '.' {
		return nil, fmt.Errorf("not a command")
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

			if argBuf.Len() == 0 && rn == ' ' {
				continue
			}

			if rn == ' ' {
				args = append(args, argBuf.ToString())
				argBuf = etc.NewBuffer()
				continue argScan
			}

			// Don't split markup block
			if rn == '|' {
				markupCode, _, _ := e.ReadRune()
				if markupCode == 'c' {
					e.Jump(-2)

					var markupText string
					for {
						r, _, _ := e.ReadRune()
						if r == 0 {
							argBuf.Write([]byte(markupText))
							break
						}

						if r == '|' {
							r2, _, _ := e.ReadRune()
							if r2 == 0 {
								argBuf.WriteRune(r)
								break
							}

							if r2 == 'r' {
								argBuf.WriteRune(r2)
								break
							}

							argBuf.WriteRune(r)
							argBuf.WriteRune(r2)
						} else {
							argBuf.WriteRune(r)
						}
					}
				} else {
					argBuf.WriteRune(rn)
					argBuf.WriteRune(markupCode)
				}
			} else {
				argBuf.WriteRune(rn)
			}
		}
	}
endScan:

	return args, nil
}
