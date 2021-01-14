//Package terrain implements decoders for the WDT and ADT terrain formats.
package terrain

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"

	"github.com/superp00t/gophercraft/format/chunked"
	"github.com/superp00t/gophercraft/vsn"
)

const TileSize = 533.33333
const ChunkSize = TileSize / 16

var (
	MapVersion         = chunked.CnkID("MVER")
	MainTileIndex      = chunked.CnkID("MAIN")
	MapIndexHeader     = chunked.CnkID("MPHD")
	MapWMO             = chunked.CnkID("MWMO")
	MapChunkInfo       = chunked.CnkID("MCIN")
	MapTextures        = chunked.CnkID("MTEX")
	TileHeader         = chunked.CnkID("MHDR")
	M2Models           = chunked.CnkID("MMDX")
	M2FilenameOffsets  = chunked.CnkID("MMID")
	WMOFilenameOffsets = chunked.CnkID("MWID")
	DoodadDefs         = chunked.CnkID("MDDF")
	WMODefs            = chunked.CnkID("MODF")
	Chunk              = chunked.CnkID("MCNK")
	Normals            = chunked.CnkID("MCNR")
	Heights            = chunked.CnkID("MCVT")
	Layer              = chunked.CnkID("MCLY")
	CollisionObjects   = chunked.CnkID("MCRF")
	ShadowMap          = chunked.CnkID("MCSH")
	Alpha              = chunked.CnkID("MCAL")
	Liquids            = chunked.CnkID("MCLQ")
	SoundEmitters      = chunked.CnkID("MCSE")
)

type HeaderFlags uint32

const (
	WDTUsesGlobalWMO HeaderFlags = 1 << iota
	ADTHasMCCV
	ADTHasBigAlpha
)

type TileFlags uint64

const (
	TileHasTerrain TileFlags = 1 << iota
)

type Index struct {
	// MVER
	Version uint32

	// MPHD
	HeaderFlags HeaderFlags

	// MAIN
	Tiles [64 * 64]TileIndex
}

type TileIndex struct {
	// Exists bool
	// Flags  uint32
	// Unk3   uint32
	Flags TileFlags
}

type MapReader struct {
	Name string
	Source
	Index
}

func NewMapReader(src Source, build vsn.Build, name string) (*MapReader, error) {
	mr := new(MapReader)
	mr.Name = name
	mr.Source = src

	indexPath := fmt.Sprintf("World/Maps/%s/%s.wdt", name, name)

	if mr.Source.Exists(indexPath) == false {
		return nil, fmt.Errorf("terrain: path %s does not exist in the source.", indexPath)
	}

	indexFile, err := mr.Source.ReadFile(indexPath)
	if err != nil {
		return nil, fmt.Errorf("terrain: error loading '%s', %s", indexPath, err)
	}

	defer indexFile.Close()

	indexReader := &chunked.Reader{Reader: indexFile}
	for {
		chunkID, chunk, err := indexReader.ReadChunk()
		if err == io.EOF {
			break
		}

		if err != nil {
			return nil, err
		}

		switch chunkID {
		case MapVersion:
			mr.Index.Version = binary.LittleEndian.Uint32(chunk)
		case MapIndexHeader:
			mr.Index.HeaderFlags = HeaderFlags(binary.LittleEndian.Uint32(chunk[:4]))
		case MainTileIndex:
			sizeof := 8

			for j := 0; j < 64; j++ {
				for i := 0; i < 64; i++ {
					offset := (j * 64 * sizeof) + (i * sizeof)

					tile := TileIndex{}
					tile.Flags = TileFlags(binary.LittleEndian.Uint64(chunk[offset : offset+8]))
					// path := fmt.Sprintf("World/Maps/%s/%s_%d_%d.adt", name, name, i, j)
					// tile.Exists = mr.Source.Exists(path)

					// if tile.Flags&TileHasTerrain != 0 {
					// 	if !tile.Exists {
					// 		tile.Exists = false
					// 		return nil, fmt.Errorf("has terrain flag set but does not exist: %s", path)
					// 	}
					// } else {
					// 	if tile.Exists {
					// 		return nil, fmt.Errorf("has no terrain flag but path exists: %s", path)
					// 	}
					// }

					mr.Index.Tiles[j*64+i] = tile
				}
			}
		case MapWMO:
			// nothing to do I think
		default:
			return nil, fmt.Errorf("unhandled chunk ID %s", chunkID)
		}
	}

	return mr, nil
}

type TileChunkLookupIndex struct {
	TileIndexX, TileIndexY   int
	ChunkIndexX, ChunkIndexY int
}

// Since maps are split into a hierarchy of tiles and chunks, we have to find their path from global coordinates.
func FindTileChunkIndex(x, y float32) (*TileChunkLookupIndex, error) {
	calcTile := func(axis float32) int {
		return int(math.Floor(32 - (float64(axis) / TileSize)))
	}

	calcChunk := func(axis float32) int {
		absZero := float64(-17066)
		absAxis := (float64(axis) - absZero)

		chunkRelTile := math.Mod(absAxis, TileSize)

		return int(((chunkRelTile) / ChunkSize))
	}

	tci := new(TileChunkLookupIndex)
	tci.TileIndexX = calcTile(x)
	tci.TileIndexY = calcTile(y)

	tci.ChunkIndexX = 15 - calcChunk(x)
	tci.ChunkIndexY = 15 - calcChunk(y)

	return tci, nil
}

func (mr *MapReader) GetChunkByPos(x, y float32) (*ChunkData, error) {
	idx, err := FindTileChunkIndex(x, y)
	if err != nil {
		return nil, err
	}

	return mr.GetChunkByIndex(idx)
}

func (mr *MapReader) GetChunkByIndex(idx *TileChunkLookupIndex) (*ChunkData, error) {
	tile, err := mr.ReadTile(idx.TileIndexY, idx.TileIndexX)
	if err != nil {
		return nil, err
	}

	if idx.ChunkIndexX >= 16 || idx.ChunkIndexX < 0 {
		return nil, fmt.Errorf("terrain: calculated chunk X position is out of range %d", idx.ChunkIndexX)
	}

	if idx.ChunkIndexY >= 16 || idx.ChunkIndexY < 0 {
		return nil, fmt.Errorf("terrain: calculated chunk Y position is out of range %d", idx.ChunkIndexY)
	}

	cd := tile.ChunkData[idx.ChunkIndexX][idx.ChunkIndexY]
	if cd == nil {
		return nil, fmt.Errorf("terrain: no chunk found at %d:%d", idx.ChunkIndexY, idx.ChunkIndexX)
	}
	return cd, nil
}
