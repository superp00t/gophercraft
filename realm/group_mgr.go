package realm

import (
	"fmt"
	"sync"

	"github.com/superp00t/gophercraft/packet/chat"
	"github.com/superp00t/gophercraft/realm/wdb"
	"github.com/superp00t/gophercraft/vsn"

	"github.com/superp00t/etc"
	"github.com/superp00t/etc/yo"
	"github.com/superp00t/gophercraft/guid"
	"github.com/superp00t/gophercraft/packet"
)

func (s *Session) SendPartyResult(operation packet.PartyOperation, memberName string, result packet.PartyResult) {
	p := packet.NewWorldPacket(packet.SMSG_PARTY_COMMAND_RESULT)
	p.WriteUint32(uint32(operation))
	p.WriteCString(memberName)
	if err := packet.EncodePartyResult(s.Build(), p, result); err != nil {
		panic(err)
	}
	s.SendAsync(p)
}

type Group struct {
	sync.Mutex
	GroupType     uint8
	Server        *Server
	Leader        guid.GUID
	Members       []guid.GUID
	LootMethod    uint8
	LootThreshold uint8
}

func (g *Group) Disband() {
	// Delete party membership from database
	g.SetLeader(guid.Nil)

	for _, member := range g.Members {
		sess, err := g.Server.GetSessionByGUID(member)
		if err != nil {
			continue
		}

		sess.SendGroupDestroyed()
		sess.Group = nil
		sess.GroupInvite = guid.Nil
	}
}

func (g *Group) Empty() bool {
	g.Lock()
	ln := len(g.Members)
	g.Unlock()

	return ln == 0
}

func (g *Group) Add(session *Session) {
	g.Members = append(g.Members, session.GUID())

	for _, member := range g.Members {
		g.Server.DB.Where("id = ?", member.Counter()).Cols("leader").Update(&wdb.Character{
			Leader: g.Leader.Counter(),
		})
	}

	g.UpdateList()
}

func (g *Group) RemoveMember(id guid.GUID) {
	for i, gid := range g.Members {
		if gid == id {
			g.Members = append(g.Members[:i], g.Members[i+1:]...)
			g.UpdateList()
			return
		}
	}
}

func (s *Session) Team() uint32 {
	return s.GetUint32("FactionTemplate")
}

func (s *Session) HasYouIgnored(g guid.GUID) bool {
	if s.IsAdmin() {
		return false
	}

	return false
}

func (s *Session) SendGroupInvite(from string) {
	p := packet.NewWorldPacket(packet.SMSG_GROUP_INVITE)
	p.WriteCString(from)
	s.SendAsync(p)
}

func (s *Session) HandleGroupInvite(e *etc.Buffer) {
	playerName := e.ReadCString()

	receiver, err := s.WS.GetSessionByPlayerName(playerName)
	if err != nil {
		s.SendPartyResult(packet.PartyInvite, playerName, packet.PartyBadPlayerName)
		return
	}

	if receiver.GUID() == s.GUID() {
		s.SendPartyResult(packet.PartyInvite, playerName, packet.PartyBadPlayerName)
		return
	}

	if s.HasYouIgnored(receiver.GUID()) {
		s.SendPartyResult(packet.PartyInvite, playerName, packet.PartyIgnoringYou)
		return
	}

	// Disallow cross-factional parties unless you are a GM OR the server has explicitly allowed them.
	if !s.IsGM() && !s.Config().Bool("PVP.CrossFactionGroups") && s.Team() != receiver.Team() {
		s.SendPartyResult(packet.PartyInvite, playerName, packet.PartyWrongFaction)
		return
	}

	if s.Group != nil {
		s.Group.Lock()
		if s.GUID() != s.Group.Leader {
			s.SendPartyResult(packet.PartyInvite, playerName, packet.PartyNotLeader)
			return
		}
		s.Group.Unlock()
	} else {
		s.GroupInvite = guid.Nil
		if s.Group != nil {
			s.Group.RemoveMember(s.GUID())
		}

		s.Group = new(Group)
		s.Group.Server = s.WS
		s.Group.Leader = s.GUID()
		s.Group.Members = append(s.Group.Members, s.GUID())
	}

	s.SendPartyResult(packet.PartyInvite, playerName, packet.PartyOK)

	receiver.GroupInvite = s.GUID()
	receiver.SendGroupInvite(s.PlayerName())
}

func (s *Session) HandleGroupDecline() {
	if s.GroupInvite == guid.Nil {
		return
	}

	inviter, err := s.WS.GetSessionByGUID(s.GroupInvite)
	if err != nil {
		return
	}

	if inviter.Group == nil {
		return
	}

	if inviter.Group.Empty() {
		inviter.Group = nil
	}

	p := packet.NewWorldPacket(packet.SMSG_GROUP_DECLINE)
	p.WriteCString(s.PlayerName())
	inviter.SendAsync(p)
}

func (s *Session) HandleGroupAccept() {
	if s.GroupInvite == guid.Nil {
		return
	}

	player, err := s.WS.GetSessionByGUID(s.GroupInvite)
	if err != nil {
		return
	}

	group := player.Group
	if group == nil {
		return
	}

	s.Group = group

	group.Add(s)
}

func (s *Session) SendSetLeader(leaderName string) {
	p := packet.NewWorldPacket(packet.SMSG_GROUP_SET_LEADER)
	p.WriteCString(leaderName)
	s.SendAsync(p)
}

func (s *Session) SendGroupList() {
	if s.Group == nil {
		return
	}

	s.Group.Lock()
	defer s.Group.Unlock()

	p := packet.NewWorldPacket(packet.SMSG_GROUP_LIST)

	if s.Build().AddedIn(vsn.V1_12_1) {
		p.WriteByte(s.Group.GroupType)
		p.WriteByte(0)
	}

	p.WriteUint32(uint32(len(s.Group.Members) - 1))

	if s.GUID() == s.Group.Leader {
		s.SendSetLeader(s.PlayerName())
	}

	// Alpha
	if s.Build().RemovedIn(vsn.V1_12_1) {
		name, _ := s.WS.GetPlayerNameByGUID(s.Group.Leader)
		p.WriteCString(name)
		s.Group.Leader.EncodeUnpacked(s.Build(), p)
		p.WriteByte(1)
	}

	for _, member := range s.Group.Members {
		if member == s.GUID() {
			continue
		}

		var flags uint8

		str, err := s.WS.GetPlayerNameByGUID(member)
		if err != nil {
			yo.Warn(err)
			str = "???"
			flags |= packet.MemberOffline
		} else {
			flags |= packet.MemberOnline
		}

		if member == s.Group.Leader {
			s.SendSetLeader(str)
		}

		p.WriteCString(str)
		member.EncodeUnpacked(s.Build(), p)
		if s.Build().AddedIn(vsn.V1_12_1) {
			p.WriteByte(flags)
		}
		p.WriteByte(0)
	}
	s.Group.Leader.EncodeUnpacked(s.Build(), p)

	p.WriteByte(s.Group.LootMethod)
	s.Group.Leader.EncodeUnpacked(s.Build(), p)
	if s.Build().AddedIn(vsn.V1_12_1) {
		p.WriteByte(s.Group.LootThreshold)
	}

	s.SendAsync(p)
}

func (g *Group) SetLeader(id guid.GUID) {
	char := wdb.Character{
		Leader: id.Counter(),
	}
	if _, err := g.Server.DB.Where("leader = ?", g.Leader.Counter()).Cols("leader").Update(&char); err != nil {
		panic(err)
	}
}

func (g *Group) UpdateList() {
	for _, member := range g.Members {
		sess, err := g.Server.GetSessionByGUID(member)
		if err == nil {
			sess.SendGroupList()
		} else {
			yo.Warn(err)
		}
	}
}

func (s *Session) SendGroupDestroyed() {
	p := packet.NewWorldPacket(packet.SMSG_GROUP_DESTROYED)
	s.SendAsync(p)
}

func (s *Session) LeaveGroup() {
	char := wdb.Character{
		Leader: 0,
	}
	s.DB().Where("id = ?", s.GUID().Counter()).Cols("leader").Update(&char)
	if s.Group != nil {
		// if len(s.Group.Members) > 2 {
		var newLeaderGUID guid.GUID
		// Set next player to leader
		for _, member := range s.Group.Members {
			if member != s.GUID() {
				newLeaderGUID = member
				break
			}
		}

		if newLeaderGUID != guid.Nil {
			s.Group.SetLeader(newLeaderGUID)
		}

		s.Group.RemoveMember(s.GUID())
		// } else {
		// 	s.Group.Disband()
		// }

		s.Group.UpdateList()
		s.Group = nil
	}

	s.SendGroupDestroyed()
}

func (s *Session) InitGroup() {
	if s.Char.Leader != 0 {
		var members []wdb.Character
		s.DB().Where("leader = ?", s.Char.Leader).Find(&members)

		// See if group already exists.
		for _, member := range members {
			sess, err := s.WS.GetSessionByGUID(guid.RealmSpecific(guid.Player, s.WS.RealmID(), member.ID))
			if err == nil {
				if sess.Group != nil {
					s.Group = sess.Group
					break
				}
			}
		}

		if s.Group == nil {
			// No one appears to be online, instantiate a new Group object.
			s.Group = &Group{
				Server: s.WS,
				// TODO: what should happen if the leader character was deleted?
				Leader: guid.RealmSpecific(guid.Player, s.WS.RealmID(), s.Char.Leader),
			}

			for _, memb := range members {
				s.Group.Members = append(s.Group.Members, guid.RealmSpecific(guid.Player, s.WS.RealmID(), memb.ID))
			}
		}

		s.Group.UpdateList()
	}
}

func (s *Session) HandleGroupDisband() {
	// will disband if only two players are currently there.
	s.LeaveGroup()
}

const statsMaskAll = packet.GroupUpdateNone |
	packet.GroupUpdateStatus |
	packet.GroupUpdateCurrentHealth |
	packet.GroupUpdateMaxHealth |
	packet.GroupUpdatePowerType |
	packet.GroupUpdateCurrentPower |
	packet.GroupUpdateMaxPower |
	packet.GroupUpdateLevel |
	packet.GroupUpdateZone |
	packet.GroupUpdatePosition | packet.GroupUpdateAuras

func (s *Session) HandleRequestPartyMemberStats(e *etc.Buffer) {
	id := s.decodeUnpackedGUID(e)

	if s.Group == nil {
		return
	}

	s.Group.Lock()

	found := false

	for _, member := range s.Group.Members {
		if member == id {
			found = true
			break
		}
	}

	if !found {
		return
	}

	s.Group.Unlock()

	player, err := s.WS.GetSessionByGUID(id)
	if err != nil {
		p := packet.NewWorldPacket(packet.SMSG_PARTY_MEMBER_STATS_FULL)
		id.EncodePacked(s.Build(), p)
		p.WriteUint32(packet.GroupUpdateStatus)
		p.WriteByte(uint8(packet.MemberOffline))
		s.SendAsync(p)
		return
	}

	var mask uint32
	mask = statsMaskAll

	if player.Pet() != nil {
		mask |= packet.GroupUpdatePetGUID |
			packet.GroupUpdatePetName |
			packet.GroupUpdatePetModelID |
			packet.GroupUpdatePetCurrentHP |
			packet.GroupUpdatePetMaxHP |
			packet.GroupUpdatePetPowerType |
			packet.GroupUpdatePetCurrentPower |
			packet.GroupUpdatePetMaxPower |
			packet.GroupUpdatePetAuras |
			packet.GroupUpdatePet
	}

	s.SendPartyMemberStats(mask, player)
}

func (s *Session) SendPartyMemberStats(mask uint32, player *Session) {
	p := packet.NewWorldPacket(packet.SMSG_PARTY_MEMBER_STATS_FULL)
	player.GUID().EncodePacked(s.Build(), p)

	pet := player.Pet()

	if pet == nil && mask&packet.GroupUpdatePetAuras != 0 {
		panic(fmt.Sprintf("0x%08X\n", mask))
	}

	p.WriteUint32(mask)

	if mask&packet.GroupUpdateStatus != 0 {
		p.WriteByte(uint8(packet.MemberOnline))
	}

	if mask&packet.GroupUpdateCurrentHealth != 0 {
		p.WriteUint16(uint16(player.Health()))
	}

	if mask&packet.GroupUpdateMaxHealth != 0 {
		p.WriteUint16(uint16(player.MaxHealth()))
	}

	if mask&packet.GroupUpdatePowerType != 0 {
		p.WriteByte(player.GetPowerType())
	}

	if mask&packet.GroupUpdateCurrentPower != 0 {
		p.WriteUint16(uint16(player.Power()))
	}

	if mask&packet.GroupUpdateMaxPower != 0 {
		p.WriteUint16(uint16(player.MaxPower()))
	}

	if mask&packet.GroupUpdateLevel != 0 {
		p.WriteUint16(uint16(player.GetLevel()))
	}

	if mask&packet.GroupUpdateZone != 0 {
		p.WriteUint16(uint16(player.ZoneID))
	}

	if mask&packet.GroupUpdatePosition != 0 {
		p.WriteUint16(uint16(player.Position().X))
		p.WriteUint16(uint16(player.Position().Y))
	}

	if mask&packet.GroupUpdateAuras != 0 {
		var auraMask uint64 = 0xFFFFFFFF
		p.WriteUint32(uint32(auraMask))
		auras := player.GetUint32Slice("Auras")
		for i := 0; i < s.MaxPositiveAuras(); i++ {
			if auraMask&(uint64(1)<<uint64(i)) != 0 {
				p.WriteUint16(uint16(auras[i]))
			}
		}
	}

	if mask&packet.GroupUpdatePetGUID != 0 {
		pet.GUID().EncodeUnpacked(s.Build(), p)
	}

	if mask&packet.GroupUpdatePetName != 0 {
		petName, _ := s.WS.GetUnitNameByGUID(pet.GUID())
		p.WriteCString(petName)
	}

	if mask&packet.GroupUpdatePetModelID != 0 {
		p.WriteUint16(uint16(pet.DisplayID()))
	}

	if mask&packet.GroupUpdatePetCurrentHP != 0 {
		p.WriteUint16(uint16(pet.Health()))
	}

	if mask&packet.GroupUpdatePetMaxHP != 0 {
		p.WriteUint16(uint16(pet.MaxHealth()))
	}

	if mask&packet.GroupUpdatePetPowerType != 0 {
		p.WriteByte(pet.GetPowerType())
	}

	if mask&packet.GroupUpdatePetCurrentPower != 0 {
		p.WriteUint16(uint16(pet.Power()))
	}

	if mask&packet.GroupUpdatePetMaxPower != 0 {
		p.WriteUint16(uint16(pet.MaxPower()))
	}

	if mask&packet.GroupUpdatePetAuras != 0 {
		var auraMask uint64 = 0xFFFFFFFF
		p.WriteUint32(uint32(auraMask))
		auras := pet.GetUint32Slice("Auras")
		for i := 0; i < s.MaxPositiveAuras(); i++ {
			if auraMask&(uint64(1)<<uint64(i)) != 0 {
				p.WriteUint16(uint16(auras[i]))
			}
		}
	}

	if mask&packet.GroupUpdateVehicleSeat != 0 {
		p.WriteUint32(s.VehicleSeatID())
	}

	s.SendAsync(p)
}

func (s *Session) HandlePartyMessage(party *chat.Message) {
	if s.Group == nil {
		return
	}

	party.SenderGUID = s.GUID()
	party.Name = s.PlayerName()
	party.ChannelName = ""
	party.Language = chat.LANG_UNIVERSAL

	s.Group.Lock()

	for _, v := range s.Group.Members {
		partyMember, err := s.WS.GetSessionByGUID(v)
		if err == nil {
			partyMember.SendChat(party)
		}
	}

	s.Group.Unlock()
}
