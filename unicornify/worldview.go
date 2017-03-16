package unicornify

import (
	"math"
)

type WorldView interface {
	ProjectBall(*Ball)
	Shifted(x, y float64) (float64, float64)
}

type OrthogonalWorldView struct {
	AngleX, AngleY float64
	RotationCenter Point3d
	Shift          Point2d
	Scale          float64
}

func (wv OrthogonalWorldView) ProjectBall(b *Ball) {
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

func (wv OrthogonalWorldView) Shifted(x, y float64) (float64, float64) {
	return x + wv.Shift[0], y + wv.Shift[1]
}

func ShiftedProjection(wv WorldView, p Point3d) (float64, float64) {
	return wv.Shifted(p.X(), p.Y())
}
