package elements

import (
	. "bitbucket.org/balpha/go-unicornify/unicornify/core"
)

type Ball struct {
	Center Vector
	Radius float64
	Color  Color
}

func NewBall(x, y, z, r float64, c Color) *Ball {
	return NewBallP(Vector{x, y, z}, r, c)
}
func NewBallP(center Vector, r float64, c Color) *Ball {
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

func (b *Ball) SetDistance(distance float64, other Ball) {
	span := b.Center.Plus(other.Center.Neg())
	b.Center = other.Center.Plus(span.Times(distance / span.Length()))
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

func (b *Ball) Shift(d Vector) {
	b.Center = b.Center.Plus(d)
}

func (b *Ball) MoveToBone(bone Bone) {
	b1 := bone.Balls[0]
	b2 := bone.Balls[1]
	span := b2.Center.Plus(b1.Center.Neg())
	bs := b.Center.Plus(b1.Center.Neg())
	f := span.ScalarProd(bs) / (span.Length() * bs.Length())
	if f <= 0 {
		b.MoveToSphere(*b1)
	} else if f >= 1 {
		b.MoveToSphere(*b2)
	} else {
		ibc := b1.Center.Plus(span.Times(f))
		ib := NewBall(ibc.X(), ibc.Y(), ibc.Z(), b1.Radius+f*(b2.Radius-b1.Radius), Color{})
		b.MoveToSphere(*ib)

	}
}

type BallProjection struct {
	SphereProjection
	BaseBall Ball // yeah sorry, but that's exactly what it is (btw, note this isn't a pointer)
}

func RenderingBoundsForBalls(bps ...BallProjection) Bounds {
	res := EmptyBounds
	for _, bp := range bps {
		x, y := bp.X(), bp.Y()
		if bp.CenterCS.Z() < 0 {
			x, y = bp.CenterCS.X(), bp.CenterCS.Y()
		}
		r := Bounds{
			XMin:  x - bp.ProjectedRadius,
			XMax:  x + bp.ProjectedRadius,
			YMin:  y - bp.ProjectedRadius,
			YMax:  y + bp.ProjectedRadius,
			ZMin:  bp.CenterCS.Z() - bp.BaseBall.Radius,
			ZMax:  bp.Z() + bp.BaseBall.Radius,
			Empty: false,
		}
		res = res.Union(r)
	}
	return res
}

func ProjectBall(wv WorldView, b *Ball) BallProjection {
	return BallProjection{
		SphereProjection: wv.ProjectSphere(b.Center, b.Radius),
		BaseBall:         *b,
	}
}
