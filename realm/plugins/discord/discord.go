package discord

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/gorilla/websocket"
	"github.com/superp00t/etc/yo"
	"github.com/superp00t/gophercraft/realm"
	"github.com/superp00t/gophercraft/realm/wdb"
)

const blurple = "7289DAFF"

var DiscordEnabled = wdb.MakePropID("DiscrdOn")

var (
	ErrAppSecretNotSet = fmt.Errorf("discord: Discord.AppSecret not set in world config")
	ErrChannelNotSet   = fmt.Errorf("discord: Discord.ChannelID not set in world config")
)

type DiscordPlugin struct {
	Server        *realm.Server
	Session       *discordgo.Session
	ChannelID     string
	NonFatalError error
}

func (p *DiscordPlugin) Activated() (bool, error) {
	return p.Session != nil, p.NonFatalError
}

func (p *DiscordPlugin) Terminate() error {
	return p.Session.Close()
}

func cmdToggle(s *realm.Session) {
	if s.HasProp(DiscordEnabled) {
		s.Warnf("Discord disabled.")
		s.RemoveProp(DiscordEnabled)
	} else {
		s.Warnf("Discord enabled.")
		s.AddProp(DiscordEnabled)
	}
}

func (p *DiscordPlugin) GetChannelName() string {
	for _, g := range p.Session.State.Guilds {
		for _, c := range g.Channels {
			if c.ID == p.ChannelID {
				return c.Name
			}
		}
	}

	panic("unknown name")
}

func (p *DiscordPlugin) Init(server *realm.Server, info *realm.PluginInfo) error {
	p.Server = server

	p.Server.Cmd(realm.Player, "discord toggle", "Enable or disable Discord on your account", cmdToggle)

	info.Name = "Gophercraft Official Discord Plugin"
	info.Version = "0.1"
	info.Authors = []string{"The Gophercraft Developers"}

	secret := server.Config.GetString("Discord.AppSecret")

	// User may not have Discord configured.
	if secret == "" {
		p.NonFatalError = ErrAppSecretNotSet
		yo.Warn(p.NonFatalError)
		return nil
	}

	p.ChannelID = server.Config.GetString("Discord.ChannelID")

	if p.ChannelID == "" {
		return ErrChannelNotSet
	}

	// p.Server.On(realm.ChatEvent, nil, func(session *realm.Session, msg *chat.Message) {

	// })

	// User network may be temporarily unreachable.
	var err error
	p.Session, err = discordgo.New("Bot " + secret)
	if err != nil {
		p.NonFatalError = err
		return nil
	}

	p.Session.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		if m.Author.ID == p.Session.State.User.ID {
			return
		}

		if m.ChannelID != p.ChannelID {
			return
		}

		msgData := fmt.Sprintf("[|c%s%s|r] %s: %s", blurple, p.GetChannelName(), m.Message.Author.Username, m.Message.Content)

		p.Server.AllSessions().Iter(func(s *realm.Session) {
			if s.HasProp(DiscordEnabled) {
				s.SystemChat(msgData)
			}
		})
	})

	p.Server.Cmd(realm.Player, "announce", "Send a message to the server's discord channel", func(s *realm.Session, msg string) {
		p.Session.ChannelMessageSendComplex(p.ChannelID, &discordgo.MessageSend{
			Content: fmt.Sprintf("**%s**: %s", s.PlayerName(), msg),
		})
	})

	if err = p.Session.Open(); err != nil {
		switch err.(type) {
		case *websocket.CloseError:
			return fmt.Errorf("discord: %s", err)
		}
		p.NonFatalError = err
		return nil
	}

	return nil
}

func init() {
	realm.RegisterPlugin("discord", &DiscordPlugin{})
}
