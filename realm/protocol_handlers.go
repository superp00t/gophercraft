package realm

import (
	p "github.com/superp00t/gophercraft/packet"
)

type Handlers struct {
	Map map[p.WorldType]*WorldClientHandler
}

func (ws *Server) initHandlers() {
	h := &Handlers{make(map[p.WorldType]*WorldClientHandler)}

	// Connection
	h.On(p.CMSG_PET_NAME_QUERY, 0, (*Session).SendPet)
	h.On(p.CMSG_WARDEN_DATA, 0, (*Session).WardenResponse)
	h.On(p.CMSG_PING, 0, (*Session).HandlePong)
	h.On(p.CMSG_UPDATE_ACCOUNT_DATA, 0, (*Session).HandleAccountDataUpdate)
	h.On(p.CMSG_REALM_SPLIT, 0, (*Session).HandleRealmSplit)
	h.On(p.CMSG_UI_TIME_REQUEST, 0, (*Session).HandleUITimeRequest)

	// Character selection menu
	h.On(p.CMSG_CHAR_ENUM, 0, (*Session).HandleRequestCharacterList)
	h.On(p.CMSG_CHAR_DELETE, 0, (*Session).DeleteCharacter)
	h.On(p.CMSG_CHAR_CREATE, 0, (*Session).CreateCharacter)
	h.On(p.CMSG_PLAYER_LOGIN, 0, (*Session).HandleJoin, OptionAsync)
	h.On(p.CMSG_LOGOUT_REQUEST, 1, (*Session).HandleLogoutRequest)

	// Social
	h.On(p.CMSG_NAME_QUERY, 1, (*Session).HandleNameQuery)
	h.On(p.CMSG_MESSAGECHAT, 1, (*Session).HandleChat)
	h.On(p.CMSG_WHO, 1, (*Session).HandleWho)
	h.On(p.CMSG_SET_SELECTION, 1, (*Session).HandleTarget)
	// Vanilla: Request friend list and ignore list as individual packets.
	h.On(p.CMSG_FRIEND_LIST, 1, (*Session).HandleFriendListRequest)
	// 2.4.3+: All social data is merged into a single list.
	h.On(p.CMSG_CONTACT_LIST, 1, (*Session).HandleSocialListRequest)
	h.On(p.CMSG_ADD_FRIEND, 1, (*Session).HandleFriendAdd)
	h.On(p.CMSG_DEL_FRIEND, 1, (*Session).HandleFriendDelete)
	h.On(p.CMSG_ADD_IGNORE, 1, (*Session).HandleIgnoreAdd)
	h.On(p.CMSG_DEL_IGNORE, 1, (*Session).HandleIgnoreDelete)

	// Party/Group
	h.On(p.CMSG_GROUP_INVITE, 1, (*Session).HandleGroupInvite)
	h.On(p.CMSG_GROUP_ACCEPT, 1, (*Session).HandleGroupAccept)
	h.On(p.CMSG_GROUP_DECLINE, 1, (*Session).HandleGroupDecline)
	h.On(p.CMSG_GROUP_DISBAND, 1, (*Session).HandleGroupDisband)
	h.On(p.CMSG_REQUEST_PARTY_MEMBER_STATS, 1, (*Session).HandleRequestPartyMemberStats)

	// Movement
	h.On(p.MSG_MOVE_HEARTBEAT, 1, (*Session).HandleMoves)
	h.On(p.MSG_MOVE_START_FORWARD, 1, (*Session).HandleMoves)
	h.On(p.MSG_MOVE_START_BACKWARD, 1, (*Session).HandleMoves)
	h.On(p.MSG_MOVE_STOP, 1, (*Session).HandleMoves)
	h.On(p.MSG_MOVE_START_STRAFE_LEFT, 1, (*Session).HandleMoves)
	h.On(p.MSG_MOVE_START_STRAFE_RIGHT, 1, (*Session).HandleMoves)
	h.On(p.MSG_MOVE_STOP_STRAFE, 1, (*Session).HandleMoves)
	h.On(p.MSG_MOVE_JUMP, 1, (*Session).HandleMoves)
	h.On(p.MSG_MOVE_START_TURN_LEFT, 1, (*Session).HandleMoves)
	h.On(p.MSG_MOVE_START_TURN_RIGHT, 1, (*Session).HandleMoves)
	h.On(p.MSG_MOVE_STOP_TURN, 1, (*Session).HandleMoves)
	h.On(p.MSG_MOVE_START_PITCH_UP, 1, (*Session).HandleMoves)
	h.On(p.MSG_MOVE_START_PITCH_DOWN, 1, (*Session).HandleMoves)
	h.On(p.MSG_MOVE_SET_PITCH, 1, (*Session).HandleMoves)
	h.On(p.MSG_MOVE_STOP_PITCH, 1, (*Session).HandleMoves)
	h.On(p.MSG_MOVE_SET_RUN_MODE, 1, (*Session).HandleMoves)
	h.On(p.MSG_MOVE_SET_FACING, 1, (*Session).HandleMoves)
	h.On(p.MSG_MOVE_FALL_LAND, 1, (*Session).HandleMoves)
	h.On(p.MSG_MOVE_START_SWIM, 1, (*Session).HandleMoves)
	h.On(p.MSG_MOVE_STOP_SWIM, 1, (*Session).HandleMoves)
	h.On(p.MSG_MOVE_WORLDPORT_ACK, 1, (*Session).HandleWorldportAck)
	h.On(p.CMSG_SUMMON_RESPONSE, 1, (*Session).HandleSummonResponse)

	// Location
	h.On(p.CMSG_ZONEUPDATE, 1, (*Session).HandleZoneUpdate)
	h.On(p.CMSG_AREATRIGGER, 1, (*Session).HandleAreaTrigger)
	h.On(p.CMSG_WORLD_TELEPORT, 1, (*Session).HandleWorldTeleport)

	// Animation
	h.On(p.CMSG_STANDSTATECHANGE, 1, (*Session).HandleStandStateChange)
	h.On(p.CMSG_TEXT_EMOTE, 1, (*Session).HandleTextEmote)
	h.On(p.CMSG_SETSHEATHED, 1, (*Session).HandleSheathe)
	// Alpha quirk
	h.On(p.CMSG_SETWEAPONMODE, 1, (*Session).HandleSetWeaponMode)

	// Gameobjects
	h.On(p.CMSG_GAMEOBJECT_QUERY, 1, (*Session).HandleGameObjectQuery)
	h.On(p.CMSG_GAMEOBJ_USE, 1, (*Session).HandleGameObjectUse)

	// Creatures
	h.On(p.CMSG_CREATURE_QUERY, 1, (*Session).HandleCreatureQuery)
	h.On(p.CMSG_QUESTGIVER_STATUS_QUERY, 1, (*Session).HandleQuestgiverStatusQuery)
	h.On(p.CMSG_GOSSIP_HELLO, 1, (*Session).HandleGossipHello)
	h.On(p.CMSG_GOSSIP_SELECT_OPTION, 1, (*Session).HandleGossipSelectOption)
	h.On(p.CMSG_NPC_TEXT_QUERY, 1, (*Session).HandleGossipTextQuery)

	// Inventory
	h.On(p.CMSG_ITEM_QUERY_SINGLE, 0, (*Session).HandleItemQuery)
	h.On(p.CMSG_SWAP_INV_ITEM, 1, (*Session).HandleSwapInventoryItem)
	h.On(p.CMSG_SWAP_ITEM, 1, (*Session).HandleSwapItem)
	h.On(p.CMSG_AUTOEQUIP_ITEM, 1, (*Session).HandleAutoEquipItem)
	h.On(p.CMSG_DESTROYITEM, 1, (*Session).HandleDestroyItem)
	h.On(p.CMSG_SPLIT_ITEM, 1, (*Session).HandleSplitItem)

	// Spells/Combat
	h.On(p.CMSG_SET_ACTION_BUTTON, 1, (*Session).HandleSetActionButton)

	ws.handlers = h
}

const (
	OptionFlagAsync = 1 << iota
)

type Options struct {
	OptionFlags uint32
}

type Option func(p.WorldType, []byte, *Options) error

func OptionAsync(t p.WorldType, data []byte, opts *Options) error {
	opts.OptionFlags |= OptionFlagAsync
	return nil
}

func (h *Handlers) On(pt p.WorldType, requiredState SessionState, function interface{}, options ...Option) {
	h.Map[pt] = &WorldClientHandler{pt, requiredState, function, options}
}

type WorldClientHandler struct {
	Op            p.WorldType
	RequiredState SessionState
	Fn            interface{}
	Options       []Option
}
