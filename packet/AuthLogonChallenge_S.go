package packet

import (
	"fmt"

	"github.com/superp00t/etc"
)

type AuthLogonChallenge_S struct {
	Cmd   AuthType
	Error ErrorType
	B     []byte // 32 long
	G     uint8
	N     []byte // 32 long
	S     []byte // 32 long
	Unk3  []byte // 16 long
	// Unk4  uint8
}

func (acls *AuthLogonChallenge_S) Encode() []byte {
	buf := etc.NewBuffer()
	buf.Write([]byte{
		uint8(acls.Cmd),
		0x00,
		uint8(acls.Error),
	})

	buf.Write(acls.B)
	// G
	buf.Write([]byte{
		1, // g_len
		7, // value
	})

	// N
	buf.Write([]byte{
		32, // N_len
	})
	buf.Write(acls.N)
	buf.Write(acls.S)
	buf.Write(acls.Unk3)
	buf.Write([]byte{0x00})
	return buf.Bytes()
}

func UnmarshalAuthLogonChallenge_S(input []byte) (*AuthLogonChallenge_S, error) {
	if len(input) < 86 {
		return nil, fmt.Errorf("Packet too small")
	}
	alcs := &AuthLogonChallenge_S{}
	alcs.Cmd = AuthType(input[0])
	alcs.Error = ErrorType(input[2])
	alcs.B = input[3:35]
	// omit input[35], we know how long G is
	alcs.G = input[36]
	// omit input[37], we know how long N is
	alcs.N = input[38:70]
	alcs.S = input[70:102]
	alcs.Unk3 = input[70:86]

	return alcs, nil
}
