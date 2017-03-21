package unicornify

import (
	"math"
)

type WorldView interface {
	ProjectBall(*Ball)
	Shifted(x, y float64) (float64, float64)
}

func ShiftedProjection(wv WorldView, p Point3d) (float64, float64) {
	return wv.Shifted(p.X(), p.Y())
}

type ParallelWorldView struct {
	AngleX, AngleY float64
	RotationCenter Point3d
	Shift          Point2d
	Scale          float64
}

func (wv ParallelWorldView) ProjectBall(b *Ball) {
	x1, y1, z1 := b.Center.Shifted(wv.RotationCenter.Neg()).Decompose()

	x2 := x1*math.Cos(wv.AngleY) - z1*math.Sin(wv.AngleY)
	y2 := y1
	z2 := x1*math.Sin(wv.AngleY) + z1*math.Cos(wv.AngleY)

	x3 := x2
	y3 := y2*math.Cos(wv.AngleX) - z2*math.Sin(wv.AngleX)
	z3 := y2*math.Sin(wv.AngleX) + z2*math.Cos(wv.AngleX)

	b.Projection = Point3d{x3, y3, z3}.Times(wv.Scale).Shifted(wv.RotationCenter)
	b.ProjectionRadius = b.Radius * wv.Scale
}

func (wv ParallelWorldView) Shifted(x, y float64) (float64, float64) {
	return x + wv.Shift[0], y + wv.Shift[1]
}

type PerspectiveWorldView struct {
	CameraPosition Point3d
	LookAtPoint    Point3d
	Shift          Point2d
	Scale          float64
	FocalLength    float64
}

func (wv PerspectiveWorldView) Shifted(x, y float64) (float64, float64) {
	return x + wv.Shift[0], y + wv.Shift[1]
}

func (wv PerspectiveWorldView) ProjectBall(b *Ball) {
	cam2c := b.Center.Shifted(wv.CameraPosition.Neg())
	view := wv.LookAtPoint.Shifted(wv.CameraPosition.Neg())
	n := view.Times(1.0 / view.Length())
	n1, n2, n3 := n.Decompose()

	var x1, x3 float64

	if n1 != 0 {
		x3 = math.Sqrt(1 / (n3*n3/(n1*n1) + 1))
		if n1 > 0 {
			x3 = -x3
		}
		x1 = -x3 * n3 / n1
	} else if n3 != 0 {
		x1 = math.Sqrt(1 / (n1*n1/(n3*n3) + 1))
		if n3 < 0 {
			x1 = -x1
		}
		x3 = -x1 * n1 / n3
	} else { // both 0 -- looking down
		x1 = 1
		x3 = 0
	}

	ux := Point3d{x1, 0, x3}

	// cross product of ux and normal (=uz) gives the y axis but in the wrong direction
	// (because x-z-y is not a right-hand rule system)
	y1 := -(-x3 * n2)
	y2 := -(x3*n1 - x1*n3)
	y3 := -(x1 * n2)

	uy := Point3d{y1, y2, y3}

	ok, intf := IntersectionOfPlaneAndLine(wv.CameraPosition.Shifted(n.Times(wv.FocalLength)), ux, uy, wv.CameraPosition, b.Center.Minus(wv.CameraPosition))
	if !ok { //FIXME
		b.Projection = Point3d{}
		b.ProjectionRadius = b.Radius
	} else {
		b.Projection = Point3d{intf[0], intf[1], cam2c.Length()}.Times(wv.Scale)
		count := 0.0
		max := 0.0
		for dx := -1.0; dx <= 1; dx += 1 {
			for dy := -1.0; dy <= 1; dy += 1 {
				for dz := -1.0; dz <= 1; dz += 1 {
					shift := Point3d{dx, dy, dz}
					if shift.Length() == 0 {
						continue
					}
					shift = shift.Times(b.Radius / shift.Length())
					ok, intf := IntersectionOfPlaneAndLine(wv.CameraPosition.Shifted(n.Times(wv.FocalLength)), ux, uy, wv.CameraPosition, b.Center.Shifted(shift).Minus(wv.CameraPosition))
					if ok {
						count++
						rp := Point3d{intf[0], intf[1], 0}.Times(wv.Scale)
						max = math.Max(max, math.Sqrt(sqr(rp[0]-b.Projection[0])+sqr(rp[1]-b.Projection[1])))
					}
				}
			}

		}

		b.ProjectionRadius = max
	}
	return

}
