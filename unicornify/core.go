package unicornify

import (
	"math"
)

const pi = math.Pi
const deg360 = 2 * pi

type Point3d [3]float64

type Point2d [2]float64

func (p Point3d) X() float64 {
	return p[0]
}

func (p Point3d) Y() float64 {
	return p[1]
}

func (p Point3d) Z() float64 {
	return p[2]
}

func (p Point3d) Shifted(delta Point3d) Point3d {
	return Point3d{
		p[0] + delta[0],
		p[1] + delta[1],
		p[2] + delta[2],
	}
}

func (p Point3d) Length() float64 {
	return math.Sqrt(p[0]*p[0] + p[1]*p[1] + p[2]*p[2])
}

func (p Point3d) Times(v float64) Point3d {
	return Point3d{
		v * p[0],
		v * p[1],
		v * p[2],
	}
}

func (p Point3d) Neg() Point3d {
	return Point3d{
		-p[0],
		-p[1],
		-p[2],
	}
}

func (p Point3d) DiscardZ() Point2d {
	return Point2d{
		p[0],
		p[1],
	}
}

func (p Point3d) Decompose() (x, y, z float64) {
	return p[0], p[1], p[2]
}

func (p Point3d) RotatedAround(other Point3d, angle float64, axis byte) Point3d {
	var swap, reverse [3]byte
	switch axis {
	case 0:
		swap = [3]byte{1, 2, 0}
		reverse = [3]byte{2, 0, 1}
	case 1:
		swap = [3]byte{0, 2, 1}
		reverse = [3]byte{0, 2, 1}
	case 2:
		swap = [3]byte{0, 1, 2}
		reverse = [3]byte{0, 1, 2}
	}

	shifted := p.Shifted(other.Neg())

	// the letters x, y, z are the correct ones for the case axis = 2
	x1, y1, z1 := shifted[swap[0]], shifted[swap[1]], shifted[swap[2]]
	var rotated [3]float64
	rotated[0] = x1*math.Cos(angle) - y1*math.Sin(angle)
	rotated[1] = x1*math.Sin(angle) + y1*math.Cos(angle)
	rotated[2] = z1

	return Point3d{rotated[reverse[0]], rotated[reverse[1]], rotated[reverse[2]]}.Shifted(other)
}

type WorldView struct {
	AngleX, AngleY float64
	RotationCenter Point3d
	Shift          Point2d
}

func MixBytes(b1, b2 byte, f float64) byte {
	return b1 + byte(f*(float64(b2)-float64(b1))+.5)
}

func MixFloats(f1, f2, f float64) float64 {
	return f1 + f*(f2-f1)
}
