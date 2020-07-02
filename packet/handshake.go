package packet

import (
	"fmt"

	"github.com/superp00t/etc"
	"github.com/superp00t/etc/yo"
	"github.com/superp00t/gophercraft/vsn"
)

const (
	OPCODE_SIZE_OUTGOING = 6
	OPCODE_SIZE_INCOMING = 4
)

type SMSGAuthPacket struct {
	Type  WorldType
	Size  uint16
	Salt  []byte
	Seed1 []byte
	Seed2 []byte
}

func UnmarshalSMSGAuthPacket(build vsn.Build, input []byte) (*SMSGAuthPacket, error) {
	in := etc.FromBytes(input)

	gp := &SMSGAuthPacket{}
	gp.Size = in.ReadBigUint16()
	gp.Type = WorldType(in.ReadUint16())

	if build.RemovedIn(8606) {
		gp.Salt = in.ReadBytes(4)
		return gp, nil
	}

	in.ReadUint32()
	gp.Salt = in.ReadBytes(4)
	gp.Seed1 = in.ReadBytes(16)
	gp.Seed2 = in.ReadBytes(16)

	return gp, nil
}

func (s *SMSGAuthPacket) Encode(version vsn.Build) []byte {
	smsg := NewWorldPacket(SMSG_AUTH_CHALLENGE)
	if version == vsn.Alpha {
		smsg.Write(make([]byte, 6))
		return smsg.ServerMessage()
	}

	if version.RemovedIn(8606) {
		smsg.Write(s.Salt)
		return smsg.ServerMessage()
	}

	smsg.WriteUint32(0x01)
	smsg.Write(s.Salt)
	smsg.Write(s.Seed1)
	smsg.Write(s.Seed2)
	return smsg.ServerMessage()
}

type CMSGAuthSession struct {
	Build           vsn.Build
	LoginServerID   uint32
	Account         string // 0-terminated string
	LoginServerType uint32
	Seed            []byte
	RegionID        uint32
	BattlegroupID   uint32
	RealmID         uint32
	DosResponse     uint64
	Digest          []byte
	AddonData       []byte
}

func UnmarshalCMSGAuthSession(input []byte) (*CMSGAuthSession, error) {
	// opcode = input[0:4]
	// len    = input[4:6]
	if len(input) < 36 {
		return nil, fmt.Errorf("packet too small")
	}

	in := etc.FromBytes(input)
	length := in.ReadBigUint16()
	opcode := WorldType(in.ReadUint32())

	yo.Ok(opcode, length)

	c := &CMSGAuthSession{}
	c.Build = vsn.Build(in.ReadUint32())
	c.LoginServerID = in.ReadUint32()
	c.Account = in.ReadCString()

	yo.Ok("Account=", c.Account, "build=", c.Build)

	if c.Build.RemovedIn(8606) {
		c.Seed = in.ReadBytes(4)
		c.Digest = in.ReadBytes(20)
		return c, nil
	} else {
		yo.Warn("unknown type", c.Build)
		c.LoginServerType = in.ReadUint32()
		c.Seed = in.ReadBytes(4)
		c.RegionID = in.ReadUint32()
		c.BattlegroupID = in.ReadUint32()
		c.RealmID = in.ReadUint32()
		c.DosResponse = in.ReadUint64()
		c.Digest = in.ReadBytes(20)
		c.AddonData = in.ReadRemainder()
	}

	return c, nil
}

func (c *CMSGAuthSession) Encode() []byte {
	app := etc.NewBuffer()
	app.WriteUint32(uint32(c.Build))
	app.WriteUint32(c.LoginServerID)
	app.WriteCString(c.Account)

	if c.Build.RemovedIn(8606) {
		app.Write(c.Seed)
		app.Write(c.Digest)
		app.Write(c.AddonData)
	} else {
		app.WriteUint32(c.LoginServerType)
		app.Write(c.Seed)
		app.WriteUint32(c.RegionID)
		app.WriteUint32(c.BattlegroupID)
		app.WriteUint32(c.RealmID)
		app.WriteUint64(c.DosResponse)
		app.Write(c.Digest)
		app.Write(c.AddonData)
	}

	// Addon data
	env := etc.NewBuffer()
	env.WriteBigUint16(uint16(app.Len() + 4))
	env.WriteUint32(uint32(CMSG_AUTH_SESSION))
	env.Write(app.Bytes())
	return env.Bytes()
}

type SMSGAuthResponse struct {
	Cmd       uint8
	WaitQueue uint32
}

func UnmarshalSMSGAuthResponse(input []byte) (*SMSGAuthResponse, error) {
	p := etc.FromBytes(input)
	s := &SMSGAuthResponse{}
	s.Cmd = p.ReadByte()

	if s.Cmd != AUTH_WAIT_QUEUE {
		return s, nil
	}

	s.WaitQueue = p.ReadUint32()
	p.ReadByte()
	return s, nil
}
