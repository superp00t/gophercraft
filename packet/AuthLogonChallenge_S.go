package packet

import (
	"fmt"

	"github.com/superp00t/etc"
)

type AuthLogonChallenge_S struct {
	Cmd              AuthType
	Error            ErrorType
	B                []byte // 32 long
	G                []byte
	N                []byte // 32 long
	S                []byte // 32 long
	VersionChallenge []byte // 16 long
	SecurityFlags    uint8
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
	buf.WriteByte(uint8(len(acls.G)))
	buf.Write(acls.G)

	// N
	buf.WriteByte(uint8(len(acls.N)))
	buf.Write(acls.N)
	buf.Write(acls.S)
	buf.Write(acls.VersionChallenge)
	buf.WriteByte(acls.SecurityFlags)
	return buf.Bytes()
}

func UnmarshalAuthLogonChallenge_S(input []byte) (*AuthLogonChallenge_S, error) {
	if len(input) < 86 {
		return nil, fmt.Errorf("Packet too small")
	}
	in := etc.FromBytes(input)

	alcs := &AuthLogonChallenge_S{}
	alcs.Cmd = AuthType(in.ReadByte())
	in.ReadByte() // Always zero
	alcs.Error = ErrorType(in.ReadByte())
	alcs.B = in.ReadBytes(32)
	gLen := in.ReadByte()
	alcs.G = in.ReadBytes(int(gLen))
	nLen := in.ReadByte()
	alcs.N = in.ReadBytes(int(nLen))
	alcs.S = in.ReadBytes(32)
	alcs.VersionChallenge = in.ReadBytes(16)
	alcs.SecurityFlags = in.ReadByte()

	return alcs, nil
}
