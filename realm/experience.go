package realm

import (
	"fmt"

	"github.com/superp00t/etc/yo"
	"github.com/superp00t/gophercraft/format/dbc"
	"github.com/superp00t/gophercraft/packet"
	"github.com/superp00t/gophercraft/vsn"
	"github.com/superp00t/gophercraft/realm/wdb"
)

func (ws *Server) GetNextLevelXP(forLevel uint32) uint32 {
	xp, ok := ws.LevelExperience[forLevel]
	if !ok {
		return 0
	}

	return xp
}

func (s *Session) GetNextLevelXP() uint32 {
	lvl := uint32(s.GetLevel())
	return s.WS.GetNextLevelXP(lvl)
}

func (s *Session) LevelUp(to uint32) {
	// Todo: reset health and other attributes

	s.SetUint32("XP", 0)
	s.SetUint32("NextLevelXP", s.GetNextLevelXP())
	s.Char.XP = 0
	s.Char.Level = to
	s.SetUint32("Level", s.Char.Level)
	s.DB().Where("id = ?", s.PlayerID()).Cols("xp", "level").Update(s.Char)

	s.UpdatePlayer()
}

func (s *Session) SetCurrentXP(current uint32) {
	s.SetUint32("XP", current)
	s.Char.XP = current
	s.DB().Where("id = ?", s.PlayerID()).Cols("xp").Update(s.Char)
}

func (s *Session) GetMaxLevel() int {
	// return 70
	return 255
}

func (s *Session) AddExperience(newXP uint32) {
	curXP := s.GetUint32("XP")
	curLevel := s.GetLevel()
	var nxtLevelXP = s.WS.GetNextLevelXP(uint32(curLevel))

	for newXP > 0 {
		// Don't add XP past level ceiling
		if curLevel == s.GetMaxLevel() {
			break
		}

		nxtLevelXP = s.WS.GetNextLevelXP(uint32(curLevel))
		if curXP+newXP >= nxtLevelXP {
			newXP -= nxtLevelXP - curXP
			if int32(newXP) < 0 {
				panic("WTF")
			}
			curXP = 0
			curLevel++
		} else {
			curXP += newXP
			break
		}

	}

	s.SetUint32("XP", curXP)
	if curLevel != s.GetLevel() {
		s.LevelUp(uint32(curLevel))
	} else {
		s.UpdateSelf()
	}
}

func (ws *Server) ExploreXPRate() float32 {
	return ws.Config.Float32("XP.Rate")
}

func (ws *Server) GetExploreXP(explorationLevel uint32) uint32 {
	return 35
}

func (s *Session) ZoneExplored(zoneID uint32) bool {
	ct, err := s.DB().Where("player = ?", s.PlayerID()).Where("zone_id = ?", zoneID).Count(new(wdb.ExploredZone))
	if err != nil {
		panic(err)
	}

	return ct > 0
}

func (s *Session) SetExplorationFlag(exploreFlag uint32) {
	blockOffset := int(exploreFlag / 32)
	bitOffset := int(exploreFlag % 32)

	sli := s.GetUint32Slice("ExploredZones")
	if blockOffset >= len(sli) {
		return
	}

	mask := sli[blockOffset]

	mask |= (1 << bitOffset)

	s.SetUint32ArrayValue("ExploredZones", blockOffset, mask)
}

func (s *Session) SendExplorationXP(zoneID, exp uint32) {
	s.Warnf("Exploring zone %d", zoneID)

	p := packet.NewWorldPacket(packet.SMSG_EXPLORATION_EXPERIENCE)
	p.WriteUint32(zoneID)
	p.WriteUint32(exp)
	s.SendAsync(p)
}

func (s *Session) HandleZoneExperience(zoneID uint32) {
	if s.ZoneExplored(zoneID) {
		fmt.Println("Zone", zoneID, "explored already")
		return
	}

	var areaTable *dbc.Ent_AreaTable
	s.DB().GetData(zoneID, &areaTable)

	yo.Spew(areaTable)

	if areaTable == nil {
		return
	}

	if s.Build().AddedIn(vsn.V1_12_1) {
		if areaTable.ExploreFlag != 0 {
			s.SetExplorationFlag(areaTable.ExploreFlag)
			s.UpdateSelf()
			s.DB().Insert(wdb.ExploredZone{
				Player: s.PlayerID(),
				ZoneID: zoneID,
			})
		}
	}

	if areaTable.ExplorationLevel != 0 {
		if s.GetLevel() >= s.GetMaxLevel() {
			s.SendExplorationXP(zoneID, 0)
		} else {
			var exp uint32
			diff := s.GetLevel() - int(areaTable.ExplorationLevel)
			if diff < -5 {
				exp = uint32(float32(s.WS.GetExploreXP(uint32(s.GetLevel()+5))) * s.WS.ExploreXPRate())
			} else if diff > 5 {
				explorationPercent := (100 - ((diff - 5) * 5))
				if explorationPercent > 100 {
					explorationPercent = 100
				}

				if explorationPercent < 0 {
					explorationPercent = 0
				}

				exp = uint32(
					float32(s.WS.GetExploreXP(uint32(areaTable.ExplorationLevel))) * (float32(explorationPercent) / 100.0) * s.WS.ExploreXPRate())
			} else {
				exp = uint32(float32(s.WS.GetExploreXP(uint32(areaTable.ExplorationLevel))) * s.WS.ExploreXPRate())
			}

			s.SendExplorationXP(zoneID, exp)
			s.AddExperience(exp)
		}
	} else {
		if areaTable.ExploreFlag != 0 {
			s.SendExplorationXP(zoneID, 0)
		}
	}
}
