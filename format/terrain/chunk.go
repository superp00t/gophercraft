package terrain

const ChunkBufferSize = 9*9 + 8*8

type ChunkHeader struct {
	Flags  uint32
	IndexX uint32
	IndexY uint32
	// Radius            float32
	NumLayers         uint32
	NumDoodadRefs     uint32
	OffsetHeight      uint32
	OffsetNormal      uint32
	OffsetLayer       uint32
	OffsetRefs        uint32
	OffsetAlpha       uint32
	SizeAlpha         uint32
	OffsetShadow      uint32
	SizeShadow        uint32
	AreaID            uint32
	NumMapObjRefs     uint32
	Holes             uint16
	Pad0              uint16
	PredTex           [8]uint16
	NumEffectDoodad   [8]byte
	OffsetSndEmitters uint32
	NumSndEmitters    uint32
	OffsetLiquid      uint32
	SizeLiquid        uint32
	Position          C3Vector
	OffsetMCCV        uint32
	Unused1           uint32
	Unused2           uint32
}

type ChunkLayer struct {
	TextureID   uint32
	Flags       uint32 // only use_alpha_map is implemented
	OffsetAlpha uint32
	EffectID    uint16
	Pad         uint16
}

type ChunkAlphaMap [64 * 64]byte

type ChunkLiquids struct {
	MinHeight float32
	MaxHeight float32

	Verts [9 * 9]struct {
		Data      [4]byte
		FloatData float32
	}

	Tiles         [8 * 8]byte
	NumFlowValues uint32

	FlowValues [2]struct {
		Sphere    CAaSphere
		Direction C3Vector
		Velocity  float32
		Amplitude float32
		Frequency float32
	}
}

type ChunkData struct {
	ChunkHeader

	// MCNR
	Normals    [ChunkBufferSize][3]byte
	NormalsPad [13]byte

	// MCVT
	Heights [ChunkBufferSize]float32

	// MCLY
	Layer []ChunkLayer

	// MCRF
	CollisionDoodads []uint32
	CollisionWMOs    []uint32

	// MCSH
	ShadowMap [64]uint64

	// MCAL
	AlphaMaps []ChunkAlphaMap

	// MCLQ
	Liquids ChunkLiquids

	// MCSE
	OldSoundEmitters []ChunkOldSoundEmitter
}

type ChunkOldSoundEmitter struct {
	/*000h*/ SoundPointID uint32
	/*004h*/ SoundNameID uint32
	/*008h*/ Position C3Vector
	/*014h*/ MinDistance float32
	/*018h*/ MaxDistance float32
	/*01Ch*/ CutoffDistance float32
	/*020h*/ StartTime uint16
	/*022h*/ EndTime uint16
	/*024h*/ Mode uint16
	/*026h*/ GroupSilenceMin uint16
	/*028h*/ GroupSilenceMax uint16
	/*02Ah*/ PlayInstancesMin uint16
	/*02Ch*/ PlayInstancesMax uint16
	/*02Eh*/ LoopCountMin byte
	/*02Fh*/ LoopCountMax byte
	/*030h*/ InterSoundGapMin uint16
	/*032h*/ InterSoundGapMax uint16
}
