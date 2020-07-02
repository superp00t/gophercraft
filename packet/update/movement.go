package update

import (
	"fmt"
	"io"
	"strings"

	"github.com/superp00t/gophercraft/guid"
	"github.com/superp00t/gophercraft/vsn"
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
	UpdateFlagDescriptors = map[vsn.Build]UpdateFlagDescriptor{
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

	MoveFlagDescriptors = map[vsn.Build]MoveFlagDescriptor{
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

	SpeedLists = map[vsn.Build][]SpeedType{
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

type Position struct {
	Point3
	O float32
}

func EncodePosition(out io.Writer, q Position) error {
	err := EncodePoint3(out, q.Point3)
	if err != nil {
		return err
	}
	return writeFloat32(out, q.O)
}

func DecodePosition(in io.Reader) (Position, error) {
	var q Position
	var err error
	q.Point3, err = DecodePoint3(in)
	if err != nil {
		return q, err
	}
	q.O, err = readFloat32(in)
	return q, err
}

type Speeds map[SpeedType]float32

type MovementBlock struct {
	ID          guid.GUID
	UpdateFlags UpdateFlags
	Info        *MovementInfo
	Speeds      Speeds
	Spline      *MoveSpline
	Position    Position
	All         uint32
	HighGUID    uint32
	Victim      guid.GUID
	WorldTime   uint32
}

func decodeUpdateFlags(version vsn.Build, in io.Reader) (UpdateFlags, error) {
	descriptor, ok := UpdateFlagDescriptors[version]
	if !ok {
		return 0, fmt.Errorf("update: no move flag descriptor for %d", version)
	}

	ufb, err := readUint8(in)
	if err != nil {
		return 0, err
	}

	uf := uint32(ufb)

	out := UpdateFlags(0)

	for k, v := range descriptor {
		if uf&v != 0 {
			out |= k
		}
	}

	return out, nil
}

func encodeUpdateFlags(version vsn.Build, outb io.Writer, uf UpdateFlags) error {
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

	return writeUint8(outb, out)
}

func decodeMoveFlags(version vsn.Build, in io.Reader) (MoveFlags, error) {
	descriptor, ok := MoveFlagDescriptors[version]
	if !ok {
		return 0, fmt.Errorf("update: no move flag descriptor for %d", version)
	}

	mf, err := readUint32(in)
	if err != nil {
		return 0, err
	}

	out := MoveFlags(0)

	for k, v := range descriptor {
		if mf&v != 0 {
			out |= k
		}
	}

	return out, nil
}

func encodeMoveFlags(version vsn.Build, outb io.Writer, mf MoveFlags) error {
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

	return writeUint32(outb, out)
}

func EncodeMovementInfo(version vsn.Build, out io.Writer, mi *MovementInfo) error {
	err := encodeMoveFlags(version, out, mi.Flags)
	if err != nil {
		return err
	}

	if err = writeUint32(out, mi.Time); err != nil {
		return err
	}

	if err = EncodePosition(out, mi.Position); err != nil {
		return err
	}

	if mi.Flags&MoveFlagOnTransport != 0 {
		mi.TransportGUID.EncodePacked(version, out)
		if err = EncodePosition(out, mi.TransportPosition); err != nil {
			return err
		}
	}

	if mi.Flags&MoveFlagSwimming != 0 {
		if err = writeFloat32(out, mi.SwimPitch); err != nil {
			return err
		}
	}

	if err = writeUint32(out, mi.FallTime); err != nil {
		return err
	}

	if mi.Flags&MoveFlagFalling != 0 {
		if err = writeFloat32(out, mi.FallVelocity); err != nil {
			return err
		}
		if err = writeFloat32(out, mi.FallSin); err != nil {
			return err
		}
		if err = writeFloat32(out, mi.FallCos); err != nil {
			return err
		}
		if err = writeFloat32(out, mi.FallXYSpeed); err != nil {
			return err
		}
	}

	if mi.Flags&MoveFlagSplineElevation != 0 {
		if err = writeFloat32(out, mi.SplineElevation); err != nil {
			return err
		}
	}

	return nil
}

func DecodeMovementInfo(version vsn.Build, in io.Reader) (*MovementInfo, error) {
	info := new(MovementInfo)
	var err error
	info.Flags, err = decodeMoveFlags(version, in)
	if err != nil {
		return nil, err
	}
	info.Time, err = readUint32(in)
	if err != nil {
		return nil, err
	}
	info.Position, err = DecodePosition(in)
	if err != nil {
		return nil, err
	}

	if info.Flags&MoveFlagOnTransport != 0 {
		info.TransportGUID, err = guid.DecodePacked(version, in)
		if err != nil {
			return nil, err
		}
		info.TransportPosition, err = DecodePosition(in)
		if err != nil {
			return nil, err
		}
	}

	if info.Flags&MoveFlagSwimming != 0 {
		info.SwimPitch, err = readFloat32(in)
		if err != nil {
			return nil, err
		}
	}

	info.FallTime, err = readUint32(in)
	if err != nil {
		return nil, err
	}

	if info.Flags&MoveFlagFalling != 0 {
		info.FallVelocity, err = readFloat32(in)
		if err != nil {
			return nil, err
		}
		info.FallSin, err = readFloat32(in)
		if err != nil {
			return nil, err
		}
		info.FallCos, err = readFloat32(in)
		if err != nil {
			return nil, err
		}
		info.FallXYSpeed, err = readFloat32(in)
		if err != nil {
			return nil, err
		}
	}

	if info.Flags&MoveFlagSplineElevation != 0 {
		info.SplineElevation, err = readFloat32(in)
		if err != nil {
			return nil, err
		}
	}

	return info, nil
}

type MovementInfo struct {
	Flags    MoveFlags
	Time     uint32
	Position Position

	TransportGUID     guid.GUID
	TransportPosition Position
	TransportTime     uint32

	SwimPitch    float32
	FallTime     uint32
	FallVelocity float32
	FallSin      float32
	FallCos      float32
	FallXYSpeed  float32

	SplineElevation float32
}

func (mBlock *MovementBlock) Type() BlockType {
	return Movement
}

func (decoder *Decoder) IsCreateBlock() bool {
	return decoder.CurrentBlockType == CreateObject || decoder.CurrentBlockType == SpawnObject
}

// only supports 5875 so far
func (decoder *Decoder) DecodeMovementBlock() (*MovementBlock, error) {
	mBlock := new(MovementBlock)
	var err error

	mBlock.UpdateFlags, err = decodeUpdateFlags(decoder.Build, decoder.Reader)
	if err != nil {
		return nil, err
	}

	if mBlock.UpdateFlags&UpdateFlagLiving != 0 {
		var err error
		mBlock.Info, err = DecodeMovementInfo(decoder.Build, decoder.Reader)
		if err != nil {
			return nil, err
		}

		mBlock.Speeds = make(map[SpeedType]float32)

		for _, speed := range SpeedLists[decoder.Build] {
			mBlock.Speeds[speed], err = readFloat32(decoder.Reader)
			if err != nil {
				return nil, err
			}
		}

		if mBlock.Info.Flags&MoveFlagSplineEnabled != 0 {
			mBlock.Spline, err = DecodeMoveSpline(decoder.Build, decoder.Reader)
			if err != nil {
				return nil, err
			}
		}
	} else if mBlock.UpdateFlags&UpdateFlagHasPosition != 0 {
		mBlock.Position, err = DecodePosition(decoder.Reader)
		if err != nil {
			return nil, err
		}
	}

	if mBlock.UpdateFlags&UpdateFlagHighGUID != 0 {
		mBlock.HighGUID, err = readUint32(decoder.Reader)
		if err != nil {
			return nil, err
		}
	}

	if mBlock.UpdateFlags&UpdateFlagAll != 0 {
		mBlock.All, err = readUint32(decoder.Reader)
		if err != nil {
			return nil, err
		}
	}

	if mBlock.UpdateFlags&UpdateFlagTransport != 0 {
		mBlock.WorldTime, err = readUint32(decoder.Reader)
		if err != nil {
			return nil, err
		}
	}

	return mBlock, nil
}

func (mb *MovementBlock) WriteData(e *Encoder, mask VisibilityFlags, create bool) error {
	if err := encodeUpdateFlags(e.Build, e, mb.UpdateFlags); err != nil {
		return err
	}

	if mb.UpdateFlags&UpdateFlagLiving != 0 {
		if mb.Info == nil {
			return fmt.Errorf("update: error serializing MovementBlock: living bit set but Info is nil")
		}

		err := EncodeMovementInfo(e.Build, e, mb.Info)
		if err != nil {
			return err
		}

		sl, ok := SpeedLists[e.Build]
		if !ok {
			return fmt.Errorf("update: no SpeedLists for version %s", e.Build)
		}

		for _, v := range sl {
			if err := writeFloat32(e, mb.Speeds[v]); err != nil {
				return err
			}
		}

		if mb.Info.Flags&MoveFlagSplineEnabled != 0 {
			err = EncodeMoveSpline(e.Build, e, mb.Spline)
			if err != nil {
				return err
			}
		}
	} else if mb.UpdateFlags&UpdateFlagHasPosition != 0 {
		EncodePosition(e, mb.Position)
	}

	if mb.UpdateFlags&UpdateFlagHighGUID != 0 {
		if err := writeUint32(e, mb.HighGUID); err != nil {
			return err
		}
	}

	if mb.UpdateFlags&UpdateFlagAll != 0 {
		if err := writeUint32(e, mb.All); err != nil {
			return err
		}
	}

	if mb.UpdateFlags&UpdateFlagTransport != 0 {
		if err := writeUint32(e, mb.WorldTime); err != nil {
			return err
		}
	}

	return nil
}
