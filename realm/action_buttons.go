package realm

import (
	"encoding/binary"

	"github.com/davecgh/go-spew/spew"
	"github.com/superp00t/etc"
	"github.com/superp00t/etc/yo"
	"github.com/superp00t/gophercraft/packet"
	"github.com/superp00t/gophercraft/vsn"
	"github.com/superp00t/gophercraft/realm/wdb"
)

func (s *Session) NumActionButtons() int {
	numActionButtons := 120

	if s.Build().AddedIn(vsn.V2_4_3) {
		numActionButtons = 132
	}

	return numActionButtons
}

func (s *Session) SendActionButtons() {
	numActionButtons := uint8(s.NumActionButtons())

	p := packet.NewWorldPacket(packet.SMSG_ACTION_BUTTONS)
	var actionButtons = make([]struct {
		Type   uint8
		Action uint32
	}, numActionButtons)

	var ab []wdb.ActionButton
	if err := s.DB().Where("player = ?", s.PlayerID()).Find(&ab); err != nil {
		panic(err)
	}

	for _, v := range ab {
		if v.Button >= numActionButtons {
			yo.Warn("Overflowed action buttons for", s.PlayerID(), spew.Sdump(v))
			continue
		}

		actionButtons[int(v.Button)].Type = v.Type
		actionButtons[int(v.Button)].Action = v.Action
	}

	// This is accomplished with bitwise operations in MaNGOS, but I think this is more readable (I hope)
	// <24 bit action ID> <8 bit button type ID>
	for _, v := range actionButtons {
		var action [4]byte
		binary.LittleEndian.PutUint32(action[:], v.Action)
		action[3] = v.Type
		p.Write(action[:])
	}

	s.SendAsync(p)
}

func (s *Session) HandleSetActionButton(e *etc.Buffer) {
	button := e.ReadByte()
	actionBytes := e.ReadBytes(3)
	actionBytes = append(actionBytes, 0)
	action := binary.LittleEndian.Uint32(actionBytes)
	actionType := e.ReadByte()

	s.DB().Where("player = ?", s.PlayerID()).Where("button = ?", button).Delete(new(wdb.ActionButton))

	if button >= uint8(s.NumActionButtons()) {
		return
	}

	if action != 0 {
		s.DB().Insert(&wdb.ActionButton{
			Player: s.PlayerID(),
			Button: button,
			Action: action,
			Type:   actionType,
		})
	}
}
