package update

import (
	"fmt"
	"io"
	"math"

	"github.com/superp00t/gophercraft/guid"
	"github.com/superp00t/gophercraft/vsn"
)

type SplineFlags uint32
type SplineFlagDescriptor map[SplineFlags]uint32

const (
	SplineNone SplineFlags = 1 << iota
	// x00-xFF(first byte) used as animation Ids storage in pair with Animation flag
	SplineDone
	SplineFalling // Affects elevation computation, can't be combined with Parabolic flag
	SplineNoSpline
	SplineParabolic // Affects elevation computation, can't be combined with Falling flag
	SplineCanSwim
	SplineFlying           // Smooth movement(Catmullrom interpolation mode), flying animation
	SplineOrientationFixed // Model ori    = 0x00008000,
	SplineFinalTarget
	SplineFinalPoint
	SplineFinalAngle
	SplineCatmullrom // Used Catmullrom interpolation mode
	SplineCyclic     // Movement by cycled spline
	SplineEnterCycle // Everytimes appears with cyclic flag in monster move packet, erases first spline vertex after first cycle done
	SplineAnimation  // Plays animation after some time passed
	SplineFrozen     // Will never arrive
	SplineTransportEnter
	SplineTransportExit
	SplineBackward
	SplineWalkmode
	SplineBoardVehicle
	SplineExitVehicle
	SplineOrientationInverted
)

var (
	SplineFlagDescriptors = map[vsn.Build]SplineFlagDescriptor{
		5875: {
			SplineDone:        0x00000001,
			SplineFalling:     0x00000002,
			SplineFlying:      0x00000200,
			SplineNoSpline:    0x00000400,
			SplineFinalPoint:  0x00010000,
			SplineFinalTarget: 0x00020000,
			SplineFinalAngle:  0x00040000,
			SplineCyclic:      0x00100000,
			SplineEnterCycle:  0x00200000,
			SplineFrozen:      0x00400000,
		},

		vsn.V3_3_5a: {
			// x00-xFF(first byte) used as animation Ids storage in pair with Animation flag
			SplineDone:                0x00000100,
			SplineFalling:             0x00000200, // Affects elevation computation, can't be combined with Parabolic flag
			SplineNoSpline:            0x00000400,
			SplineParabolic:           0x00000800, // Affects elevation computation, can't be combined with Falling flag
			SplineWalkmode:            0x00001000,
			SplineFlying:              0x00002000, // Smooth movement(Catmullrom interpolation mode), flying animation
			SplineOrientationFixed:    0x00004000, // Model orientation fixed
			SplineFinalPoint:          0x00008000,
			SplineFinalTarget:         0x00010000,
			SplineFinalAngle:          0x00020000,
			SplineCatmullrom:          0x00040000, // Used Catmullrom interpolation mode
			SplineCyclic:              0x00080000, // Movement by cycled spline
			SplineEnterCycle:          0x00100000, // Everytimes appears with cyclic flag in monster move packet, erases first spline vertex after first cycle done
			SplineAnimation:           0x00200000, // Plays animation after some time passed
			SplineFrozen:              0x00400000, // Will never arrive
			SplineBoardVehicle:        0x00800000,
			SplineExitVehicle:         0x01000000,
			SplineOrientationInverted: 0x08000000,
		},
	}
)

type MoveSpline struct {
	Flags                SplineFlags
	Facing               Point3
	FacingTarget         guid.GUID
	FacingAngle          float32
	TimePassed           int32
	Duration             int32
	ID                   uint32
	DurationMod          float32
	DurationModNext      float32
	VerticalAcceleration float32
	EffectStartTime      int32
	Spline               []Point3
	SplineMode           uint8
	Endpoint             Point3
}

type Point3 struct {
	X, Y, Z float32
}

func (p1 Point3) Dist2D(p2 Point3) float32 {
	x1 := float64(p1.X)
	y1 := float64(p1.Y)

	x2 := float64(p2.X)
	y2 := float64(p2.Y)

	dist := math.Sqrt(
		math.Pow(x2-x1, 2) +
			math.Pow(y2-y1, 2))

	return float32(dist)
}

func (p1 Point3) Dist3D(p2 Point3) float32 {
	x1 := float64(p1.X)
	y1 := float64(p1.Y)
	z1 := float64(p1.Z)

	x2 := float64(p2.X)
	y2 := float64(p2.Y)
	z2 := float64(p2.Z)

	dist := math.Sqrt(
		math.Pow(x2-x1, 2) +
			math.Pow(y2-y1, 2) +
			math.Pow(z2-z1, 2))

	return float32(dist)
}

func EncodePoint3(out io.Writer, p3 Point3) error {
	if err := writeFloat32(out, p3.X); err != nil {
		return err
	}
	if err := writeFloat32(out, p3.Y); err != nil {
		return err
	}
	err := writeFloat32(out, p3.Z)
	return err
}

func DecodePoint3(in io.Reader) (Point3, error) {
	var err error
	var p3 Point3
	p3.X, err = readFloat32(in)
	if err != nil {
		return p3, err
	}
	p3.Y, err = readFloat32(in)
	if err != nil {
		return p3, err
	}
	p3.Z, err = readFloat32(in)
	if err != nil {
		return p3, err
	}
	return p3, nil
}

func decodeSplineFlags(version vsn.Build, in io.Reader) (SplineFlags, error) {
	descriptor, ok := SplineFlagDescriptors[version]
	if !ok {
		return 0, fmt.Errorf("update: unsupported spline version %d", version)
	}

	sf, err := readUint32(in)
	if err != nil {
		return 0, err
	}

	out := SplineFlags(0)
	// translate packet bits to virtual Gophercraft bits
	for k, v := range descriptor {
		if sf&v != 0 {
			out |= k
		}
	}

	return out, nil
}

func encodeSplineFlags(version vsn.Build, out io.Writer, sf SplineFlags) error {
	descriptor, ok := SplineFlagDescriptors[version]
	if !ok {
		return fmt.Errorf("update: unsupported spline version %d", version)
	}

	u32 := uint32(0)

	for k, v := range descriptor {
		if sf&k != 0 {
			u32 |= v
		}
	}

	return writeUint32(out, u32)
}

func DecodeMoveSpline(version vsn.Build, in io.Reader) (*MoveSpline, error) {
	ms := new(MoveSpline)
	var err error
	ms.Flags, err = decodeSplineFlags(version, in)
	if err != nil {
		return nil, err
	}

	// Flag order reversed
	if version.AddedIn(vsn.V2_4_3) {
		if ms.Flags&SplineFinalAngle != 0 {
			ms.FacingAngle, err = readFloat32(in)
			if err != nil {
				return nil, err
			}

		} else if ms.Flags&SplineFinalTarget != 0 {
			ms.FacingTarget, err = guid.DecodeUnpacked(version, in)
			if err != nil {
				return nil, err
			}
		} else if ms.Flags&SplineFinalPoint != 0 {
			ms.Facing, err = DecodePoint3(in)
			if err != nil {
				return nil, err
			}
		}
	} else {
		if ms.Flags&SplineFinalPoint != 0 {
			ms.Facing, err = DecodePoint3(in)
			if err != nil {
				return nil, err
			}
		} else if ms.Flags&SplineFinalTarget != 0 {
			ms.FacingTarget, err = guid.DecodeUnpacked(version, in)
			if err != nil {
				return nil, err
			}
		} else if ms.Flags&SplineFinalAngle != 0 {
			ms.FacingAngle, err = readFloat32(in)
			if err != nil {
				return nil, err
			}
		}
	}

	ms.TimePassed, err = readInt32(in)
	if err != nil {
		return nil, err
	}
	ms.Duration, err = readInt32(in)
	if err != nil {
		return nil, err
	}
	ms.ID, err = readUint32(in)
	if err != nil {
		return nil, err
	}

	if version.AddedIn(vsn.V3_3_5a) {
		ms.DurationMod, err = readFloat32(in)
		if err != nil {
			return nil, err
		}
		ms.DurationModNext, err = readFloat32(in)
		if err != nil {
			return nil, err
		}
		ms.VerticalAcceleration, err = readFloat32(in)
		if err != nil {
			return nil, err
		}
		ms.EffectStartTime, err = readInt32(in)
		if err != nil {
			return nil, err
		}
	}

	nodeLength, err := readInt32(in)
	if err != nil {
		return nil, err
	}

	if nodeLength > 0xFFFF {
		return nil, fmt.Errorf("spline overread")
	}

	for i := int32(0); i < nodeLength; i++ {
		p3, err := DecodePoint3(in)
		if err != nil {
			return nil, err
		}
		ms.Spline = append(ms.Spline, p3)
	}

	if version.AddedIn(vsn.V3_3_5a) {
		ms.SplineMode, err = readUint8(in)
		if err != nil {
			return nil, err
		}
	}

	ms.Endpoint, err = DecodePoint3(in)
	return ms, err
}

func EncodeMoveSpline(version vsn.Build, out io.Writer, ms *MoveSpline) error {
	if err := encodeSplineFlags(version, out, ms.Flags); err != nil {
		return err
	}

	if ms.Flags&SplineFinalPoint != 0 {
		if err := EncodePoint3(out, ms.Facing); err != nil {
			return err
		}
	} else if ms.Flags&SplineFinalTarget != 0 {
		ms.FacingTarget.EncodeUnpacked(version, out)
	} else if ms.Flags&SplineFinalAngle != 0 {
		writeFloat32(out, ms.FacingAngle)
	}

	writeInt32(out, ms.TimePassed)
	writeInt32(out, ms.Duration)
	writeUint32(out, ms.ID)

	if version.AddedIn(vsn.V3_3_5a) {
		writeFloat32(out, ms.DurationMod)
		writeFloat32(out, ms.DurationModNext)
		writeFloat32(out, ms.VerticalAcceleration)
		writeInt32(out, ms.EffectStartTime)
	}

	writeUint32(out, uint32(len(ms.Spline)))

	for _, p3 := range ms.Spline {
		EncodePoint3(out, p3)
	}

	if version.AddedIn(vsn.V3_3_5a) {
		writeUint8(out, ms.SplineMode)
	}

	EncodePoint3(out, ms.Endpoint)

	return nil
}
