package unicornify

import (
	"image"
)

type Ball struct {
	Center           Point3d
	Radius           float64
	Color            Color
	Projection       BallProjection
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

func (b *Ball) GetTracer(wv WorldView) Tracer {
	result := NewBone(b, b).GetTracer(wv)
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

func (b *Ball) MoveToBone(bone Bone) {
	b1 := bone.Balls[0]
	b2 := bone.Balls[1]
	span := b2.Center.Shifted(b1.Center.Neg())
	bs := b.Center.Shifted(b1.Center.Neg())
	f := span.ScalarProd(bs) / (span.Length() * bs.Length())
	if f <= 0 {
		b.MoveToSphere(*b1)
	} else if f >= 1 {
		b.MoveToSphere(*b2)
	} else {
		ibc := b1.Center.Shifted(span.Times(f))
		ib := NewBall(ibc.X(), ibc.Y(), ibc.Z(), b1.Radius+f*(b2.Radius-b1.Radius), Color{})
		b.MoveToSphere(*ib)

	}
}

func (b Ball) Bounding() image.Rectangle {
	x, y := b.Projection.X(), b.Projection.Y()
	r := b.Projection.Radius
	return image.Rect(int(x-r), int(y-r), int(x+r+2), int(y+r+1))
}

type BallProjection struct {
	CenterCS Point3d // the center in camera space (camera at (0,0,0), Z axis in view direction)
	ProjectedCenterCS Point3d // the projection in camera space (note that by definition, Z will always be = focal length)
	ProjectedCenterOS Point3d // the projection in original space
	Radius float64
}

func (bp BallProjection) X() float64 {
	return bp.ProjectedCenterCS.X()
}

func (bp BallProjection) Y() float64 {
	return bp.ProjectedCenterCS.Y()
}

func (bp BallProjection) Z() float64 {
	return bp.CenterCS.Length()
}

