package elements

import (
	. "bitbucket.org/balpha/go-unicornify/unicornify/core"
)

type FlatTracer struct {
	p1, p2, p3  BallProjection
	bounds      Bounds
	fourCorners bool
	w1, w2      Vector
	dir         Vector
	fourthColor Color
	wv          WorldView
}

func NewFlatTracer(wv WorldView, b1, b2, b3 *Ball, fourCorners bool, fourthColor Color, roughDirection Vector) *FlatTracer {
	t := &FlatTracer{
		p1:          ProjectBall(wv, b1),
		p2:          ProjectBall(wv, b2),
		p3:          ProjectBall(wv, b3),
		wv:          wv,
		fourCorners: fourCorners,
	}
	t.w1 = t.p2.CenterCS.Minus(t.p1.CenterCS)
	t.w2 = t.p3.CenterCS.Minus(t.p1.CenterCS)

	t.dir = t.w1.CrossProd(t.w2)
	p := Vector{0, 0, 10000} // FIXME
	pp := wv.ProjectSphere(p, 0).CenterCS
	roughDirCS := wv.ProjectSphere(p.Plus(roughDirection), 0).CenterCS.Minus(pp)
	if t.dir.ScalarProd(roughDirCS) < 0 {
		t.dir = t.dir.Neg()
	}
	bounds := RenderingBoundsForBalls(t.p1, t.p2, t.p3)
	if fourCorners {
		t.fourthColor = fourthColor
		b4 := NewBallP(b1.Center.Plus(b2.Center.Minus(b1.Center)).Plus(b3.Center.Minus(b1.Center)), 0, Color{})
		p4 := ProjectBall(wv, b4)
		bounds = bounds.Union(RenderingBoundsForBalls(p4))
	}
	t.bounds = bounds
	return t
}

func (t *FlatTracer) TraceDeep(x, y float64, ray Vector) (bool, TraceIntervals) {
	return DeepifyTrace(t, x, y, ray)
}

func (t *FlatTracer) Pruned(rp RenderingParameters) Tracer {
	return SimplyPruned(t, rp) //FIXME
}

func (t *FlatTracer) Trace(x, y float64, ray Vector) (bool, float64, Vector, Color) {
	ok, inter := IntersectionOfPlaneAndLine(t.p1.CenterCS, t.w1, t.w2, Vector{0, 0, 0}, ray)
	if !ok || inter[2] < 0 {
		return false, 0, NoDirection, Color{}
	}
	if inter[0] < 0 || inter[0] > 1 || inter[1] < 0 || inter[1] > 1 {
		return false, 0, NoDirection, Color{}
	}
	if !t.fourCorners && inter[0]+inter[1] > 1 {
		return false, 0, NoDirection, Color{}
	}
	z := inter[2]

	var col Color
	if t.fourCorners {
		col = MixColors(MixColors(t.p1.BaseBall.Color, t.p2.BaseBall.Color, inter[0]), MixColors(t.p3.BaseBall.Color, t.fourthColor, inter[0]), inter[1])
	} else {
		f1 := 1.0
		if inter[1] < 1 {
			f1 = inter[0] / (1 - inter[1])
		}
		col = MixColors(MixColors(t.p1.BaseBall.Color, t.p2.BaseBall.Color, f1), t.p3.BaseBall.Color, inter[1])
	}
	return true, z, t.dir, col
}

func (t *FlatTracer) GetBounds() Bounds {
	return t.bounds
}
