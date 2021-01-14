package auth

import (
	"fmt"
	"strings"

	"github.com/superp00t/etc"
	"github.com/superp00t/gophercraft/vsn"
)

// AuthLogonChallenge_C is the first packet sent by a client
// while initiating a connection to an authserver.
type AuthLogonChallenge_C struct {
	Cmd          AuthType
	Error        ErrorType
	Size         uint16
	GameName     string // Encode in reverse.
	Version      [3]byte
	Build        uint16
	Platform     string
	OS           string
	Country      string
	TimezoneBias uint32
	IP           uint32
	I            string
}

func (alcc *AuthLogonChallenge_C) Encode() []byte {
	a := etc.NewBuffer()
	a.WriteByte(uint8(alcc.Cmd))
	a.WriteByte(uint8(alcc.Error))

	b := etc.NewBuffer()
	b.WriteInvertedString(4, alcc.GameName)
	b.Write(alcc.Version[:])
	b.WriteUint16(alcc.Build)
	b.WriteInvertedString(4, alcc.Platform)
	b.WriteInvertedString(4, alcc.OS)
	b.WriteInvertedString(4, alcc.Country)
	b.WriteUint32(alcc.TimezoneBias)
	b.WriteUint32(alcc.IP)
	b.WriteByte(uint8(len(alcc.I)))
	b.Write([]byte(alcc.I))

	a.WriteUint16(uint16(b.Len()))
	a.Write(b.Bytes())

	return a.Bytes()
}

func (alcc *AuthLogonChallenge_C) VersionString() string {
	return fmt.Sprintf("%d.%d.%d", alcc.Version[0], alcc.Version[1], alcc.Version[2])
}

func UnmarshalAuthLogonChallenge_C(data []byte) (*AuthLogonChallenge_C, error) {
	if len(data) < 34 {
		return nil, fmt.Errorf("Packet too small")
	}
	in := etc.FromBytes(data)
	ac := &AuthLogonChallenge_C{}
	ac.Cmd = AuthType(in.ReadByte())
	ac.Error = ErrorType(in.ReadByte())
	ac.Size = in.ReadUint16()
	ac.GameName = in.ReadInvertedString(4)
	copy(ac.Version[:], in.ReadBytes(3))
	ac.Build = in.ReadUint16()
	ac.Platform = in.ReadInvertedString(4)
	ac.OS = in.ReadInvertedString(4)
	ac.Country = in.ReadInvertedString(4)
	ac.TimezoneBias = in.ReadUint32()
	ac.IP = in.ReadUint32()
	ac.I = string(in.ReadBytes(int(in.ReadByte())))

	return ac, nil
}

// LogonChallengePacket_C is a helper function to simplify the client library.
func LogonChallengePacket_C(build vsn.Build, username string) []byte {
	alcc := &AuthLogonChallenge_C{
		Cmd:          AUTH_LOGON_CHALLENGE,
		Error:        8,
		GameName:     "WoW",
		Version:      Version(build),
		Build:        uint16(build),
		Platform:     "x86",
		OS:           "Win",
		Country:      "enGB",
		TimezoneBias: 4294966996,
		IP:           16777343,
		I:            strings.ToUpper(username),
	}

	return alcc.Encode()
}

func Version(build vsn.Build) [3]byte {
	switch build {
	case 5875:
		return [3]byte{1, 12, 1}
	case 12340:
		return [3]byte{3, 3, 5}
	}

	return [3]byte{1, 12, 1}
}
