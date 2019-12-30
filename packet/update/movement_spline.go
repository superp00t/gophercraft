package update

import (
	"fmt"
	"math"

	"github.com/superp00t/etc"
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
	SplineFlagDescriptors = map[uint32]SplineFlagDescriptor{
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
	FacingTarget uint64
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

func EncodePoint3(out *etc.Buffer, p3 Point3) {
	out.WriteFloat32(p3.X)
	out.WriteFloat32(p3.Y)
	out.WriteFloat32(p3.Z)
}

func DecodePoint3(in *etc.Buffer) Point3 {
	p3 := Point3{}
	p3.X = in.ReadFloat32()
	p3.Y = in.ReadFloat32()
	p3.Z = in.ReadFloat32()
	return p3
}

func decodeSplineFlags(version uint32, in *etc.Buffer) (SplineFlags, error) {
	descriptor, ok := SplineFlagDescriptors[version]
	if !ok {
		return 0, fmt.Errorf("update: unsupported spline version %d", version)
	}

	sf := in.ReadUint32()

	out := SplineFlags(0)
	// translate packet bits to virtual Gophercraft bits
	for k, v := range descriptor {
		if sf&v != 0 {
			out |= k
		}
	}

	return out, nil
}

func encodeSplineFlags(version uint32, out *etc.Buffer, sf SplineFlags) error {
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

	out.WriteUint32(u32)
	return nil
}

func DecodeMoveSpline(version uint32, in *etc.Buffer) (*MoveSpline, error) {
	ms := new(MoveSpline)
	var err error
	ms.Flags, err = decodeSplineFlags(version, in)
	if err != nil {
		return nil, err
	}

	if ms.Flags&SplineFinalPoint != 0 {
		ms.Facing = DecodePoint3(in)
	} else if ms.Flags&SplineFinalTarget != 0 {
		ms.FacingTarget = in.ReadUint64()
	} else if ms.Flags&SplineFinalAngle != 0 {
		ms.FacingAngle = in.ReadFloat32()
	}

	ms.TimePassed = in.ReadInt32()
	ms.Duration = in.ReadInt32()
	ms.ID = in.ReadUint32()

	nodeLength := in.ReadInt32()
	if int64(nodeLength) > (4 * (in.Size() - in.Rpos())) {
		return nil, fmt.Errorf("update: spline length exceeded size of input buffer. (%d)", nodeLength)
	}

	for i := int32(0); i < nodeLength; i++ {
		ms.Spline = append(ms.Spline, DecodePoint3(in))
	}

	ms.Endpoint = DecodePoint3(in)

	return ms, nil
}

func EncodeMoveSpline(version uint32, out *etc.Buffer, ms *MoveSpline) error {
	if err := encodeSplineFlags(version, out, ms.Flags); err != nil {
		return err
	}

	if ms.Flags&SplineFinalPoint != 0 {
		EncodePoint3(out, ms.Facing)
	} else if ms.Flags&SplineFinalTarget != 0 {
		out.WriteUint64(ms.FacingTarget)
	} else if ms.Flags&SplineFinalAngle != 0 {
		out.WriteFloat32(ms.FacingAngle)
	}

	out.WriteInt32(ms.TimePassed)
	out.WriteInt32(ms.Duration)
	out.WriteUint32(ms.ID)

	out.WriteUint32(uint32(len(ms.Spline)))

	for _, p3 := range ms.Spline {
		EncodePoint3(out, p3)
	}

	EncodePoint3(out, ms.Endpoint)

	return nil
}
