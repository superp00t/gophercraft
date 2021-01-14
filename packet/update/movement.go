package update

import (
	"encoding/binary"
	"fmt"
	"io"
	"strconv"
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
	UpdateFlagHasAttackingTarget
	UpdateFlagLowGUID
	UpdateFlagHighGUID
	UpdateFlagAll
	UpdateFlagLiving
	UpdateFlagHasPosition
	UpdateFlagVehicle
	UpdateFlagPosition
	UpdateFlagRotation
)

const (
	MoveFlagForward MoveFlags = 1 << iota
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
	// Alpha
	MoveFlagTimeValid
	MoveFlagImmobilized
	MoveFlagDontCollide
	MoveFlagRedirected
	MoveFlagPendingStop
	MoveFlagPendingUnstrafe
	MoveFlagPendingFall
	MoveFlagPendingForward
	MoveFlagPendingBackward
	MoveFlagPendingStrLeft
	MoveFlagPendingStrRight
	MoveFlagMoved
	MoveFlagSliding
	// TBC
	MoveFlagAscending
	// WOTLK
	MoveFlagNoStrafe
	MoveFlagFullSpeedTurning
	MoveFlagFullSpeedPitching
	MoveFlagAllowPitching
	MoveFlagInterpolateMovement
	MoveFlagInterpolateTurning
	MoveFlagInterpolatePitching
)

func (uf UpdateFlags) String() string {
	s := []string{}

	if uf&UpdateFlagSelf != 0 {
		s = append(s, "UpdateFlagSelf")
	}

	if uf&UpdateFlagTransport != 0 {
		s = append(s, "UpdateFlagTransport")
	}

	if uf&UpdateFlagHasAttackingTarget != 0 {
		s = append(s, "UpdateFlagHasAttackingTarget")
	}

	if uf&UpdateFlagLowGUID != 0 {
		s = append(s, "UpdateFlagLowGUID")
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

	if uf&UpdateFlagVehicle != 0 {
		s = append(s, "UpdateFlagVehicle")
	}

	if uf&UpdateFlagPosition != 0 {
		s = append(s, "UpdateFlagPosition")
	}

	if uf&UpdateFlagRotation != 0 {
		s = append(s, "UpdateFlagRotation")
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
	if mf&MoveFlagAscending != 0 {
		s = append(s, "MoveFlagAscending")
	}
	if mf&MoveFlagNoStrafe != 0 {
		s = append(s, "MoveFlagNoStrafe")
	}
	if mf&MoveFlagFullSpeedTurning != 0 {
		s = append(s, "MoveFlagFullSpeedTurning")
	}
	if mf&MoveFlagFullSpeedPitching != 0 {
		s = append(s, "MoveFlagFullSpeedPitching")
	}
	if mf&MoveFlagAllowPitching != 0 {
		s = append(s, "MoveFlagAllowPitching")
	}
	if mf&MoveFlagInterpolateMovement != 0 {
		s = append(s, "MoveFlagInterpolateMovement")
	}
	if mf&MoveFlagInterpolateTurning != 0 {
		s = append(s, "MoveFlagInterpolateTurning")
	}
	if mf&MoveFlagInterpolatePitching != 0 {
		s = append(s, "MoveFlagInterpolatePitching")
	}
	if len(s) == 0 {
		return "MoveFlagNone"
	}
	return strings.Join(s, "|")
}

// map serverside storage codes to per-version client codes
type UpdateFlagDescriptor map[UpdateFlags]uint32
type MoveFlagDescriptor map[MoveFlags]uint64

var (
	UpdateFlagDescriptors = map[vsn.Build]UpdateFlagDescriptor{
		vsn.V1_12_1: {
			UpdateFlagSelf:               0x0001,
			UpdateFlagTransport:          0x0002,
			UpdateFlagHasAttackingTarget: 0x0004,
			UpdateFlagHighGUID:           0x0008,
			UpdateFlagAll:                0x0010,
			UpdateFlagLiving:             0x0020,
			UpdateFlagHasPosition:        0x0040,
		},

		vsn.V2_4_3: {
			UpdateFlagSelf:               0x0001,
			UpdateFlagTransport:          0x0002,
			UpdateFlagHasAttackingTarget: 0x0004,
			UpdateFlagLowGUID:            0x0008,
			UpdateFlagHighGUID:           0x0010,
			UpdateFlagLiving:             0x0020,
			UpdateFlagHasPosition:        0x0040,
		},

		vsn.V3_3_5a: {
			UpdateFlagSelf:               0x0001,
			UpdateFlagTransport:          0x0002,
			UpdateFlagHasAttackingTarget: 0x0004,
			UpdateFlagLowGUID:            0x0008,
			UpdateFlagHighGUID:           0x0010,
			UpdateFlagLiving:             0x0020,
			UpdateFlagHasPosition:        0x0040,
			UpdateFlagVehicle:            0x0080,
			UpdateFlagPosition:           0x0100,
			UpdateFlagRotation:           0x0200,
		},
	}

	MoveFlagDescriptors = map[vsn.Build]MoveFlagDescriptor{
		vsn.Alpha: {
			MoveFlagForward:         0x1,
			MoveFlagBackward:        0x2,
			MoveFlagStrafeLeft:      0x4,
			MoveFlagStrafeRight:     0x8,
			MoveFlagTurnLeft:        0x10,
			MoveFlagTurnRight:       0x20,
			MoveFlagPitchUp:         0x40,
			MoveFlagPitchDown:       0x80,
			MoveFlagWalkMode:        0x100,
			MoveFlagTimeValid:       0x200,
			MoveFlagImmobilized:     0x400,
			MoveFlagDontCollide:     0x800,
			MoveFlagRedirected:      0x1000,
			MoveFlagRoot:            0x2000,
			MoveFlagFalling:         0x4000,
			MoveFlagFallingFar:      0x8000,
			MoveFlagPendingStop:     0x10000,
			MoveFlagPendingUnstrafe: 0x20000,
			MoveFlagPendingFall:     0x40000,
			MoveFlagPendingForward:  0x80000,
			MoveFlagPendingBackward: 0x100000,
			MoveFlagPendingStrLeft:  0x200000,
			MoveFlagPendingStrRight: 0x400000,
			MoveFlagMoved:           0x800000,
			MoveFlagSliding:         0x1000000,
			MoveFlagSwimming:        0x2000000,
			MoveFlagSplineEnabled:   0x4000000,
		},

		vsn.V1_12_1: {
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

		vsn.V2_4_3: {
			MoveFlagForward:         0x00000001,
			MoveFlagBackward:        0x00000002,
			MoveFlagStrafeLeft:      0x00000004,
			MoveFlagStrafeRight:     0x00000008,
			MoveFlagTurnLeft:        0x00000010,
			MoveFlagTurnRight:       0x00000020,
			MoveFlagPitchUp:         0x00000040,
			MoveFlagPitchDown:       0x00000080,
			MoveFlagWalkMode:        0x00000100, // Walking
			MoveFlagOnTransport:     0x00000200, // Used for flying on some creatures
			MoveFlagLevitating:      0x00000400,
			MoveFlagRoot:            0x00000800,
			MoveFlagFalling:         0x00001000,
			MoveFlagFallingFar:      0x00004000,
			MoveFlagSwimming:        0x00200000, // appears with fly flag also
			MoveFlagAscending:       0x00400000, // swim up also
			MoveFlagCanFly:          0x00800000,
			MoveFlagFlyingOld:       0x01000000,
			MoveFlagFlying:          0x02000000, // Actual flying mode
			MoveFlagSplineElevation: 0x04000000, // used for flight paths
			MoveFlagSplineEnabled:   0x08000000, // used for flight paths
			MoveFlagWaterwalking:    0x10000000, // prevent unit from falling through water
			MoveFlagSafeFall:        0x20000000, // active rogue safe fall spell (passive)
			MoveFlagHover:           0x40000000,
		},

		vsn.V3_3_5a: {
			MoveFlagForward:         0x00000001,
			MoveFlagBackward:        0x00000002,
			MoveFlagStrafeLeft:      0x00000004,
			MoveFlagStrafeRight:     0x00000008,
			MoveFlagTurnLeft:        0x00000010,
			MoveFlagTurnRight:       0x00000020,
			MoveFlagPitchUp:         0x00000040,
			MoveFlagPitchDown:       0x00000080,
			MoveFlagWalkMode:        0x00000100, // Walking
			MoveFlagOnTransport:     0x00000200, // Used for flying on some creatures
			MoveFlagLevitating:      0x00000400,
			MoveFlagRoot:            0x00000800,
			MoveFlagFalling:         0x00001000,
			MoveFlagFallingFar:      0x00004000,
			MoveFlagSwimming:        0x00200000, // appears with fly flag also
			MoveFlagAscending:       0x00400000, // swim up also
			MoveFlagCanFly:          0x00800000,
			MoveFlagFlyingOld:       0x01000000,
			MoveFlagFlying:          0x02000000, // Actual flying mode
			MoveFlagSplineElevation: 0x04000000, // used for flight paths
			MoveFlagSplineEnabled:   0x08000000, // used for flight paths
			MoveFlagWaterwalking:    0x10000000, // prevent unit from falling through water
			MoveFlagSafeFall:        0x20000000, // active rogue safe fall spell (passive)
			MoveFlagHover:           0x40000000,
			// unused 0x80000000 (1 << 31), start of second flag
			MoveFlagNoStrafe:            0x10000000,
			MoveFlagFullSpeedTurning:    0x800000000,
			MoveFlagFullSpeedPitching:   0x1000000000,
			MoveFlagAllowPitching:       0x2000000000,
			MoveFlagInterpolateMovement: 0x40000000000,
			MoveFlagInterpolateTurning:  0x80000000000,
			MoveFlagInterpolatePitching: 0x100000000000,
		},
	}

	SpeedLists = map[vsn.Build][]SpeedType{
		vsn.Alpha: {
			Walk,
			Run,
			Swim,
			Turn,
		},

		vsn.V1_12_1: {
			Walk,
			Run,
			RunBackward,
			Swim,
			SwimBackward,
			Turn,
		},

		vsn.V2_4_3: {
			Walk,
			Run,
			RunBackward,
			Swim,
			SwimBackward,
			Flight,
			FlightBackward,
			Turn,
		},

		vsn.V3_3_5a: {
			Walk,
			Run,
			RunBackward,
			Swim,
			SwimBackward,
			Flight,
			FlightBackward,
			Turn,
			Pitch,
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

func parseFloat(s string) (float32, error) {
	f, err := strconv.ParseFloat(s, 32)
	return float32(f), err
}

func ParsePosition(in string) (Position, error) {
	strs := strings.Split(in, " ")

	var (
		pos Position
		err error
	)

	if len(strs) != 4 {
		return pos, fmt.Errorf("update: invalid Position string: only has %d coordinates", len(strs))
	}

	pos.X, err = parseFloat(strs[0])
	if err != nil {
		return pos, err
	}

	pos.Y, err = parseFloat(strs[1])
	if err != nil {
		return pos, err
	}
	pos.Z, err = parseFloat(strs[2])
	if err != nil {
		return pos, err
	}
	pos.O, err = parseFloat(strs[3])
	if err != nil {
		return pos, err
	}

	return pos, nil
}

type Speeds map[SpeedType]float32

type MovementBlock struct {
	ID              guid.GUID
	UpdateFlags     UpdateFlags
	Info            *MovementInfo
	Speeds          Speeds
	Spline          *MoveSpline
	Position        Position
	All             uint32
	LowGUID         uint32
	HighGUID        uint32
	Victim          guid.GUID
	WorldTime       uint32
	VehicleID       uint32
	VehicleRotation float32
	Rotation        Quaternion
}

func decodeUpdateFlags(version vsn.Build, in io.Reader) (UpdateFlags, error) {
	descriptor, ok := UpdateFlagDescriptors[version]
	if !ok {
		return 0, fmt.Errorf("update: no update flag descriptor for %d", version)
	}

	var updateFlags uint32
	if version.AddedIn(vsn.V3_3_5a) {
		uf, err := readUint16(in)
		if err != nil {
			return 0, err
		}
		updateFlags = uint32(uf)
	} else {
		uf, err := readUint8(in)
		if err != nil {
			return 0, err
		}

		updateFlags = uint32(uf)
	}

	out := UpdateFlags(0)

	// Map bits from version-dependent code to Gophercraft virtual bits
	for k, v := range descriptor {
		if updateFlags&v != 0 {
			out |= k
		}
	}

	return out, nil
}

func encodeUpdateFlags(version vsn.Build, outb io.Writer, uf UpdateFlags) error {
	descriptor, ok := UpdateFlagDescriptors[version]
	if !ok {
		return fmt.Errorf("update: no update flag descriptor for %d", version)
	}

	out := uint16(0)

	for k, v := range descriptor {
		if uf&k != 0 {
			out |= uint16(v)
		}
	}

	if version.AddedIn(vsn.V3_3_5a) {
		return writeUint16(outb, out)
	}

	return writeUint8(outb, uint8(out))
}

func decodeMoveFlags(version vsn.Build, in io.Reader) (MoveFlags, error) {
	descriptor, ok := MoveFlagDescriptors[version]
	if !ok {
		return 0, fmt.Errorf("update: no move flag descriptor for %d", version)
	}

	var raw [8]byte

	flagsSize := 4

	if version.AddedIn(vsn.V2_4_3) {
		if version.RemovedIn(vsn.V3_3_5a) {
			flagsSize += 1
		} else {
			flagsSize += 2
		}
	}

	_, err := in.Read(raw[:flagsSize])
	if err != nil {
		return 0, err
	}

	mf := binary.LittleEndian.Uint64(raw[:])

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

	out := uint64(0)

	for k, v := range descriptor {
		if mf&k != 0 {
			out |= v
		}
	}

	var data [8]byte
	binary.LittleEndian.PutUint64(data[:], out)

	flagsSize := 4

	// Byte flags
	if version.AddedIn(vsn.V2_4_3) && version.RemovedIn(vsn.V3_3_5a) {
		flagsSize = 5
	}

	// Uint16 flags
	if version.AddedIn(vsn.V3_3_5a) {
		flagsSize = 6
	}

	_, err := outb.Write(data[0:flagsSize])
	if err != nil {
		return err
	}

	return nil
}

func EncodeMovementInfo(version vsn.Build, out io.Writer, mi *MovementInfo) error {
	// Alpha uses a simpler format
	if version.RemovedIn(vsn.V1_12_1) {
		if mi == nil {
			mi = &MovementInfo{}
		}

		mi.TransportGUID.EncodeUnpacked(version, out)
		EncodePosition(out, mi.TransportPosition)
		EncodePosition(out, mi.Position)
		writeFloat32(out, mi.Pitch)
		err := encodeMoveFlags(version, out, mi.Flags)
		if err != nil {
			return err
		}
		return nil
	}

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
		if version.AddedIn(vsn.V2_4_3) {
			if err = writeUint32(out, mi.TransportTime); err != nil {
				return err
			}

			if version.AddedIn(vsn.V3_3_5a) {
				if err = writeUint8(out, uint8(mi.TransportSeat)); err != nil {
					return err
				}

				if mi.Flags&MoveFlagInterpolateMovement != 0 {
					if err = writeUint32(out, mi.TransportInterpolateTime); err != nil {
						return err
					}
				}
			}
		}
	}

	if mi.Flags&(MoveFlagSwimming|MoveFlagFlying|MoveFlagAllowPitching) != 0 {
		if err = writeFloat32(out, mi.Pitch); err != nil {
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
	// Alpha uses a very strange format.
	if version.RemovedIn(vsn.V1_12_1) {
		info.TransportGUID, err = guid.DecodeUnpacked(version, in)
		if err != nil {
			return nil, err
		}
		info.TransportPosition, err = DecodePosition(in)
		if err != nil {
			return nil, err
		}

		info.Position, err = DecodePosition(in)
		if err != nil {
			return nil, err
		}

		info.Pitch, err = readFloat32(in)
		if err != nil {
			return nil, err
		}

		info.Flags, err = decodeMoveFlags(version, in)
		if err != nil {
			return nil, err
		}

		return info, nil
	}

	info.Flags, err = decodeMoveFlags(version, in)
	if err != nil {
		return nil, err
	}

	// if version.AddedIn(vsn.V2_4_3) {
	// 	moveflags2, err := readUint8(in)
	// 	if err != nil {
	// 		return nil, err
	// 	}

	// 	if moveflags2 != 0 {
	// 		return nil, fmt.Errorf("update: extra moveflag has data4")
	// 	}
	// }

	info.Time, err = readUint32(in)
	if err != nil {
		return nil, err
	}
	info.Position, err = DecodePosition(in)
	if err != nil {
		return nil, err
	}

	if info.Flags&MoveFlagOnTransport != 0 {
		if version.AddedIn(vsn.V3_3_5a) {
			info.TransportGUID, err = guid.DecodePacked(version, in)
			if err != nil {
				return nil, err
			}
		} else {
			info.TransportGUID, err = guid.DecodeUnpacked(version, in)
			if err != nil {
				return nil, err
			}
		}

		info.TransportPosition, err = DecodePosition(in)
		if err != nil {
			return nil, err
		}
		if version.AddedIn(vsn.V2_4_3) {
			info.TransportTime, err = readUint32(in)
			if err != nil {
				return nil, err
			}

			if version.AddedIn(vsn.V3_3_5a) {
				ts, err := readUint8(in)
				if err != nil {
					return nil, err
				}

				info.TransportSeat = int8(ts)

				if info.Flags&MoveFlagInterpolateMovement != 0 {
					info.TransportInterpolateTime, err = readUint32(in)
					if err != nil {
						return nil, err
					}
				}
			}
		}
	}

	if info.Flags&(MoveFlagSwimming|MoveFlagFlying|MoveFlagAllowPitching) != 0 {
		info.Pitch, err = readFloat32(in)
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

	TransportGUID            guid.GUID
	TransportPosition        Position
	TransportTime            uint32
	TransportSeat            int8
	TransportInterpolateTime uint32

	Pitch        float32
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
	} else {
		if mBlock.UpdateFlags&UpdateFlagPosition != 0 {
			// Unknown guid
			_, err := guid.DecodePacked(decoder.Build, decoder.Reader)
			if err != nil {
				return nil, err
			}

			pos1, err := DecodePoint3(decoder.Reader)
			if err != nil {
				return nil, err
			}
			mBlock.Position.Point3 = pos1
			// Unknown what this second position does
			_, err = DecodePoint3(decoder.Reader)
			if err != nil {
				return nil, err
			}
			mBlock.Position.O, err = readFloat32(decoder.Reader)
			if err != nil {
				return nil, err
			}
			// Second orientation
			_, err = readFloat32(decoder.Reader)
			if err != nil {
				return nil, err
			}
		} else if mBlock.UpdateFlags&UpdateFlagHasPosition != 0 {
			mBlock.Position, err = DecodePosition(decoder.Reader)
			if err != nil {
				return nil, err
			}
		}
	}

	if mBlock.UpdateFlags&UpdateFlagLowGUID != 0 {
		mBlock.LowGUID, err = readUint32(decoder.Reader)
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

	if mBlock.UpdateFlags&UpdateFlagHasAttackingTarget != 0 {
		mBlock.Victim, err = guid.DecodePacked(decoder.Build, decoder.Reader)
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

	if mBlock.UpdateFlags&UpdateFlagVehicle != 0 {
		mBlock.VehicleID, err = readUint32(decoder.Reader)
		if err != nil {
			return nil, err
		}

		mBlock.VehicleRotation, err = readFloat32(decoder.Reader)
		if err != nil {
			return nil, err
		}
	}

	if mBlock.UpdateFlags&UpdateFlagRotation != 0 {
		mBlock.Rotation, err = DecodePackedQuaternion(decoder.Reader)
		if err != nil {
			return nil, err
		}
	}

	return mBlock, nil
}

func (mb *MovementBlock) WriteData(e *Encoder, mask VisibilityFlags, create bool) error {
	if e.Descriptor.DescriptorOptions&DescriptorOptionAlpha != 0 {
		return mb.writeDataAlpha(e)
	}

	if err := encodeUpdateFlags(e.Build, e, mb.UpdateFlags); err != nil {
		return err
	}

	if mb.UpdateFlags&UpdateFlagLiving != 0 {
		if mb.Info == nil {
			return fmt.Errorf("update: error serializing MovementBlock: UpdateFlagLiving is set but Info is nil")
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
	} else {
		if e.Build.AddedIn(vsn.V3_3_5a) {
			// Two options for setting position, it's unknown why this is.
			if mb.UpdateFlags&UpdateFlagPosition != 0 {
				if err := writeUint8(e, 0); err != nil {
					return err
				}

				// Also unknown why two positions get encoded.
				for x := 0; x < 2; x++ {
					if err := EncodePoint3(e, mb.Position.Point3); err != nil {
						return err
					}
				}

				for x := 0; x < 2; x++ {
					if err := writeFloat32(e, mb.Position.O); err != nil {
						return err
					}
				}
			} else {
				if mb.UpdateFlags&UpdateFlagHasPosition != 0 {
					EncodePosition(e, mb.Position)
				}
			}
		} else {
			// Only one option
			if mb.UpdateFlags&UpdateFlagHasPosition != 0 {
				EncodePosition(e, mb.Position)
			}
		}
	}

	if mb.UpdateFlags&UpdateFlagLowGUID != 0 {
		if err := writeUint32(e, mb.LowGUID); err != nil {
			return err
		}
	}

	if mb.UpdateFlags&UpdateFlagHighGUID != 0 {
		if err := writeUint32(e, mb.HighGUID); err != nil {
			return err
		}
	}

	if e.Build.RemovedIn(vsn.V2_4_3) {
		if mb.UpdateFlags&UpdateFlagAll != 0 {
			if err := writeUint32(e, mb.All); err != nil {
				return err
			}
		}
	}

	if mb.UpdateFlags&UpdateFlagHasAttackingTarget != 0 {
		mb.Victim.EncodePacked(e.Build, e)
	}

	if mb.UpdateFlags&UpdateFlagTransport != 0 {
		if err := writeUint32(e, mb.WorldTime); err != nil {
			return err
		}
	}

	if e.Build.AddedIn(vsn.V3_3_5a) {
		if mb.UpdateFlags&UpdateFlagVehicle != 0 {
			if err := writeUint32(e, mb.VehicleID); err != nil {
				return err
			}

			if err := writeFloat32(e, mb.VehicleRotation); err != nil {
				return err
			}
		}

		if mb.UpdateFlags&UpdateFlagRotation != 0 {
			if err := mb.Rotation.EncodePacked(e); err != nil {
				return err
			}
		}
	}

	return nil
}
