package bnet

import v1 "github.com/superp00t/gophercraft/bnet/bgs/protocol/account/v1"

func (s *Listener) ResolveAccount(conn *Conn, token uint32, args *v1.ResolveAccountRequest) {
	conn.SendResponseCode(token, ERROR_RPC_NOT_IMPLEMENTED)
}

// func (s *Listener) IsIgrAddress(conn *Conn, token uint32, args *v1.IsIgrAddressRequest) {
// 	conn.SendResponseCode(token, ERROR_RPC_NOT_IMPLEMENTED)
// }
func (s *Listener) Subscribe(conn *Conn, token uint32, args *v1.SubscriptionUpdateRequest) {
	conn.SendResponseCode(token, ERROR_RPC_NOT_IMPLEMENTED)
}
func (s *Listener) Unsubscribe(conn *Conn, token uint32, args *v1.SubscriptionUpdateRequest) {
	conn.SendResponseCode(token, ERROR_RPC_NOT_IMPLEMENTED)
}
func (s *Listener) GetAccountState(conn *Conn, token uint32, args *v1.GetAccountStateRequest) {
	resp := &v1.GetAccountStateResponse{
		State: &v1.AccountState{
			PrivacyInfo: &v1.PrivacyInfo{
				IsUsingRid:               boolp(false),
				IsVisibleForViewFriends:  boolp(false),
				IsHiddenFromFriendFinder: boolp(true),
			},
		},

		Tags: &v1.AccountFieldTags{
			PrivacyInfoTag: u32p(0xD7CA834D),
		},
	}

	conn.SendResponse(token, resp)
}

func (s *Listener) GetGameAccountState(conn *Conn, token uint32, args *v1.GetGameAccountStateRequest) {
	resp := &v1.GetGameAccountStateResponse{}
	resp.State = &v1.GameAccountState{}
	resp.Tags = &v1.GameAccountFieldTags{}

	if args.GetOptions().GetFieldGameLevelInfo() {
		resp.State.GameLevelInfo = &v1.GameLevelInfo{
			Name:    strp(conn.user),
			Program: u32p(5730135),
		}

		resp.Tags.GameLevelInfoTag = u32p(0x5C46D483)
	}

	if args.GetOptions().GetFieldGameStatus() {
		resp.State.GameStatus = &v1.GameStatus{
			IsSuspended: boolp(false),
			IsBanned:    boolp(false),
			Program:     u32p(5730135),
		}

		resp.Tags.GameStatusTag = u32p(0x98B75F99)
	}

	conn.SendResponse(token, resp)
}

func (s *Listener) GetLicenses(conn *Conn, token uint32, args *v1.GetLicensesRequest) {
	conn.SendResponseCode(token, ERROR_RPC_NOT_IMPLEMENTED)
}
func (s *Listener) GetGameTimeRemainingInfo(conn *Conn, token uint32, args *v1.GetGameTimeRemainingInfoRequest) {
	conn.SendResponseCode(token, ERROR_RPC_NOT_IMPLEMENTED)
}
func (s *Listener) GetGameSessionInfo(conn *Conn, token uint32, args *v1.GetGameSessionInfoRequest) {
	conn.SendResponseCode(token, ERROR_RPC_NOT_IMPLEMENTED)
}
func (s *Listener) GetCAISInfo(conn *Conn, token uint32, args *v1.GetCAISInfoRequest) {
	conn.SendResponseCode(token, ERROR_RPC_NOT_IMPLEMENTED)
}
func (s *Listener) GetAuthorizedData(conn *Conn, token uint32, args *v1.GetAuthorizedDataRequest) {
	conn.SendResponseCode(token, ERROR_RPC_NOT_IMPLEMENTED)
}
func (s *Listener) GetSignedAccountState(conn *Conn, token uint32, args *v1.GetSignedAccountStateRequest) {
	conn.SendResponseCode(token, ERROR_RPC_NOT_IMPLEMENTED)
}
