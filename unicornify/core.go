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

func (p Point3d) Minus(other Point3d) Point3d {
	return p.Shifted(other.Neg())
}

func (p Point3d) ScalarProd(v Point3d) float64 {
	return p[0]*v[0] + p[1]*v[1] + p[2]*v[2]
}

func (p Point3d) Unit() Point3d {
	return p.Times(1 / p.Length())
}

func (a Point3d) CrossProd(b Point3d) Point3d {
	return Point3d{
		a.Y()*b.Z()-a.Z()*b.Y(),
		a.Z()*b.X()-a.X()*b.Z(),
		a.X()*b.Y()-a.Y()*b.X(),
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

func MixBytes(b1, b2 byte, f float64) byte {
	return b1 + byte(f*(float64(b2)-float64(b1))+.5)
}

func MixFloats(f1, f2, f float64) float64 {
	return f1 + f*(f2-f1)
}

func IntersectionOfPlaneAndLine(p0, ep1, ep2, l0, el Point3d) (ok bool, intersection Point3d) {
	A := [3][3]float64{
		[3]float64{ep1.X(), ep2.X(), -el.X()},
		[3]float64{ep1.Y(), ep2.Y(), -el.Y()},
		[3]float64{ep1.Z(), ep2.Z(), -el.Z()},
	}
	b := l0.Minus(p0)
	// need to solve Ax = b where x = (fp1, fp2, fl)

	for i := 0; i <= 2; i++ {
		if A[i][i] != 0 {
			continue
		}
		for j := i + 1; j <= 2; j++ {
			if A[j][i] != 0 {
				A[i], A[j] = A[j], A[i]
				b[i], b[j] = b[j], b[i]
				break
			}
		}
	}

	for i := 0; i <= 1; i++ {
		for k := i + 1; k <= 2; k++ {
			A[k][i] = A[k][i] / A[i][i] //!
			for j := i + 1; j <= 2; j++ {
				A[k][j] = A[k][j] - A[k][i]*A[i][j]
			}
		}
	}

	y := Point3d{}
	for i := 0; i <= 2; i++ {
		y[i] = b[i]
		for k := 0; k <= i-1; k++ {
			y[i] -= A[i][k] * y[k]
		}
	}

	x := Point3d{}
	for i := 2; i >= 0; i-- {
		x[i] = y[i]
		for k := i + 1; k <= 2; k++ {
			x[i] -= A[i][k] * x[k]
		}
		x[i] /= A[i][i]
	}

	return true, x
}
