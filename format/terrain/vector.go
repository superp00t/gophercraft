package terrain

type C3Vector [3]float32

type CAaSphere struct {
	Position C3Vector
	Radius   float32
}

type CAaBox struct {
	Min C3Vector
	Max C3Vector
}
