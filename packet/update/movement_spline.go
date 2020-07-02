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
	SplineUnknown7
	SplineUnknown8
	SplineBackward
	SplineUnknown10
	SplineUnknown11
	SplineUnknown12
	SplineUnknown13
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
	}
)

type MoveSpline struct {
	Flags        SplineFlags
	Facing       Point3
	FacingTarget guid.GUID
	FacingAngle  float32
	TimePassed   int32
	Duration     int32
	ID           uint32
	Spline       []Point3
	Endpoint     Point3
}

type Point3 struct {
	X, Y, Z float32
}

func (p1 Point3) Dist2D(p2 Point3) float32 {
	x1 := float64(p1.X)
	y1 := float64(p1.Y)

	x2 := float64(p2.X)
	y2 := float64(p2.Y)

	dist := math.Sqrt(math.Pow(x2-x1, 2) + math.Pow(y2-y1, 2))

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

	writeUint32(out, uint32(len(ms.Spline)))

	for _, p3 := range ms.Spline {
		EncodePoint3(out, p3)
	}

	EncodePoint3(out, ms.Endpoint)

	return nil
}
