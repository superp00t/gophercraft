package packet

import (
	"fmt"

	"github.com/superp00t/etc"
)

type AuthLogonProof_S struct {
	Cmd          AuthType
	Error        ErrorType
	M2           []byte
	AccountFlags uint32
	SurveyID     uint32
	Unk3         uint16
}

func (alps *AuthLogonProof_S) Encode(build uint32) []byte {
	buf := etc.NewBuffer()
	buf.WriteByte(uint8(alps.Cmd))
	buf.WriteByte(uint8(alps.Error))
	buf.Write(alps.M2)
	if build == 12340 {
		buf.WriteUint32(alps.AccountFlags)
		buf.WriteUint32(alps.SurveyID)
		buf.WriteUint16(alps.Unk3)
	}

	if build == 5875 {
		buf.WriteUint32(0)
	}

	return buf.Bytes()
}

func UnmarshalAuthLogonProof_S(build uint32, input []byte) (*AuthLogonProof_S, error) {
	if len(input) < 26 {
		return nil, fmt.Errorf("packet: too small")
	}

	in := etc.FromBytes(input)
	alps := &AuthLogonProof_S{}
	alps.Cmd = AuthType(in.ReadByte())
	alps.Error = ErrorType(in.ReadByte())
	alps.M2 = in.ReadBytes(20)

	if build == 12340 {
		alps.AccountFlags = in.ReadUint32()
		alps.SurveyID = in.ReadUint32()
		alps.Unk3 = in.ReadUint16()
	}

	return alps, nil
}
