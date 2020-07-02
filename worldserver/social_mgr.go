package worldserver

import (
	"sort"

	"github.com/superp00t/etc"
	"github.com/superp00t/etc/yo"

	"github.com/superp00t/gophercraft/guid"
	"github.com/superp00t/gophercraft/packet"
	"github.com/superp00t/gophercraft/vsn"
	"github.com/superp00t/gophercraft/worldserver/wdb"
)

func (ws *WorldServer) GetFriendStatus(id guid.GUID) packet.FriendStatus {
	_, err := ws.GetSessionByGUID(id)
	if err != nil {
		return packet.FriendOffline
	}

	return packet.FriendOnline
}

func (s *Session) SendSocialList() {
	p := packet.NewWorldPacket(packet.SMSG_CONTACT_LIST)

	var social []wdb.Contact

	if s.Build().AddedIn(vsn.V2_4_3) {
		s.DB().Where("player = ?", s.GUID().Counter()).Find(&social)

		p.WriteUint32(0x7)
		p.WriteUint32(uint32(len(social)))
	} else {
		if err := s.DB().Where("player = ?", s.GUID().Counter()).Where("friended = 1").Find(&social); err != nil {
			panic(err)
		}
		yo.Spew(social)
		p.WriteByte(uint8(len(social)))
	}

	for _, contact := range social {
		id := guid.RealmSpecific(guid.Player, s.WS.RealmID(), contact.Friend)
		id.EncodeUnpacked(s.Build(), p)

		if s.Build().AddedIn(vsn.V2_4_3) {
			flags := uint32(0)
			if contact.Friended {
				flags |= packet.SocialFlagFriend
			}
			if contact.Muted {
				flags |= packet.SocialFlagMuted
			}
			if contact.Ignored {
				flags |= packet.SocialFlagIgnored
			}
			p.WriteUint32(flags)
			p.WriteCString(contact.Note)
		}

		if contact.Friended {
			// Is player online?
			var status packet.FriendStatus
			fsession, err := s.WS.GetSessionByGUID(id)
			if err != nil {
				status = packet.FriendOffline
			} else {
				status = packet.FriendOnline
			}
			p.WriteByte(uint8(status))
			if status > 0 {
				// If online, show where player is.
				p.WriteUint32(fsession.ZoneID)
				p.WriteUint32(uint32(fsession.GetLevel()))
				p.WriteUint32(uint32(fsession.GetPlayerClass()))
			}
		}
	}

	s.SendAsync(p)

	// Ignore list used to be a separate packet.
	if s.Build().RemovedIn(vsn.V2_4_3) {
		// CMSG_SET_CONTACT_NOTES was SMSG_IGNORE_LIST
		social = nil
		p := packet.NewWorldPacket(packet.CMSG_SET_CONTACT_NOTES)
		if err := s.DB().Where("player = ?", s.GUID().Counter()).Where("ignored = 1").Find(&social); err != nil {
			panic(err)
		}
		p.WriteByte(uint8(len(social)))
		for _, contact := range social {
			id := guid.RealmSpecific(guid.Player, s.WS.RealmID(), contact.Friend)
			id.EncodeUnpacked(s.Build(), p)
		}
		s.SendAsync(p)
	}
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
			Level:      uint32(playerPtr.GetLevel()),
			Class:      uint32(playerPtr.GetPlayerClass()),
			Race:       uint32(playerPtr.GetPlayerRace()),
			ZoneID:     playerPtr.ZoneID,
		}
	}

	w.WhoMatches = whoMatches
	s.WS.PlayersL.Unlock()

	s.SendAsync(w.Packet())
}

func (s *Session) SendFriendStatus(result uint8, id guid.GUID, note string, status packet.FriendStatus, area, level, class uint32) {
	data := packet.NewWorldPacket(packet.SMSG_FRIEND_STATUS)
	data.WriteByte(result)
	id.EncodeUnpacked(s.Build(), data)

	switch result {
	case packet.FRIEND_ADDED_OFFLINE, packet.FRIEND_ADDED_ONLINE:
		if s.Build().AddedIn(vsn.V2_4_3) {
			data.WriteCString(note)
		}
	default:
	}

	switch result {
	case packet.FRIEND_ADDED_ONLINE, packet.FRIEND_ONLINE:
		data.WriteByte(uint8(status))
		data.WriteUint32(area)
		data.WriteUint32(level)
		data.WriteUint32(class)
	default:
	}

	s.SendAsync(data)
}

func (s *Session) GetContact(friend guid.GUID) *wdb.Contact {
	var contact wdb.Contact
	found, err := s.DB().Where("player = ?", s.GUID().Counter()).Where("friend = ?", friend.Counter()).Get(&contact)
	if err != nil {
		panic(err)
	}
	if !found {
		contact.Player = s.GUID().Counter()
		contact.Friend = friend.Counter()
		if _, err := s.DB().Insert(&contact); err != nil {
			panic(err)
		}
		return &contact
	}
	return &contact
}

func (s *Session) HandleFriendAdd(e *etc.Buffer) {
	name := e.ReadCString()
	id, err := s.WS.GetGUIDByPlayerName(name)
	if err != nil {
		s.SendFriendStatus(packet.FRIEND_NOT_FOUND, guid.Nil, "", 0, 0, 0, 0)
		return
	}

	if id == s.GUID() {
		s.SendFriendStatus(packet.FRIEND_SELF, guid.Nil, "", 0, 0, 0, 0)
		return
	}

	contact := s.GetContact(id)
	if contact.Friended {
		s.SendFriendStatus(packet.FRIEND_ALREADY, guid.Nil, "", 0, 0, 0, 0)
		return
	}

	contact.Friended = true
	s.DB().Where("player = ?", s.GUID().Counter()).Where("friend = ?", id.Counter()).Cols("friended").Update(contact)

	status := s.WS.GetFriendStatus(id)

	var area, level, class uint32
	if status&packet.FriendOnline != 0 {
		session, err := s.WS.GetSessionByGUID(id)
		if err != nil {
			s.SendFriendStatus(packet.FRIEND_DB_ERROR, guid.Nil, "", 0, 0, 0, 0)
			return
		}

		area = session.ZoneID
		level = uint32(session.GetLevel())
		class = uint32(session.GetPlayerClass())
		s.SendFriendStatus(packet.FRIEND_ADDED_ONLINE, id, contact.Note, status, area, level, class)
	} else {
		s.SendFriendStatus(packet.FRIEND_ADDED_OFFLINE, id, contact.Note, status, area, level, class)
	}
}

func (s *Session) HandleFriendDelete(e *etc.Buffer) {
	id := s.decodeUnpackedGUID(e)
	if id == guid.Nil {
		return
	}

	if id == s.GUID() {
		s.SendFriendStatus(packet.FRIEND_SELF, guid.Nil, "", 0, 0, 0, 0)
		return
	}

	var contact wdb.Contact
	found, err := s.DB().Where("player = ?", s.GUID().Counter()).Where("friend = ?", id.Counter()).Get(&contact)
	if err != nil {
		panic(err)
	}

	if !found {
		s.SendFriendStatus(packet.FRIEND_NOT_FOUND, guid.Nil, "", 0, 0, 0, 0)
		return
	}

	if contact.Friended == false {
		s.SendFriendStatus(packet.FRIEND_ALREADY, guid.Nil, "", 0, 0, 0, 0)
		return
	}

	contact.Friended = false
	s.DB().Where("player = ?", s.GUID().Counter()).Where("friend = ?", id.Counter()).Cols("friended").Update(&contact)
	s.SendFriendStatus(packet.FRIEND_REMOVED, id, "", 0, 0, 0, 0)
}

func (s *Session) HandleContactListRequest(b []byte) {
	s.SendSocialList()
}

func (s *Session) BroadcastStatus(status packet.FriendStatus) {
	player := s.GUID().Counter()

	var contacts []wdb.Contact
	s.DB().Where("friend = ?", player).Where("friended = 1").Find(&contacts)

	for _, contact := range contacts {
		id := guid.RealmSpecific(guid.Player, s.WS.RealmID(), contact.Player)
		sess, _ := s.WS.GetSessionByGUID(id)
		if sess != nil {
			var result uint8
			var area, level, class uint32
			if status&packet.FriendOnline != 0 {
				result = packet.FRIEND_ONLINE
				area = s.ZoneID
				level = uint32(s.GetLevel())
				class = uint32(s.GetPlayerClass())
			} else {
				result = packet.FRIEND_OFFLINE
			}

			sess.SendFriendStatus(result, s.GUID(), contact.Note, status, area, level, class)
		}
	}
}

func (s *Session) HandleIgnoreAdd(e *etc.Buffer) {
	name := e.ReadCString()

	var char wdb.Character
	found, err := s.DB().Where("name = ?", name).Get(&char)
	if err != nil {
		panic(err)
	}

	if !found {
		s.SendFriendStatus(packet.FRIEND_IGNORE_NOT_FOUND, guid.Nil, "", 0, 0, 0, 0)
		return
	}

	id := guid.RealmSpecific(guid.Player, s.WS.RealmID(), char.ID)

	if id == s.GUID() {
		s.SendFriendStatus(packet.FRIEND_IGNORE_SELF, guid.Nil, "", 0, 0, 0, 0)
		return
	}

	contact := s.GetContact(id)
	if contact.Ignored == true {
		s.SendFriendStatus(packet.FRIEND_IGNORE_ALREADY, id, "", 0, 0, 0, 0)
		return
	}

	contact.Ignored = true

	if _, err := s.DB().Where("player = ?", s.GUID().Counter()).Where("friend = ?", id.Counter()).Cols("ignored").Update(contact); err != nil {
		panic(err)
	}

	s.SendFriendStatus(packet.FRIEND_IGNORE_ADDED, id, "", 0, 0, 0, 0)
}

func (s *Session) HandleIgnoreDelete(e *etc.Buffer) {
	id := s.decodeUnpackedGUID(e)

	if id == s.GUID() {
		s.SendFriendStatus(packet.FRIEND_IGNORE_SELF, guid.Nil, "", 0, 0, 0, 0)
		return
	}

	var contact wdb.Contact
	found, err := s.DB().Where("player = ?", s.GUID().Counter()).Where("friend = ?", id.Counter()).Get(&contact)
	if err != nil {
		panic(err)
	}

	if !found {
		s.SendFriendStatus(packet.FRIEND_IGNORE_NOT_FOUND, id, "", 0, 0, 0, 0)
		return
	}

	yo.Spew(contact)

	if contact.Ignored == false {
		s.SendFriendStatus(packet.FRIEND_IGNORE_ALREADY, id, "", 0, 0, 0, 0)
		return
	}

	contact.Ignored = false
	s.DB().Where("player = ?", s.GUID().Counter()).Where("friend = ?", id.Counter()).Cols("ignored").Update(&contact)
	s.SendFriendStatus(packet.FRIEND_IGNORE_REMOVED, id, "", 0, 0, 0, 0)
}
