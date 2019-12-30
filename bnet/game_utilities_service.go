package bnet

import (
	"compress/zlib"
	"fmt"
	"strings"

	"github.com/superp00t/etc"
	"github.com/superp00t/etc/yo"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/superp00t/gophercraft/bnet/bgs/protocol"
	v1 "github.com/superp00t/gophercraft/bnet/bgs/protocol/game_utilities/v1"
	"github.com/superp00t/gophercraft/bnet/realmlist"
)

const (
	REALM_TYPE_NORMAL     = 0
	REALM_TYPE_PVP        = 1
	REALM_TYPE_NORMAL2    = 4
	REALM_TYPE_RP         = 6
	REALM_TYPE_RPPVP      = 8
	MAX_CLIENT_REALM_TYPE = 14
	REALM_TYPE_FFA_PVP    = 16

	REALM_FLAG_NONE             = 0x00
	REALM_FLAG_VERSION_MISMATCH = 0x01
	REALM_FLAG_OFFLINE          = 0x02
	REALM_FLAG_SPECIFYBUILD     = 0x04
	REALM_FLAG_UNK1             = 0x08
	REALM_FLAG_UNK2             = 0x10
	REALM_FLAG_RECOMMENDED      = 0x20
	REALM_FLAG_NEW              = 0x40
	REALM_FLAG_FULL             = 0x80
)

var (
	RealmTypes = [MAX_CLIENT_REALM_TYPE]uint32{
		1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14,
	}
)

type RealmHandle struct {
	Region uint8
	Site   uint8
	ID     uint32
}

func (rh RealmHandle) String() string {
	return fmt.Sprintf("%d-%d-%d", rh.Region, rh.Site, rh.ID)
}

func (rh RealmHandle) GetAddress() uint32 {
	return (uint32(rh.Region) << 24) | (uint32(rh.Site) << 16) | uint32(uint16(rh.ID))
}

func (l *Listener) ProcessClientRequest(conn *Conn, token uint32, args *v1.ClientRequest) {
	var command *protocol.Attribute

	params := make(map[string]*protocol.Variant)

	for _, v := range args.GetAttribute() {
		params[v.GetName()] = v.GetValue()
		if strings.HasPrefix(v.GetName(), "Command_") {
			command = v
		}
	}

	if command == nil {
		conn.SendResponseCode(token, ERROR_RPC_MALFORMED_REQUEST)
		return
	}

	yo.Ok("ClientRequest::" + command.GetName() + "()")

	switch command.GetName() {
	case "Command_RealmListTicketRequest_v1_b9":
		l.HandleRealmListTicketRequest(conn, token, params)
	case "Command_LastCharPlayedRequest_v1_b9":
		// NYI
		conn.SendResponseOK(token, &v1.ClientResponse{})
	case "Command_RealmListRequest_v1_b9":
		l.HandleRealmListRequest(conn, token, params)
	default:
		yo.Fatal(command.GetName())
	}
}

func getParamsJSON(params map[string]*protocol.Variant, key string) string {
	if params[key] == nil {
		return ""
	}

	sdat := strings.SplitN(string(params[key].GetBlobValue()), ":", 2)
	if len(sdat) != 2 {
		return ""
	}

	return sdat[1]
}

func decodeParamsJSON(params map[string]*protocol.Variant, key string, v proto.Message) error {
	data := getParamsJSON(params, key)
	err := jsonpb.Unmarshal(strings.NewReader(data), v)
	return err
}

func (l *Listener) HandleRealmListTicketRequest(conn *Conn, token uint32, params map[string]*protocol.Variant) {
	var rltr realmlist.RealmListTicketClientInformation

	err := decodeParamsJSON(params, "Param_ClientInfo", &rltr)
	if err != nil {
		conn.SendResponseCode(token, ERROR_RPC_MALFORMED_REQUEST)
		return
	}

	conn.versionInfo = rltr.GetInfo().GetVersion()
	conn.version = rltr.GetInfo().GetVersion().GetVersionBuild()
	copy(conn.clientSecret[:], rltr.GetInfo().GetSecret())

	resp := &v1.ClientResponse{
		Attribute: []*protocol.Attribute{
			{
				Name: strp("Param_RealmListTicket"),
				Value: &protocol.Variant{
					BlobValue: []byte("AuthRealmListTicket"),
				},
			},
		},
	}

	conn.SendResponse(token, resp)
	// v := params[""]
}

func (l *Listener) HandleRealmListRequest(conn *Conn, token uint32, params map[string]*protocol.Variant) {
	var updates realmlist.RealmListUpdates
	updates.Updates = []*realmlist.RealmState{
		{
			Update: &realmlist.RealmEntry{
				WowRealmAddress: u32p(RealmHandle{1, 1, 1}.GetAddress()),
				Version:         conn.versionInfo,
				PopulationState: u32p(1),

				CfgRealmsID:     u32p(1),
				Flags:           u32p(REALM_FLAG_NEW | REALM_FLAG_RECOMMENDED | REALM_FLAG_SPECIFYBUILD),
				Name:            strp("Bnet Test"),
				CfgConfigsID:    u32p(RealmTypes[REALM_TYPE_PVP]),
				CfgLanguagesID:  u32p(1),
				CfgTimezonesID:  u32p(1),
				CfgCategoriesID: u32p(1),
			},

			Deleting: boolp(false),
		},
	}

	counts := &realmlist.RealmCharacterCountList{
		Counts: []*realmlist.RealmCharacterCountEntry{
			{
				WowRealmAddress: u32p(RealmHandle{1, 1, 1}.GetAddress()),
				Count:           u32p(0),
			},
		},
	}

	resp := &v1.ClientResponse{}
	resp.Attribute = []*protocol.Attribute{}
	appendCompressedJSON(&resp.Attribute, "Param_RealmList", "JSONRealmListUpdates", &updates)
	appendCompressedJSON(&resp.Attribute, "Param_CharacterCountList", "JSONRealmCharacterCountList", counts)

	conn.SendResponseOK(token, resp)
}

func appendCompressedJSON(attr *[]*protocol.Attribute, key, name string, value proto.Message) {
	js := etc.NewBuffer()
	if err := marshal().Marshal(js, value); err != nil {
		panic(err)
	}

	data := name + ":" + js.ToString()

	out := etc.NewBuffer()
	// little endian
	out.WriteUint32(uint32(len(data) + 1))

	// compress JSON data
	z := zlib.NewWriter(out)
	z.Write(append([]byte(data), 0))
	z.Close()

	variant := &protocol.Variant{
		BlobValue: out.Bytes(),
	}

	*attr = append(*attr, &protocol.Attribute{
		Name:  strp(key),
		Value: variant,
	})
}

func (l *Listener) PresenceChannelCreated(conn *Conn, token uint32, args *v1.PresenceChannelCreatedRequest) {
	conn.SendResponseCode(token, ERROR_RPC_NOT_IMPLEMENTED)
}
func (l *Listener) GetPlayerVariables(conn *Conn, token uint32, args *v1.GetPlayerVariablesRequest) {
	conn.SendResponseCode(token, ERROR_RPC_NOT_IMPLEMENTED)
}
func (l *Listener) ProcessServerRequest(conn *Conn, token uint32, args *v1.ServerRequest) {
	conn.SendResponseCode(token, ERROR_RPC_NOT_IMPLEMENTED)
}
func (l *Listener) OnGameAccountOnline(conn *Conn, token uint32, args *v1.GameAccountOnlineNotification) {
	conn.SendResponseCode(token, ERROR_RPC_NOT_IMPLEMENTED)
}
func (l *Listener) OnGameAccountOffline(conn *Conn, token uint32, args *v1.GameAccountOfflineNotification) {
	conn.SendResponseCode(token, ERROR_RPC_NOT_IMPLEMENTED)
}
func (l *Listener) GetAchievementsFile(conn *Conn, token uint32, args *v1.GetAchievementsFileRequest) {
	conn.SendResponseCode(token, ERROR_RPC_NOT_IMPLEMENTED)
}
func (l *Listener) GetAllValuesForAttribute(conn *Conn, token uint32, args *v1.GetAllValuesForAttributeRequest) {
	if args.GetAttributeKey() == "Command_RealmListRequest_v1_b9" {
		resp := &v1.GetAllValuesForAttributeResponse{}
		resp.AttributeValue = append(resp.AttributeValue, &protocol.Variant{
			StringValue: strp(RealmHandle{1, 1, 0}.String()),
		})
		conn.SendResponse(token, resp)
		return
	}

	yo.Puke(args)

	conn.SendResponseCode(token, ERROR_RPC_NOT_IMPLEMENTED)
}
func (l *Listener) RegisterUtilities(conn *Conn, token uint32, args *v1.RegisterUtilitiesRequest) {
	conn.SendResponseCode(token, ERROR_RPC_NOT_IMPLEMENTED)
}
func (l *Listener) UnregisterUtilities(conn *Conn, token uint32, args *v1.UnregisterUtilitiesRequest) {
	conn.SendResponseCode(token, ERROR_RPC_NOT_IMPLEMENTED)
}
