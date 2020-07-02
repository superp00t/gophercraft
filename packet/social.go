package packet

type FriendStatus uint8

const (
	FriendOffline FriendStatus = 0x00
	FriendOnline  FriendStatus = 0x01
	FriendAFK     FriendStatus = 0x02
	FriendDND     FriendStatus = 0x04
	FriendRAF     FriendStatus = 0x08
)

const (
	SocialFlagFriend  = 0x01
	SocialFlagIgnored = 0x02
	SocialFlagMuted   = 0x04
)

const (
	FRIEND_DB_ERROR         = 0x00
	FRIEND_LIST_FULL        = 0x01
	FRIEND_ONLINE           = 0x02
	FRIEND_OFFLINE          = 0x03
	FRIEND_NOT_FOUND        = 0x04
	FRIEND_REMOVED          = 0x05
	FRIEND_ADDED_ONLINE     = 0x06
	FRIEND_ADDED_OFFLINE    = 0x07
	FRIEND_ALREADY          = 0x08
	FRIEND_SELF             = 0x09
	FRIEND_ENEMY            = 0x0A
	FRIEND_IGNORE_FULL      = 0x0B
	FRIEND_IGNORE_SELF      = 0x0C
	FRIEND_IGNORE_NOT_FOUND = 0x0D
	FRIEND_IGNORE_ALREADY   = 0x0E
	FRIEND_IGNORE_ADDED     = 0x0F
	FRIEND_IGNORE_REMOVED   = 0x10
	FRIEND_IGNORE_AMBIGUOUS = 0x11 // That name is ambiguous, type more of the player's server name
	FRIEND_MUTE_FULL        = 0x12
	FRIEND_MUTE_SELF        = 0x13
	FRIEND_MUTE_NOT_FOUND   = 0x14
	FRIEND_MUTE_ALREADY     = 0x15
	FRIEND_MUTE_ADDED       = 0x16
	FRIEND_MUTE_REMOVED     = 0x17
	FRIEND_MUTE_AMBIGUOUS   = 0x18 // That name is ambiguous, type more of the player's server name
	FRIEND_UNK1             = 0x19 // no message at client
	FRIEND_UNK2             = 0x1A
	FRIEND_UNK3             = 0x1B
	FRIEND_UNKNOWN          = 0x1C // Unknown friend response from server
)
