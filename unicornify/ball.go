package unicornify

import (
	"image"
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

func (b *Ball) GetTracer(img *image.RGBA, wv WorldView, shading bool) Tracer {
	b2 := *b
	b2.Projection[0] += 2 * b.ProjectionRadius
	result := NewBone(b, b).GetTracer(img, wv, shading)
	return result
}
func (b *Ball) Project(wv WorldView) {
	wv.ProjectBall(b)
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
