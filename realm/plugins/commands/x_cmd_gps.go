package commands

import (
	"github.com/superp00t/gophercraft/format/dbc"
	"github.com/superp00t/gophercraft/realm"
)

func cmdGPS(s *realm.Session) {
	pos := s.Position()

	var zoneName, areaName string
	var areaEntry *dbc.Ent_AreaTable
	// Query main zone
	s.DB().GetData(s.ZoneID, &areaEntry)
	if areaEntry != nil {
		zoneName = areaEntry.Name
	}
	areaEntry = nil
	// Query specific zone/subzone
	s.DB().GetData(s.CurrentArea, &areaEntry)
	if areaEntry != nil {
		areaName = areaEntry.Name
	}

	s.Warnf("Position:")
	s.Warnf("World State %d", s.State)
	s.Warnf(" X: %f", pos.X)
	s.Warnf(" Y: %f", pos.Y)
	s.Warnf(" Z: %f", pos.Z)
	s.Warnf(" Facing: %f", pos.O)
	s.Warnf("Map: %d", s.CurrentMap)
	s.Warnf("Zone: %d '%s'", s.ZoneID, zoneName)
	s.Warnf("Area: %d '%s'", s.CurrentArea, areaName)
	s.Warnf("Phase: %s", s.CurrentPhase)

	if idx := s.CurrentChunkIndex; idx != nil {
		s.Warnf("Tile: %d:%d", idx.TileIndexX, idx.TileIndexY)
		s.Warnf("Chunk: %d:%d", idx.ChunkIndexX, idx.ChunkIndexY)
	}
}
