package realm

import (
	"fmt"
	"sync"
	"time"

	"github.com/superp00t/etc/yo"
	"github.com/superp00t/gophercraft/format/dbc"
	"github.com/superp00t/gophercraft/format/terrain"
	"github.com/superp00t/gophercraft/realm/wdb"
)

type TerrainMgr struct {
	Chunks *sync.Map
}

type MapChunkKey struct {
	MapID uint32
	Index terrain.TileChunkLookupIndex
}

type MapChunkValue struct {
	Loaded    time.Time
	ChunkData *terrain.ChunkData
}

func (ws *Server) LookupMapChunk(key MapChunkKey) (*terrain.ChunkData, error) {
	mapChunkValue, ok := ws.TerrainMgr.Chunks.Load(key)
	if !ok {
		// Load Map data from pack
		var cmap *wdb.Map
		ws.DB.GetData(key.MapID, &cmap)
		if cmap == nil {
			return nil, fmt.Errorf("no map for %d", key.MapID)
		}

		mr, err := terrain.NewMapReader(ws.PackLoader, ws.Build(), cmap.GetDirectory())
		if err != nil {
			return nil, err
		}

		cnkData, err := mr.GetChunkByIndex(&key.Index)
		if err != nil {
			return nil, err
		}

		ws.TerrainMgr.Chunks.Store(key, &MapChunkValue{
			Loaded:    time.Now(),
			ChunkData: cnkData,
		})
		return cnkData, nil
	}
	mcv := mapChunkValue.(*MapChunkValue)
	mcv.Loaded = time.Now()
	return mcv.ChunkData, nil
}

func (s *Session) UpdateArea() {
	if s.State != InWorld {
		return
	}

	// We need to see which chunk
	pos := s.Position()

	tcli, err := terrain.FindTileChunkIndex(pos.X, pos.Y)
	if err != nil {
		// Todo: Send player somewhere if they're out of bounds
		yo.Warn(err)
		return
	}

	if s.CurrentChunkIndex != nil {
		// Player is in same chunk. Nothing to update
		if *s.CurrentChunkIndex == *tcli {
			return
		}
	}

	s.CurrentChunkIndex = tcli

	lookup := MapChunkKey{
		MapID: s.CurrentMap,
		Index: *tcli,
	}

	cnk, err := s.WS.LookupMapChunk(lookup)
	if err != nil {
		yo.Warn(err)
		return
	}

	if cnk.AreaID != s.CurrentArea {
		s.CurrentArea = cnk.AreaID

		var area *dbc.Ent_AreaTable
		s.DB().GetData(s.CurrentArea, &area)

		if area == nil {
			return
		}

		s.HandleZoneExperience(s.CurrentArea)

		for area.ParentArea != 0 {
			s.DB().GetData(area.ParentArea, &area)
			if area == nil {
				break
			}

			s.HandleZoneExperience(area.ParentArea)
		}
	}

	s.CurrentArea = cnk.AreaID
}

const terrainMgrSweepInterval = 2 * time.Minute
const unusedChunkLifetime = 2 * time.Minute

func (ws *Server) InitTerrainMgr() {
	ws.TerrainMgr.Chunks = new(sync.Map)

	for {
		time.Sleep(terrainMgrSweepInterval)
		ws.TerrainMgr.Chunks.Range(func(k, v interface{}) bool {
			key := k.(MapChunkKey)
			value := v.(*MapChunkValue)
			if time.Since(value.Loaded) > unusedChunkLifetime {
				ws.TerrainMgr.Chunks.Delete(key)
			}
			return true
		})
	}
}
