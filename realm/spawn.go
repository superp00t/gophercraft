package realm

import (
	"encoding/hex"
	"fmt"
	"io"
	"time"

	"github.com/superp00t/etc"
	"github.com/superp00t/etc/yo"
	"github.com/superp00t/gophercraft/format/dbc"
	"github.com/superp00t/gophercraft/gcore"
	"github.com/superp00t/gophercraft/guid"
	"github.com/superp00t/gophercraft/packet"
	"github.com/superp00t/gophercraft/packet/update"
	"github.com/superp00t/gophercraft/vsn"
	"github.com/superp00t/gophercraft/realm/wdb"
)

var (
	DefaultSpeeds = update.Speeds{
		update.Walk:           2.5,
		update.Run:            7,
		update.RunBackward:    4.5,
		update.Swim:           4.722222,
		update.SwimBackward:   2.5,
		update.Turn:           3.141594,
		update.Flight:         7.0,
		update.FlightBackward: 4.7222,
		update.Pitch:          3.14,
	}
)

func (s *Session) SendSystemFeatures() {
	if s.Build().AddedIn(vsn.V2_4_3) {
		features := packet.NewWorldPacket(packet.SMSG_FEATURE_SYSTEM_STATUS)
		features.WriteByte(2) // Can complain (0 = disabled, 1 = enabled, don't auto ignore, 2 = enabled, auto ignore)
		features.WriteByte(1) // voice chat toggled
		s.SendAsync(features)
		yo.Println("Sent features")
	}
}

func (s *Session) SetupOnLogin() {
	// s.SendNameQueryResponseFor(s.Char)
	if s.Build().AddedIn(vsn.V1_12_1) {
		s.SendVerifyLoginPacket()
	}

	s.SendAccountDataTimes()

	// time.Sleep(3 * time.Second)
	s.SendTutorialFlags()
	s.SendSystemFeatures()

	s.BindpointUpdate()

	s.MOTD("G O P H E R C R A F T\n"+
		"   Version %s", gcore.Version)

	// s.SendRestStart()

	s.SendSpellList()
	// s.SendUnlearnSpell()
	s.SendActionButtons()

	if s.Build().AddedIn(vsn.V1_12_1) {
		// s.SendTutorialFlags()
		// s.SendFactions()
	}
	// s.SendInitWorldStates()
	s.SetTimeSpeed()

	s.SpawnPlayer()

	s.SyncTime()

	s.BroadcastStatus(packet.FriendOnline)
	s.InitGroup()
	s.SendSocialList()

	// Show cinematic sequence on first login
	if s.Char.FirstLogin && s.Config().Bool("Char.StartingCinematic") {
		var race *dbc.Ent_ChrRaces
		s.DB().GetData(uint32(s.Char.Race), &race)

		if race != nil && race.CinematicSequenceID != 0 {
			p := packet.NewWorldPacket(packet.SMSG_TRIGGER_CINEMATIC)
			p.WriteUint32(race.CinematicSequenceID)
			s.SendAsync(p)
		}

		// Don't show same cinematic twice.
	}

	s.Char.FirstLogin = false
	if _, err := s.DB().Cols("first_login").Update(s.Char); err != nil {
		panic(err)
	}

	go func() {
		time.Sleep(500 * time.Millisecond)

		s.SystemChat("|TInterface\\OptionsFrame\\NvidiaLogo:128:300:0:0:128:64:0:0:0:0|t")
		s.SystemChat("|TInterface\\Icons\\rats:256:512:0:0:128:64:0:0:0:0|t")
	}()

	// s.SendLoginSpell()
}

func packTime(t time.Time) int32 {
	year, month, day := t.Date()
	tm_year := int32(year)
	tm_mon := int32(month)
	tm_mday := int32(day)
	tm_wday := int32(t.Weekday())
	tm_hour := int32(t.Hour())
	tm_min := int32(t.Minute())

	return (tm_year-100)<<24 | tm_mon<<20 | (tm_mday-1)<<14 | tm_wday<<11 | tm_hour<<6 | tm_min
}

func (s *Session) SendInitWorldStates() {
	p := packet.NewWorldPacket(packet.SMSG_INIT_WORLD_STATES)
	p.WriteUint32(0)
	s.SendAsync(p)
}

func (s *Session) SetTimeSpeed() {
	pkt := packet.NewWorldPacket(packet.SMSG_LOGIN_SETTIMESPEED)
	pkt.WriteInt32(packTime(time.Now()))
	pkt.WriteFloat32(0.01666667)

	if s.Build().AddedIn(vsn.V3_3_5a) {
		pkt.WriteUint32(0)
	}

	s.SendAsync(pkt)

	yo.Ok("Send gamespeed")
}

func (s *Session) SendVerifyLoginPacket() {
	v := packet.NewWorldPacket(packet.SMSG_LOGIN_VERIFY_WORLD)

	v.WriteUint32(s.Char.Map)
	v.WriteFloat32(s.Char.X)
	v.WriteFloat32(s.Char.Y)
	v.WriteFloat32(s.Char.Z)
	v.WriteFloat32(s.Char.O)

	s.SendAsync(v)
	yo.Ok("Sent verify login packet")
}

func (s *Session) SendLoginFailure(failure packet.CharLoginResult) {
	p := packet.NewWorldPacket(packet.SMSG_CHARACTER_LOGIN_FAILED)
	result, ok := packet.CharLoginResultDescriptors[s.Build()][failure]
	if !ok {
		panic(fmt.Sprintf("no result found for %v", failure))
	}
	p.WriteByte(result)
	s.SendAsync(p)
}

func (s *Session) SendRestStart() {
	if s.Build().RemovedIn(vsn.V3_3_5a) {
		v := packet.NewWorldPacket(packet.SMSG_SET_REST_START)
		v.WriteUint32(0)
		s.SendAsync(v)
	}
}

func (s *Session) SendAccountDataTimes() {
	const (
		GLOBAL_CONFIG_CACHE      = 0x1
		CHARACTER_CONFIG_CACHE   = 0x2
		GLOBAL_BINDINGS_CACHE    = 0x4
		CHARACTER_BINDINGS_CACHE = 0x8
		GLOBAL_MACROS_CACHE      = 0x10
		CHARACTER_MACROS_CACHE   = 0x20
		CHARACTER_LAYOUT_CACHE   = 0x40
		CHARACTER_CHAT_CACHE     = 0x80
		//
		GLOBAL    = GLOBAL_BINDINGS_CACHE | GLOBAL_CONFIG_CACHE | GLOBAL_MACROS_CACHE
		CHARACTER = CHARACTER_CONFIG_CACHE | CHARACTER_BINDINGS_CACHE | CHARACTER_MACROS_CACHE | CHARACTER_LAYOUT_CACHE | CHARACTER_CHAT_CACHE
		ALL       = GLOBAL | CHARACTER
	)

	v := packet.NewWorldPacket(packet.SMSG_ACCOUNT_DATA_TIMES)
	if s.Build().AddedIn(vsn.V3_3_5a) {
		v.WriteInt32(int32(time.Now().Unix()))
		v.WriteByte(1)
		v.WriteUint32(ALL)
		for i := 0; i < 8; i++ {
			v.WriteUint32(0)
		}
	} else {
		for i := 0; i < 32; i++ {
			v.WriteUint32(0)
		}
	}

	s.SendAsync(v)
	yo.Println("Account data times sent.")
}

func (s *Session) SendTutorialFlags() {
	v3 := packet.NewWorldPacket(packet.SMSG_TUTORIAL_FLAGS)
	for i := 0; i < 8; i++ {
		v3.WriteUint32(0xFFFFFFFF)
	}
	s.SendAsync(v3)
	yo.Println("Tutorial flags sent.")
}

func (s *Session) HandleAccountDataUpdate(data []byte) {
	yo.Spew(data)
}

func (s *Session) SendFactions() {
	d, _ := hex.DecodeString("80000000020000000000000000000200000000020000000010000000000000000000020000000000000000001600000000000000000008000000000e00000000190000000000000000001100000000110000000011000000001100000000060000000006000000000600000000060000000004000000000400000000040000000004000000000400000000060000000000000000000400000000040000000004000000000400000000040000000004000000000200000000000000000010000000000200000000100000000002000000001000000000000000000000000000001000000000060000000010000000000000000000180000000006000000001000000000100000000010000000000200000000020000000011000000001000000000100000000050000000001c0000000010000000001000000000500000000010000000001000000000000000000002000000001000000000040000000010000000001000000000020000000010000000001000000000000000000000000000000000000000000000000000000000000000000000100000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000")
	p := packet.NewWorldPacket(packet.SMSG_INITIALIZE_FACTIONS)
	p.Buffer = etc.FromBytes(d)
	s.SendAsync(p)
}

func (s *Session) GetPlayerRace() packet.Race {
	return packet.Race(s.GetByte("Race"))
}

func (s *Session) GetLevel() int {
	return int(s.GetUint32("Level"))
}

func (s *Session) isTrackedGUID(g guid.GUID) bool {
	for _, v := range s.TrackedGUIDs {
		if v == g {
			return true
		}
	}
	return false
}

func (s *Session) ChangeDefaultSpeeds(modifier float32) {
	s.MoveSpeeds = make(update.Speeds)
	for speedType, speed := range DefaultSpeeds {
		s.MoveSpeeds[speedType] = speed * modifier
	}
}

func (s *Session) SyncTime() {
	if s.Build().AddedIn(vsn.V2_4_3) {
		p := packet.NewWorldPacket(packet.SMSG_TIME_SYNC_REQ)
		p.WriteUint32(s.WS.UptimeMS())
		s.SendAsync(p)
		yo.Println("Synced time with client")
	}
}

// SpawnPlayer initializes the player into the object manager and sends the packets needed to log the client into the world.
func (s *Session) SpawnPlayer() {
	s.WS.PlayersL.Lock()
	s.WS.PlayerList[s.PlayerName()] = s
	s.WS.PlayersL.Unlock()

	var exploredZones []wdb.ExploredZone
	s.DB().Where("player = ?", s.PlayerID()).Find(&exploredZones)

	// fill out attribute fields
	s.MovementInfo = &update.MovementInfo{
		Flags: 0,
		Time:  s.WS.UptimeMS(),
		Position: update.Position{
			Point3: update.Point3{
				X: s.Char.X,
				Y: s.Char.Y,
				Z: s.Char.Z,
			},
			O: s.Char.O,
		},
	}

	s.ChangeDefaultSpeeds(1.0)

	var err error
	s.ValuesBlock, err = update.NewValuesBlock(
		s.Build(),
		guid.TypeMaskObject|guid.TypeMaskUnit|guid.TypeMaskPlayer,
	)

	if err != nil {
		panic(err)
	}

	s.InitInventoryManager()

	s.SetGUID("GUID", s.GUID())
	s.SetFloat32("ScaleX", 1.0)
	s.SetUint32("Health", 80)
	s.SetUint32("MaxHealth", 80)
	s.SetUint32("Mana", 4143)
	s.SetUint32("MaxMana", 4143)
	s.SetUint32("Energy", 100)
	s.SetUint32("MaxRage", 1000)
	s.SetUint32("MaxEnergy", 100)
	s.SetUint32("Level", uint32(s.Char.Level))
	s.SetUint32("FactionTemplate", 1)

	if s.Build().AddedIn(vsn.V2_4_3) {
		s.SetUint32("MaxLevel", uint32(s.GetMaxLevel()))
	}

	s.SetByte("Race", uint8(s.Char.Race))
	s.SetByte("Class", uint8(s.Char.Class))
	s.SetByte("Gender", uint8(s.Char.Gender))
	s.SetByte("Power", PowerType(packet.Class(s.Char.Class)))

	if s.Build().AddedIn(vsn.V1_12_1) {
		s.SetByte("PlayerGender", uint8(s.Char.Gender))
	}

	// Player flags
	s.SetBit("PlayerControlled", true)
	// s.SetBit("Resting", true)
	// s.SetBit("AurasVisible", true)

	s.SetUint32("BaseAttackTime", 2900)
	s.SetUint32("OffhandAttackTime", 2000)

	s.SetFloat32("BoundingRadius", 1.0)
	s.SetFloat32("CombatReach", 1.0)

	s.SetUint32("DisplayID", s.WS.GetNative(packet.Race(s.Char.Race), s.Char.Gender))
	if s.Build().AddedIn(vsn.V1_12_1) {
		s.SetUint32("NativeDisplayID", s.WS.GetNative(packet.Race(s.Char.Race), s.Char.Gender))

		for _, ez := range exploredZones {
			var area *dbc.Ent_AreaTable
			s.DB().GetData(ez.ZoneID, &area)
			if area != nil {
				s.SetExplorationFlag(area.ExploreFlag)
			}
		}
	}

	if s.Build().AddedIn(vsn.V1_12_1) {
		s.SetFloat32("MinDamage", 50)
		s.SetFloat32("MaxDamage", 50)
		s.SetUint32("MinOffhandDamage", 50)
		s.SetUint32("MaxOffhandDamage", 50)
	} else {
		s.SetFloat32("Damage", 50)
	}

	s.SetByte("LoyaltyLevel", 0xEE)

	s.SetFloat32("ModCastSpeed", 30)

	s.SetUint32("BaseMana", 60)
	// todo: replace with bit fields
	// s.SetBit("AuraByteFlagSupportable, true)
	// s.SetBit("AuraByteFlagNoDispel, true)
	s.SetByte("AuraByteFlags", 0x08|0x20)

	if s.Build().AddedIn(vsn.V1_12_1) {
		s.SetInt32("AttackPower", 20)
		s.SetInt32("AttackPowerMods", 0)

		s.SetInt32("RangedAttackPower", 1)
		s.SetInt32("RangedAttackPowerMods", 0)

		s.SetFloat32("MinRangedDamage", 0)
		s.SetFloat32("MaxRangedDamage", 0)
	}

	s.SetByte("Skin", s.Char.Skin)
	s.SetByte("Face", s.Char.Face)
	s.SetByte("HairStyle", s.Char.HairStyle)
	s.SetByte("HairColor", s.Char.HairColor)

	s.SetByte("FacialHair", s.Char.FacialHair)
	s.SetByte("BankBagSlotCount", 8)
	s.SetByte("RestState", 0x01)

	for i := 0; i < 5; i++ {
		s.SetUint32ArrayValue("Stats", i, 10)
	}

	s.SetByte("Gender", s.Char.Gender)

	s.SetUint32("XP", s.Char.XP)
	s.SetUint32("NextLevelXP", s.GetNextLevelXP())

	s.SetUint32ArrayValue("CharacterPoints", 0, 51)
	s.SetUint32ArrayValue("CharacterPoints", 1, 2)

	s.SetFloat32("BlockPercentage", 4.0)
	s.SetFloat32("DodgePercentage", 4.0)
	s.SetFloat32("ParryPercentage", 4.0)

	if s.Build().AddedIn(vsn.V1_12_1) {
		s.SetFloat32("CritPercentage", 4.0)
		s.SetUint32("RestStateExperience", 200)
	}

	if s.Build().AddedIn(vsn.V2_4_3) {
		s.SetFloat32("RangedCritPercentage", 4.0)
	}

	s.SetInt32("Coinage", int32(s.Char.Coinage))

	if s.Build().AddedIn(vsn.V2_4_3) {

		s.SetStructArrayValue("SkillInfos", 0, "ID", uint16(98))
		s.SetStructArrayValue("SkillInfos", 0, "SkillLevel", uint16(300))
		s.SetStructArrayValue("SkillInfos", 0, "SkillCap", uint16(300))
		s.SetStructArrayValue("SkillInfos", 0, "Bonus", uint32(0))

		s.SetStructArrayValue("SkillInfos", 1, "ID", uint16(109))
		s.SetStructArrayValue("SkillInfos", 1, "SkillLevel", uint16(300))
		s.SetStructArrayValue("SkillInfos", 1, "SkillCap", uint16(300))
		s.SetStructArrayValue("SkillInfos", 1, "Bonus", uint32(0))
	}

	if s.Build().AddedIn(vsn.V3_3_5a) {
		s.SetBit("RegeneratePower", true)
	}

	s.SetInt32("WatchedFactionIndex", -1)

	// s.ClearChanges()

	s.CurrentPhase = "main"
	s.CurrentMap = s.Char.Map
	s.ZoneID = s.Char.Zone

	// send player create packet of themself
	s.State = InWorld
	s.SendObjectCreate(s)
	yo.Ok("Sent spawn packet.")

	cMap := s.Map()

	// add our player to map, and notify nearby players of their presence
	cMap.AddObject(s)

	// notify our player of nearby objects.
	nearbyObjects := cMap.NearObjects(s)
	createObjects := make([]Object, len(nearbyObjects))

	for i := range nearbyObjects {
		createObjects[i] = nearbyObjects[i]
	}

	for _, wo := range nearbyObjects {
		s.TrackedGUIDs = append(s.TrackedGUIDs, wo.GUID())
	}
	s.SendObjectCreate(createObjects...)
}

func (s *Session) BindpointUpdate() {
	//goldshire
	p := packet.NewWorldPacket(packet.SMSG_BINDPOINTUPDATE)
	p.WriteFloat32(-8949.95)
	p.WriteFloat32(-132.493)
	p.WriteFloat32(83.5312)
	p.WriteUint32(0)
	p.WriteUint32(12)

	s.SendAsync(p)
}

func (s *Session) HandleNameQuery(e *etc.Buffer) {
	in := e.ReadUint64()

	yo.Warnf("Name query 0x%016X\n", in)

	g := guid.Classic(in)

	yo.Warn(g)

	var chars []wdb.Character

	s.WS.DB.Where("id = ?", g.Counter()).Find(&chars)
	if len(chars) == 0 {
		yo.Warn("No such data exists for", g)
		return
	}

	s.SendNameQueryResponseFor(&chars[0])
}

func (s *Session) SendNameQueryResponseFor(char *wdb.Character) {
	resp := packet.NewWorldPacket(packet.SMSG_NAME_QUERY_RESPONSE)
	id := guid.RealmSpecific(guid.Player, s.WS.RealmID(), char.ID)
	if s.Build().AddedIn(vsn.V3_3_5a) {
		id.EncodePacked(s.Build(), resp)
	} else {
		id.EncodeUnpacked(s.Build(), resp)
	}

	if s.Build().AddedIn(vsn.V3_3_5a) {
		if char == nil || char.Name == "" {
			resp.WriteByte(1)
			s.SendAsync(resp)
			return
		} else {
			resp.WriteByte(0)
		}
	}

	resp.WriteCString(char.Name)
	// resp.WriteCString(s.Config().RealmName)
	resp.WriteByte(0)
	resp.WriteUint32(uint32(char.Race))
	resp.WriteUint32(uint32(char.Gender))
	resp.WriteUint32(uint32(char.Class))

	if s.Build().AddedIn(vsn.V2_4_3) {
		resp.WriteByte(0)
	}

	s.SendAsync(resp)
}

func (s *Session) encodePackedGUID(wr io.Writer, g guid.GUID) {
	g.EncodePacked(s.Build(), wr)
}

func (ws *Server) RemovePlayerFromList(name string) {
	ws.PlayersL.Lock()
	delete(ws.PlayerList, name)
	ws.PlayersL.Unlock()
}

func (s *Session) CleanupPlayer() {
	s.WS.RemovePlayerFromList(s.PlayerName())
	s.BroadcastStatus(packet.FriendOffline)

	if s.State == InWorld {
		s.Map().RemoveObject(s.GUID())
	}

	if s.Group != nil {
		s.Group.UpdateList()
		s.Group = nil
	}

	s.MovementInfo = nil
	s.CurrentMap = 0
	s.CurrentPhase = ""
	s.Char = nil
	s.TrackedGUIDs = nil
}

func (s *Session) HandleLogoutRequest(b []byte) {
	if s.State != InWorld {
		return
	}

	// TODO: deny if in combat
	// TODO: Impose timeout if configured

	if s.Build().AddedIn(vsn.V1_12_1) {
		resp := packet.NewWorldPacket(packet.SMSG_LOGOUT_RESPONSE)
		resp.WriteUint32(0)
		resp.WriteByte(0)
		s.SendAsync(resp)
	}

	s.CleanupPlayer()

	s.State = CharacterSelectMenu

	resp := packet.NewWorldPacket(packet.SMSG_LOGOUT_COMPLETE)
	s.SendAsync(resp)
}
