package update

import (
	"fmt"
	"strings"

	"github.com/superp00t/etc"
	"github.com/superp00t/gophercraft/guid"
)

//go:generate gcraft_stringer -type=SpeedType
type SpeedType uint8
type UpdateFlags uint32
type MoveFlags uint64

const (
	Walk SpeedType = iota
	Run
	RunBackward
	Swim
	SwimBackward
	Flight
	FlightBackward
	Turn
	Pitch
)

const (
	UpdateFlagNone UpdateFlags = 1 << iota
	UpdateFlagSelf
	UpdateFlagTransport
	UpdateFlagFullGUID
	UpdateFlagHighGUID
	UpdateFlagAll
	UpdateFlagLiving
	UpdateFlagHasPosition
)

const (
	MoveFlagNone MoveFlags = 1 << iota
	MoveFlagForward
	MoveFlagBackward
	MoveFlagStrafeLeft
	MoveFlagStrafeRight
	MoveFlagTurnLeft
	MoveFlagTurnRight
	MoveFlagPitchUp
	MoveFlagPitchDown
	MoveFlagWalkMode
	MoveFlagLevitating
	MoveFlagFlying
	MoveFlagFalling
	MoveFlagFallingFar
	MoveFlagSwimming
	MoveFlagSplineEnabled
	MoveFlagCanFly
	MoveFlagFlyingOld
	MoveFlagOnTransport
	MoveFlagSplineElevation
	MoveFlagRoot
	MoveFlagWaterwalking
	MoveFlagSafeFall
	MoveFlagHover
)

func (uf UpdateFlags) String() string {
	s := []string{}

	if uf&UpdateFlagSelf != 0 {
		s = append(s, "UpdateFlagSelf")
	}

	if uf&UpdateFlagTransport != 0 {
		s = append(s, "UpdateFlagTransport")
	}

	if uf&UpdateFlagFullGUID != 0 {
		s = append(s, "UpdateFlagFullGUID")
	}

	if uf&UpdateFlagHighGUID != 0 {
		s = append(s, "UpdateFlagHighGUID")
	}

	if uf&UpdateFlagAll != 0 {
		s = append(s, "UpdateFlagAll")
	}

	if uf&UpdateFlagLiving != 0 {
		s = append(s, "UpdateFlagLiving")
	}

	if uf&UpdateFlagHasPosition != 0 {
		s = append(s, "UpdateFlagHasPosition")
	}

	return strings.Join(s, "|")
}

func (mf MoveFlags) String() string {
	var s []string
	if mf&MoveFlagForward != 0 {
		s = append(s, "MoveFlagForward")
	}
	if mf&MoveFlagBackward != 0 {
		s = append(s, "MoveFlagBackward")
	}
	if mf&MoveFlagStrafeLeft != 0 {
		s = append(s, "MoveFlagStrafeLeft")
	}
	if mf&MoveFlagStrafeRight != 0 {
		s = append(s, "MoveFlagStrafeRight")
	}
	if mf&MoveFlagTurnLeft != 0 {
		s = append(s, "MoveFlagTurnLeft")
	}
	if mf&MoveFlagTurnRight != 0 {
		s = append(s, "MoveFlagTurnRight")
	}
	if mf&MoveFlagPitchUp != 0 {
		s = append(s, "MoveFlagPitchUp")
	}
	if mf&MoveFlagPitchDown != 0 {
		s = append(s, "MoveFlagPitchDown")
	}
	if mf&MoveFlagWalkMode != 0 {
		s = append(s, "MoveFlagWalkMode")
	}
	if mf&MoveFlagLevitating != 0 {
		s = append(s, "MoveFlagLevitating")
	}
	if mf&MoveFlagFlying != 0 {
		s = append(s, "MoveFlagFlying")
	}
	if mf&MoveFlagFalling != 0 {
		s = append(s, "MoveFlagFalling")
	}
	if mf&MoveFlagFallingFar != 0 {
		s = append(s, "MoveFlagFallingFar")
	}
	if mf&MoveFlagSwimming != 0 {
		s = append(s, "MoveFlagSwimming")
	}
	if mf&MoveFlagSplineEnabled != 0 {
		s = append(s, "MoveFlagSplineEnabled")
	}
	if mf&MoveFlagCanFly != 0 {
		s = append(s, "MoveFlagCanFly")
	}
	if mf&MoveFlagFlyingOld != 0 {
		s = append(s, "MoveFlagFlyingOld")
	}
	if mf&MoveFlagOnTransport != 0 {
		s = append(s, "MoveFlagOnTransport")
	}
	if mf&MoveFlagSplineElevation != 0 {
		s = append(s, "MoveFlagSplineElevation")
	}
	if mf&MoveFlagRoot != 0 {
		s = append(s, "MoveFlagRoot")
	}
	if mf&MoveFlagWaterwalking != 0 {
		s = append(s, "MoveFlagWaterwalking")
	}
	if mf&MoveFlagSafeFall != 0 {
		s = append(s, "MoveFlagSafeFall")
	}
	if mf&MoveFlagHover != 0 {
		s = append(s, "MoveFlagHover")
	}
	if len(s) == 0 {
		return "MoveFlagNone"
	}
	return strings.Join(s, "|")
}

// map serverside storage codes to per-version client codes
type UpdateFlagDescriptor map[UpdateFlags]uint32
type MoveFlagDescriptor map[MoveFlags]uint32

var (
	UpdateFlagDescriptors = map[uint32]UpdateFlagDescriptor{
		5875: {
			UpdateFlagSelf:        0x0001,
			UpdateFlagTransport:   0x0002,
			UpdateFlagFullGUID:    0x0004,
			UpdateFlagHighGUID:    0x0008,
			UpdateFlagAll:         0x0010,
			UpdateFlagLiving:      0x0020,
			UpdateFlagHasPosition: 0x0040,
		},
	}

	MoveFlagDescriptors = map[uint32]MoveFlagDescriptor{
		5875: {
			MoveFlagForward:     0x00000001,
			MoveFlagBackward:    0x00000002,
			MoveFlagStrafeLeft:  0x00000004,
			MoveFlagStrafeRight: 0x00000008,
			MoveFlagTurnLeft:    0x00000010,
			MoveFlagTurnRight:   0x00000020,
			MoveFlagPitchUp:     0x00000040,
			MoveFlagPitchDown:   0x00000080,
			MoveFlagWalkMode:    0x00000100, // Walking

			MoveFlagLevitating:    0x00000400,
			MoveFlagFlying:        0x00000800, // [-ZERO] is it really need and correct value,
			MoveFlagFalling:       0x00002000,
			MoveFlagFallingFar:    0x00004000,
			MoveFlagSwimming:      0x00200000, // appears with fly flag also
			MoveFlagSplineEnabled: 0x00400000,
			MoveFlagCanFly:        0x00800000, // [-ZERO] is it really need and correct value,
			MoveFlagFlyingOld:     0x01000000, // [-ZERO] is it really need and correct value,

			MoveFlagOnTransport:     0x02000000, // Used for flying on some creatures,
			MoveFlagSplineElevation: 0x04000000, // used for flight paths,
			MoveFlagRoot:            0x08000000, // used for flight paths,
			MoveFlagWaterwalking:    0x10000000, // prevent unit from falling through water,
			MoveFlagSafeFall:        0x20000000, // active rogue safe fall spell (passive),
			MoveFlagHover:           0x40000000,
		},
	}

	SpeedLists = map[uint32][]SpeedType{
		5875: {
			Walk,
			Run,
			RunBackward,
			Swim,
			SwimBackward,
			Turn,
		},
	}
)

type Quaternion struct {
	Point3
	O float32
}

func EncodeQuaternion(out *etc.Buffer, q Quaternion) {
	EncodePoint3(out, q.Point3)
	out.WriteFloat32(q.O)
}

func DecodeQuaternion(in *etc.Buffer) Quaternion {
	q := Quaternion{}
	q.Point3 = DecodePoint3(in)
	q.O = in.ReadFloat32()
	return q
}

type Speeds map[SpeedType]float32

type MovementBlock struct {
	UpdateFlags UpdateFlags
	Info        *MovementInfo
	Speeds      Speeds
	Spline      *MoveSpline
	Position    Quaternion
	All         uint32
	HighGUID    uint32
	Victim      guid.GUID
	WorldTime   uint32
}

func decodeUpdateFlags(version uint32, in *etc.Buffer) (UpdateFlags, error) {
	descriptor, ok := UpdateFlagDescriptors[version]
	if !ok {
		return 0, fmt.Errorf("update: no move flag descriptor for %d", version)
	}

	uf := uint32(in.ReadByte())

	out := UpdateFlags(0)

	for k, v := range descriptor {
		if uf&v != 0 {
			out |= k
		}
	}

	return out, nil
}

func encodeUpdateFlags(version uint32, outb *etc.Buffer, uf UpdateFlags) error {
	descriptor, ok := UpdateFlagDescriptors[version]
	if !ok {
		return fmt.Errorf("update: no move flag descriptor for %d", version)
	}

	out := uint8(0)

	for k, v := range descriptor {
		if uf&k != 0 {
			out |= uint8(v)
		}
	}

	outb.WriteByte(out)

	return nil
}

func decodeMoveFlags(version uint32, in *etc.Buffer) (MoveFlags, error) {
	descriptor, ok := MoveFlagDescriptors[version]
	if !ok {
		return 0, fmt.Errorf("update: no move flag descriptor for %d", version)
	}

	mf := in.ReadUint32()

	out := MoveFlags(0)

	for k, v := range descriptor {
		if mf&v != 0 {
			out |= k
		}
	}

	return out, nil
}

func encodeMoveFlags(version uint32, outb *etc.Buffer, mf MoveFlags) error {
	descriptor, ok := MoveFlagDescriptors[version]
	if !ok {
		return fmt.Errorf("update: no move flag descriptor for %d", version)
	}

	out := uint32(0)

	for k, v := range descriptor {
		if mf&k != 0 {
			out |= v
		}
	}

	outb.WriteUint32(out)

	return nil
}

func EncodeMovementInfo(version uint32, out *etc.Buffer, mi *MovementInfo) error {
	err := encodeMoveFlags(version, out, mi.Flags)
	if err != nil {
		return err
	}

	out.WriteUint32(mi.Time)
	EncodeQuaternion(out, mi.Position)

	if mi.Flags&MoveFlagOnTransport != 0 {
		mi.TransportGUID.EncodePacked(version, out)
		EncodeQuaternion(out, mi.TransportPosition)
	}

	if mi.Flags&MoveFlagSwimming != 0 {
		out.WriteFloat32(mi.SwimPitch)
	}

	out.WriteUint32(mi.FallTime)

	if mi.Flags&MoveFlagFalling != 0 {
		out.WriteFloat32(mi.FallVelocity)
		out.WriteFloat32(mi.FallSin)
		out.WriteFloat32(mi.FallCos)
		out.WriteFloat32(mi.FallXYSpeed)
	}

	if mi.Flags&MoveFlagSplineElevation != 0 {
		out.WriteFloat32(mi.SplineElevation)
	}

	return nil
}

func DecodeMovementInfo(version uint32, in *etc.Buffer) (*MovementInfo, error) {
	info := new(MovementInfo)
	var err error
	info.Flags, err = decodeMoveFlags(version, in)
	if err != nil {
		return nil, err
	}
	info.Time = in.ReadUint32()
	info.Position = DecodeQuaternion(in)

	if info.Flags&MoveFlagOnTransport != 0 {
		info.TransportGUID, err = guid.DecodePacked(version, in)
		if err != nil {
			return nil, err
		}
		info.TransportPosition = DecodeQuaternion(in)
	}

	if info.Flags&MoveFlagSwimming != 0 {
		info.SwimPitch = in.ReadFloat32()
	}

	info.FallTime = in.ReadUint32()

	if info.Flags&MoveFlagFalling != 0 {
		info.FallVelocity = in.ReadFloat32()
		info.FallSin = in.ReadFloat32()
		info.FallCos = in.ReadFloat32()
		info.FallXYSpeed = in.ReadFloat32()
	}

	if info.Flags&MoveFlagSplineElevation != 0 {
		info.SplineElevation = in.ReadFloat32()
	}

	return info, nil
}

type MovementInfo struct {
	Flags    MoveFlags
	Time     uint32
	Position Quaternion

	TransportGUID     guid.GUID
	TransportPosition Quaternion
	TransportTime     uint32

	SwimPitch    float32
	FallTime     uint32
	FallVelocity float32
	FallSin      float32
	FallCos      float32
	FallXYSpeed  float32

	SplineElevation float32
}

// only supports 5875 so far
func DecodeMovementBlock(version uint32, in *etc.Buffer) (*MovementBlock, error) {
	mBlock := new(MovementBlock)

	var err error
	mBlock.UpdateFlags, err = decodeUpdateFlags(version, in)
	if err != nil {
		return nil, err
	}

	if mBlock.UpdateFlags&UpdateFlagLiving != 0 {
		var err error
		mBlock.Info, err = DecodeMovementInfo(version, in)
		if err != nil {
			return nil, err
		}

		mBlock.Speeds = make(map[SpeedType]float32)

		for _, speed := range SpeedLists[version] {
			mBlock.Speeds[speed] = in.ReadFloat32()
		}

		if mBlock.Info.Flags&MoveFlagSplineEnabled != 0 {
			mBlock.Spline, err = DecodeMoveSpline(version, in)
			if err != nil {
				return nil, err
			}
		}
	} else if mBlock.UpdateFlags&UpdateFlagHasPosition != 0 {
		mBlock.Position = DecodeQuaternion(in)
	}

	if mBlock.UpdateFlags&UpdateFlagHighGUID != 0 {
		mBlock.HighGUID = in.ReadUint32()
	}

	if mBlock.UpdateFlags&UpdateFlagAll != 0 {
		mBlock.All = in.ReadUint32()
	}

	if mBlock.UpdateFlags&UpdateFlagTransport != 0 {
		mBlock.WorldTime = in.ReadUint32()
	}

	return mBlock, nil
}

func (mb *MovementBlock) WriteTo(g guid.GUID, e *Encoder) error {
	if err := encodeUpdateFlags(e.Version, e.Buffer, mb.UpdateFlags); err != nil {
		return err
	}

	if mb.UpdateFlags&UpdateFlagLiving != 0 {
		if mb.Info == nil {
			return fmt.Errorf("update: error serializing MovementBlock: living bit set but Info is nil")
		}

		err := EncodeMovementInfo(e.Version, e.Buffer, mb.Info)
		if err != nil {
			return err
		}

		sl, ok := SpeedLists[e.Version]
		if !ok {
			return fmt.Errorf("update: no SpeedLists for version %d", e.Version)
		}

		for _, v := range sl {
			e.Buffer.WriteFloat32(mb.Speeds[v])
		}

		if mb.Info.Flags&MoveFlagSplineEnabled != 0 {
			err = EncodeMoveSpline(e.Version, e.Buffer, mb.Spline)
			if err != nil {
				return err
			}
		}
	} else if mb.UpdateFlags&UpdateFlagHasPosition != 0 {
		EncodeQuaternion(e.Buffer, mb.Position)
	}

	if mb.UpdateFlags&UpdateFlagHighGUID != 0 {
		e.WriteUint32(mb.HighGUID)
	}

	if mb.UpdateFlags&UpdateFlagAll != 0 {
		e.WriteUint32(mb.All)
	}

	if mb.UpdateFlags&UpdateFlagTransport != 0 {
		e.WriteUint32(mb.WorldTime)
	}

	return nil
}
