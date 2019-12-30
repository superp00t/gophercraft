package worldserver

import (
	"sort"

	"github.com/superp00t/etc/yo"
	"github.com/superp00t/gophercraft/packet"
)

func (s *Session) SendSocialList() {
	p := packet.NewWorldPacket(packet.SMSG_CONTACT_LIST)
	Buf_sociallist := []byte{0x7, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}
	p.Write(Buf_sociallist)
	s.SendAsync(p)
}

func (s *Session) SendDanceMoves() {
	p := packet.NewWorldPacket(packet.SMSG_LEARNED_DANCE_MOVES)
	Buf_dance := []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}
	p.Write(Buf_dance)
	s.SendAsync(p)
}

func (s *Session) HandleWho(b []byte) {
	_, err := packet.UnmarshalWhoRequest(b)
	if err != nil {
		yo.Warn(err)
		return
	}

	w := &packet.Who{}
	var usernames []string

	s.WS.PlayersL.Lock()
	for k := range s.WS.PlayerList {
		usernames = append(usernames, k)
	}

	sort.Strings(usernames)
	whoMatches := make([]packet.WhoMatch, len(usernames))

	for _i, user := range usernames {
		playerPtr := s.WS.PlayerList[user]

		whoMatches[_i] = packet.WhoMatch{
			PlayerName: user,
			GuildName:  "",
			Level:      uint32(playerPtr.GetPlayerLevel()),
			Class:      uint32(playerPtr.GetPlayerClass()),
			Race:       uint32(playerPtr.GetPlayerRace()),
		}
	}

	w.WhoMatches = whoMatches
	s.WS.PlayersL.Unlock()

	s.SendAsync(w.Packet())
}
