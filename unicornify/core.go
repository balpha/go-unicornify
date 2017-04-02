package unicornify

import (
	"math"
)

const pi = math.Pi
const deg360 = 2 * pi

type Vector [3]float64

type Point2d [2]float64

func (p Vector) X() float64 {
	return p[0]
}

func (p Vector) Y() float64 {
	return p[1]
}

func (p Vector) Z() float64 {
	return p[2]
}

func (p Vector) Shifted(delta Vector) Vector {
	return Vector{
		p[0] + delta[0],
		p[1] + delta[1],
		p[2] + delta[2],
	}
}

func (p Vector) Length() float64 {
	return math.Sqrt(p[0]*p[0] + p[1]*p[1] + p[2]*p[2])
}

func (p Vector) Times(v float64) Vector {
	return Vector{
		v * p[0],
		v * p[1],
		v * p[2],
	}
}

func (p Vector) Minus(other Vector) Vector {
	return p.Shifted(other.Neg())
}

func (p Vector) ScalarProd(v Vector) float64 {
	return p[0]*v[0] + p[1]*v[1] + p[2]*v[2]
}

func (p Vector) Unit() Vector {
	l := 1 / p.Length()
	return Vector{p[0] * l, p[1] * l, p[2] * l}
}

func (a Vector) CrossProd(b Vector) Vector {
	return Vector{
		a.Y()*b.Z() - a.Z()*b.Y(),
		a.Z()*b.X() - a.X()*b.Z(),
		a.X()*b.Y() - a.Y()*b.X(),
	}
}

func (p Vector) Neg() Vector {
	return Vector{
		-p[0],
		-p[1],
		-p[2],
	}
}

func (p Vector) DiscardZ() Point2d {
	return Point2d{
		p[0],
		p[1],
	}
}

func (p Vector) Decompose() (x, y, z float64) {
	return p[0], p[1], p[2]
}

func (p Vector) RotatedAround(other Vector, angle float64, axis byte) Vector {
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

	return Vector{rotated[reverse[0]], rotated[reverse[1]], rotated[reverse[2]]}.Shifted(other)
}

func MixBytes(b1, b2 byte, f float64) byte {
	return b1 + byte(f*(float64(b2)-float64(b1))+.5)
}

func MixFloats(f1, f2, f float64) float64 {
	return f1 + f*(f2-f1)
}

func intersectionOfPlaneAndLineReadable(p0, ep1, ep2, l0, el Vector) (ok bool, intersection Vector) {
	A := [3][3]float64{
		[3]float64{ep1.X(), ep2.X(), -el.X()},
		[3]float64{ep1.Y(), ep2.Y(), -el.Y()},
		[3]float64{ep1.Z(), ep2.Z(), -el.Z()},
	}
	b := l0.Minus(p0)
	// need to solve Ax = b where x = (fp1, fp2, fl)
	for i := 0; i <= 2; i++ {
		maxabs := 0.0
		maxi := -1
		for j := i; j <= 2; j++ {
			thisabs := math.Abs(A[j][i])
			if thisabs > maxabs {
				maxi = j
				maxabs = thisabs
			}
		}
		if maxi > -1 && maxi != i {
			A[i], A[maxi] = A[maxi], A[i]
			b[i], b[maxi] = b[maxi], b[i]

		}
		if A[i][i] == 0 {
			return false, Vector{}
		}
		for k := i + 1; k <= 2; k++ {
			A[k][i] = A[k][i] / A[i][i]
			for j := i + 1; j <= 2; j++ {
				A[k][j] = A[k][j] - A[k][i]*A[i][j]
			}
		}
	}

	y := Vector{}
	for i := 0; i <= 2; i++ {
		y[i] = b[i]
		for k := 0; k <= i-1; k++ {
			y[i] -= A[i][k] * y[k]
		}
	}

	x := Vector{}
	for i := 2; i >= 0; i-- {
		x[i] = y[i]
		for k := i + 1; k <= 2; k++ {
			x[i] -= A[i][k] * x[k]
		}
		x[i] /= A[i][i]
	}

	return true, x
}

func IntersectionOfPlaneAndLine(p0, ep1, ep2, l0, el Vector) (ok bool, intersection Vector) {
	A00, A01, A02, A10, A11, A12, A20, A21, A22 := ep1.X(), ep2.X(), -el.X(), ep1.Y(), ep2.Y(), -el.Y(), ep1.Z(), ep2.Z(), -el.Z()
	b0, b1, b2 := l0.Minus(p0).Decompose()
	// need to solve Ax = b where x = (fp1, fp2, fl)

	/*
		for i := 0; i <= 2; i++ {
			maxabs := 0.0
			maxi := -1
			for j := i; j <= 2; j++ {
				thisabs := math.Abs(A[j][i])
				if thisabs > maxabs {
					maxi = j
					maxabs = thisabs
				}
			}
			if maxi > -1 && maxi != i {
				A[i], A[maxi] = A[maxi], A[i]
				b[i], b[maxi] = b[maxi], b[i]

			}
			if A[i][i] == 0 {
				return false, Vector{}
			}
			for k := i + 1; k <= 2; k++ {
				A[k][i] = A[k][i] / A[i][i]
				for j := i + 1; j <= 2; j++ {
					A[k][j] = A[k][j] - A[k][i]*A[i][j]
				}
			}
		}*/

	//i = 0
	abs0 := math.Abs(A00)
	abs1 := math.Abs(A10)
	abs2 := math.Abs(A20)
	if abs1 > abs0 || abs2 > abs0 {
		if abs1 > abs2 {
			A00, A01, A02, A10, A11, A12 = A10, A11, A12, A00, A01, A02
			b0, b1 = b1, b0
		} else {
			A00, A01, A02, A20, A21, A22 = A20, A21, A22, A00, A01, A02
			b0, b2 = b2, b0
		}
	}
	if A00 == 0 {
		return false, Vector{}
	}
	//   k=1
	A10 /= A00
	//     j=1
	A11 -= A10 * A01
	//     j=2
	A12 -= A10 * A02
	//   k=2
	A20 /= A00
	//     j=1
	A21 -= A20 * A01
	//     j=2
	A22 -= A20 * A02

	//i=1
	abs1 = math.Abs(A11)
	abs2 = math.Abs(A21)
	if abs2 > abs1 {
		A10, A11, A12, A20, A21, A22 = A20, A21, A22, A10, A11, A12
		b1, b2 = b2, b1
	}
	if A11 == 0 {
		return false, Vector{}
	}
	//   k=2
	A21 /= A11
	//     j=2
	A22 -= A21 * A12

	//i=2
	if A22 == 0 {
		return false, Vector{}
	}

	// ---------------------------------------------

	/*y := Vector{}
	for i := 0; i <= 2; i++ {
		y[i] = b[i]
		for k := 0; k <= i-1; k++ {
			y[i] -= A[i][k] * y[k]
		}
	}*/

	b1 -= A10 * b0
	b2 -= A20*b0 + A21*b1

	// -------------------------------------

	/*x := Vector{}
	for i := 2; i >= 0; i-- {
		x[i] = y[i]
		for k := i + 1; k <= 2; k++ {
			x[i] -= A[i][k] * x[k]
		}
		x[i] /= A[i][i]
	}*/

	b2 /= A22
	b1 = (b1 - A12*b2) / A11
	b0 = (b0 - (A01*b1 + A02*b2)) / A00

	return true, Vector{b0, b1, b2}
}
