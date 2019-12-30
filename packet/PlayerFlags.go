package packet

import "strings"

//go:generate gcraft_stringer -type=PlayerFlags -method=toString
type PlayerFlags uint32

const (
	PLAYER_FLAGS_NONE              PlayerFlags = 0x00000000
	PLAYER_FLAGS_GROUP_LEADER      PlayerFlags = 0x00000001
	PLAYER_FLAGS_AFK               PlayerFlags = 0x00000002
	PLAYER_FLAGS_DND               PlayerFlags = 0x00000004
	PLAYER_FLAGS_GM                PlayerFlags = 0x00000008
	PLAYER_FLAGS_GHOST             PlayerFlags = 0x00000010
	PLAYER_FLAGS_RESTING           PlayerFlags = 0x00000020
	PLAYER_FLAGS_UNK7              PlayerFlags = 0x00000040 // admin?
	PLAYER_FLAGS_FFA_PVP           PlayerFlags = 0x00000080
	PLAYER_FLAGS_CONTESTED_PVP     PlayerFlags = 0x00000100 // Player has been involved in a PvP combat and will be attacked by contested guards
	PLAYER_FLAGS_PVP_DESIRED       PlayerFlags = 0x00000200 // Stores player's permanent PvP flag preference
	PLAYER_FLAGS_HIDE_HELM         PlayerFlags = 0x00000400
	PLAYER_FLAGS_HIDE_CLOAK        PlayerFlags = 0x00000800
	PLAYER_FLAGS_PARTIAL_PLAY_TIME PlayerFlags = 0x00001000 // played long time
	PLAYER_FLAGS_NO_PLAY_TIME      PlayerFlags = 0x00002000 // played too long time
	PLAYER_FLAGS_UNK15             PlayerFlags = 0x00004000
	PLAYER_FLAGS_UNK16             PlayerFlags = 0x00008000 // strange visual effect (2.0.1) looks like PLAYER_FLAGS_GHOST flag
	PLAYER_FLAGS_SANCTUARY         PlayerFlags = 0x00010000 // player entered sanctuary
	PLAYER_FLAGS_TAXI_BENCHMARK    PlayerFlags = 0x00020000 // taxi benchmark mode (on/off) (2.0.1)
	PLAYER_FLAGS_PVP_TIMER         PlayerFlags = 0x00040000 // 3.0.2 pvp timer active (after you disable pvp manually)
)

func (p PlayerFlags) String() string {
	var s []string

	for x := PlayerFlags(1); x <= PLAYER_FLAGS_PVP_TIMER; x++ {
		if !strings.HasPrefix(x.toString(), "PlayerFlags(") {
			if p&x != 0 {
				s = append(s, x.toString())
			}
		}
	}

	return strings.Join(s, " | ")
}
