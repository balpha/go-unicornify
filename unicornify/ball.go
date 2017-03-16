package unicornify

import (
	"image"
	"math"
)

type Ball struct {
	Center           Point3d
	Radius           float64
	Color            Color
	Projection       Point3d
	ProjectionRadius float64
}

func NewBall(x, y, z, r float64, c Color) *Ball {
	return NewBallP(Point3d{x, y, z}, r, c)
}
func NewBallP(center Point3d, r float64, c Color) *Ball {
	return &Ball{
		Center: center,
		Radius: r,
		Color:  c,
	}
}

func (b Ball) Draw(img *image.RGBA, wv WorldView, shading bool) {
	sh := 0.25
	if !shading {
		sh = 0
	}
	CircleF(img, b.Projection.X()+wv.Shift[0], b.Projection.Y()+wv.Shift[1], b.ProjectionRadius, b.Color, DefaultGradientWithShading(sh))
}

func (b *Ball) Project(wv WorldView) {
	x1, y1, z1 := b.Center.Shifted(wv.RotationCenter.Neg()).Decompose()

	x2 := x1*math.Cos(wv.AngleY) - z1*math.Sin(wv.AngleY)
	y2 := y1
	z2 := x1*math.Sin(wv.AngleY) + z1*math.Cos(wv.AngleY)

	x3 := x2
	y3 := y2*math.Cos(wv.AngleX) - z2*math.Sin(wv.AngleX)
	z3 := y2*math.Sin(wv.AngleX) + z2*math.Cos(wv.AngleX)

	b.Projection = Point3d{x3, y3, z3}.Shifted(wv.RotationCenter)
	b.ProjectionRadius = b.Radius
}

func (b *Ball) SetDistance(distance float64, other Ball) {
	span := b.Center.Shifted(other.Center.Neg())
	b.Center = other.Center.Shifted(span.Times(distance / span.Length()))
}

func (b *Ball) RotateAround(other Ball, angle float64, axis byte) {
	b.Center = b.Center.RotatedAround(other.Center, angle, axis)
}

func (b *Ball) MoveToSphere(other Ball) {
	b.SetDistance(other.Radius, other)
}

func (b *Ball) SetGap(gap float64, other Ball) {
	b.SetDistance(b.Radius+other.Radius+gap, other)
}

func (b Ball) Bounding() image.Rectangle {
	x, y, _ := b.Projection.Decompose()
	r := b.ProjectionRadius
	return image.Rect(int(x-r), int(y-r), int(x+r+2), int(y+r+1))
}

func (b *Ball) Sort(wv WorldView) {
	return
}
