package client

import (
	"fmt"
	"log"

	"github.com/superp00t/etc"

	"github.com/davecgh/go-spew/spew"
	"github.com/superp00t/etc/yo"
	"github.com/superp00t/gophercraft/packet"
)

type ClientHandler struct {
	Type packet.WorldType
	Fn   func([]byte)
}

func (cl *Client) On(t packet.WorldType, fn func([]byte)) {
	cl.Handlers[t] = &ClientHandler{
		Type: t,
		Fn:   fn,
	}
}

func (cl *Client) HandleCharList(b []byte) {
	data, err := packet.UnmarshalCharacterList(cl.Cfg.Build, b)
	if err != nil {
		log.Fatal(err)
	}

	yo.Warn(spew.Sdump(data))
	if len(b) > 100 {
		yo.Warn("Weird packet: ", b)
	}

	for _, v := range data.Characters {
		if v.Name == cl.Player {
			cl.PlayerGUID = v.GUID
			pkt := packet.NewGamePacket(packet.CMSG_PLAYER_LOGIN)
			pkt.WriteUint64(v.GUID.U64())
			cl.Send(pkt)
			return
		}
	}
}

func (cl *Client) HandleMOTD(b []byte) {
	yo.Println(spew.Sdump(b))
}

func (cl *Client) HandleLogin(b []byte) {
	d, _ := packet.UnmarshalSMSGAuthResponse(b)
	yo.Println(spew.Sdump(d))
	if d.Cmd == packet.AUTH_OK {
		pkt := packet.NewGamePacket(packet.CMSG_CHAR_ENUM)
		cl.Send(pkt)
	}
}

func pbt(fix string, input []byte) {
	stb := "Buf_" + fix + " := []byte{ "
	for _, v := range input {
		stb += fmt.Sprintf("0x%X, ", v)
	}
	stb += " }"
	yo.Println(stb)
}

func (cl *Client) HandleMessage(b []byte) {
	msg := packet.UnmarshalChatMessage(b)
	yo.Spew(msg)
}

func (cl *Client) HandleEquip(d []byte) {
	pbt("equip", d)
}

func (cl *Client) HandleActions(d []byte) {
	pbt("actionbuttons", d)
}

func (cl *Client) HandleReputations(d []byte) {
	pbt("reps", d)
}

func (cl *Client) HandleSocialList(d []byte) {
	pbt("sociallist", d)
}

func (cl *Client) HandleDanceMoves(d []byte) {
	pbt("dance", d)
}

func (cl *Client) HandleForcedReactions(d []byte) {
	pbt("forced", d)
}

func (cl *Client) HandleSpellList(d []byte) {
	pbt("spelllist", d)
}

func (cl *Client) HandleBindPointUpdate(d []byte) {
	in := etc.FromBytes(d)
	x := in.ReadFloat32()
	y := in.ReadFloat32()
	z := in.ReadFloat32()
	mapID := in.ReadUint32()
	zoneID := in.ReadUint32()

	yo.Ok(x)
	yo.Ok(y)
	yo.Ok(z)
	yo.Ok(mapID)
	yo.Ok(zoneID)

}

func (cl *Client) HandleCompressedUpdate(d []byte) {
	in := etc.FromBytes(d)
	length := in.ReadUint32()
	body := in.ReadRemainder()
	decompressedBody := packet.Uncompress(body)
	yo.Ok("Compressed: ", len(body))
	yo.Ok("Decompressed: ", len(decompressedBody))
	yo.Ok("Sized: ", length)
	yo.Ok("packet should contain", in.Len()-4)
	// yo.Fatal(hex.EncodeToString(d))
	cl.HandleUpdateData(decompressedBody)
}

func (cl *Client) HandleUpdateData(d []byte) {
	s, err := packet.UnmarshalUpdateObject(cl.Cfg.Build, d)
	if err != nil {
		log.Fatal(err)
	}

	for _, v := range s.Blocks {
		switch v.BlockData.Type() {
		case packet.UPDATETYPE_CREATE_OBJECT:
			yo.Ok("Creating", v.GUID)

		case packet.UPDATETYPE_CREATE_OBJECT2:
			yo.Ok("Spawn", v.GUID)

			if v.GUID == cl.PlayerGUID {
				yo.Puke(v)
			}
		}
	}
}

func (cl *Client) Handle() error {
	cl.On(packet.SMSG_WARDEN_DATA, cl.HandleWarden)
	cl.On(packet.SMSG_AUTH_RESPONSE, cl.HandleLogin)
	cl.On(packet.SMSG_CHAR_ENUM, cl.HandleCharList)
	cl.On(packet.SMSG_MOTD, cl.HandleMOTD)

	cl.On(packet.SMSG_INITIAL_SPELLS, cl.HandleSpellList)
	cl.On(packet.SMSG_EQUIPMENT_SET_LIST, cl.HandleEquip)
	cl.On(packet.SMSG_ACTION_BUTTONS, cl.HandleActions)
	cl.On(packet.SMSG_INITIALIZE_FACTIONS, cl.HandleReputations)
	cl.On(packet.SMSG_CONTACT_LIST, cl.HandleSocialList)
	cl.On(packet.SMSG_LEARNED_DANCE_MOVES, cl.HandleDanceMoves)
	cl.On(packet.SMSG_SET_FORCED_REACTIONS, cl.HandleForcedReactions)

	cl.On(packet.SMSG_UPDATE_OBJECT, cl.HandleUpdateData)
	cl.On(packet.SMSG_COMPRESSED_UPDATE_OBJECT, cl.HandleCompressedUpdate)
	cl.On(packet.SMSG_BINDPOINTUPDATE, cl.HandleBindPointUpdate)

	cl.On(packet.SMSG_MESSAGECHAT, cl.HandleMessage)

	for {
		yo.Warn("Reading frame...")
		frame, err := cl.Crypter.ReadFrame()
		if err != nil {
			return err
		}

		yo.Println(frame.Type)

		if h := cl.Handlers[frame.Type]; h != nil {
			h.Fn(frame.Data)
		}
	}
}

func (c *Client) Send(gp *packet.GamePacket) {
	c.Crypter.SendFrame(packet.Frame{
		gp.Type,
		gp.Bytes(),
	})
}
