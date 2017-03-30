package unicornify

import (
	"image"
)

type Steak struct {
	Balls       [3]*Ball
	FourCorners bool
	FourthColor Color // ignored if FourCorners is false
	Rounded     bool
}

func (s *Steak) Project(wv WorldView) {
	for _, b := range s.Balls {
		b.Project(wv)
	}
}

func NewSteak(b1, b2, b3 *Ball) *Steak {
	return &Steak{Balls: [3]*Ball{b1, b2, b3}}
}

func (s *Steak) GetTracer(wv WorldView) Tracer {
	return NewSteakTracer(
		wv.ProjectBall(s.Balls[0]),
		wv.ProjectBall(s.Balls[1]),
		wv.ProjectBall(s.Balls[2]),
		s.FourCorners,
		s.FourthColor,
		s.Rounded)
}

type SteakTracer struct {
	// b4 can be nil, in which case it's a triangle. the w*4 vectors are ignored then
	b1, b2, b3, b4          BallProjection
	w12, w13, w23, w24, w34 Point3d
	cross                   Point3d
	bounds                  image.Rectangle
	fourCorners             bool
}

func NewSteakTracer(b1, b2, b3 BallProjection, fourCorners bool, fourthColor Color, rounded bool) Tracer {
	if b1.BaseBall.Radius != b2.BaseBall.Radius || b1.BaseBall.Radius != b3.BaseBall.Radius {
		panic("Steak with differing radii not implemented yet")
	}
	if b1.WorldView != b2.WorldView || b1.WorldView != b3.WorldView {
		panic("Steak balls have been projected with differing world views")
	}

	w12 := b2.CenterCS.Minus(b1.CenterCS)
	w13 := b3.CenterCS.Minus(b1.CenterCS)

	t := &SteakTracer{
		b1: b1, b2: b2, b3: b3,
		w12: w12,
		w13: w13,
		w23: b3.CenterCS.Minus(b2.CenterCS),
	}

	if fourCorners {
		t.b4 = b1.WorldView.ProjectBall(NewBallP(t.b2.BaseBall.Center.Shifted(t.b3.BaseBall.Center.Minus(t.b1.BaseBall.Center)), t.b1.BaseBall.Radius, fourthColor))
		t.w24 = t.b4.CenterCS.Minus(b2.CenterCS)
		t.w34 = t.b4.CenterCS.Minus(b3.CenterCS)
		t.fourCorners = true
	}

	t.cross = w12.CrossProd(w13).Unit().Times(b1.BaseBall.Radius)

	wv := t.b1.WorldView
	r := t.b1.BaseBall.GetTracer(wv).GetBounds().Union(t.b2.BaseBall.GetTracer(wv).GetBounds()).Union(t.b3.BaseBall.GetTracer(wv).GetBounds())
	if fourCorners {
		r = r.Union(t.b4.BaseBall.GetTracer(wv).GetBounds())
	}
	t.bounds = r

	if rounded {
		sides := &Figure{}
		sides.Add(
			NewBone(&b1.BaseBall, &b2.BaseBall),
			NewBone(&b1.BaseBall, &b3.BaseBall),
		)
		if fourCorners {
			sides.Add(
				NewBone(&b2.BaseBall, &t.b4.BaseBall),
				NewBone(&b3.BaseBall, &t.b4.BaseBall),
			)
		} else {
			sides.Add(NewBone(&b2.BaseBall, &t.b3.BaseBall))
		}
		gt := NewGroupTracer()
		st := sides.GetTracer(b1.WorldView)
		gt.Add(st, t)
		return gt
	}

	return t
}

func (t *SteakTracer) Trace(x, y float64) (bool, float64, Point3d, Color) {
	return UnDeepifyTrace(t, x, y)
}

func (t *SteakTracer) TraceDeep(x, y float64) (bool, TraceIntervals) {
	focalLength := t.b1.ProjectedCenterCS.Z()
	v := Point3d{x, y, focalLength}.Unit()
	cam := Point3d{0, 0, 0}
	c1 := t.b1.CenterCS
	var minz, maxz float64
	var minc, maxc Color
	var mina1, mina2, maxa1, maxa2, minopp, maxopp Point3d
	any := false

	add := func(tri bool, p0, ep1, ep2 Point3d, col1, col2, col3, col4 Color, perpOpposit Point3d) {
		ok, inter := IntersectionOfPlaneAndLine(p0, ep1, ep2, cam, v)
		if !ok {
			return
		}
		if inter[0] < 0 || inter[0] > 1 || inter[1] < 0 || inter[1] > 1 {
			return
		}
		if tri && inter[0]+inter[1] > 1 {
			return
		}
		z := inter[2]
		var col Color
		if tri {
			f1 := 1.0
			if inter[1] < 1 {
				f1 = inter[0] / (1 - inter[1])
			}
			col = MixColors(MixColors(col1, col2, f1), col3, inter[1])
		} else {
			col = MixColors(MixColors(col1, col2, inter[0]), MixColors(col3, col4, inter[0]), inter[1])
		}
		if !any || z > maxz {
			maxz = z
			maxc = col
			maxa1 = ep1
			maxa2 = ep2
			maxopp = perpOpposit
		}
		if !any || z < minz {
			minz = z
			minc = col
			mina1 = ep1
			mina2 = ep2
			minopp = perpOpposit
		}
		any = true
	}
	col4 := Color{}
	if t.fourCorners {
		col4 = t.b4.BaseBall.Color
	}
	col1 := t.b1.BaseBall.Color
	col2 := t.b2.BaseBall.Color
	col3 := t.b3.BaseBall.Color

	add(!t.fourCorners, c1.Shifted(t.cross), t.w12, t.w13, col1, col2, col3, col4, t.cross.Neg())
	add(!t.fourCorners, c1.Shifted(t.cross.Neg()), t.w12, t.w13, col1, col2, col3, col4, t.cross)
	add(false, c1.Shifted(t.cross), t.cross.Times(-2), t.w12, col1, col1, col2, col2, t.w13)
	add(false, c1.Shifted(t.cross), t.cross.Times(-2), t.w13, col1, col1, col3, col2, t.w12)
	if !t.fourCorners {
		add(false, t.b2.CenterCS.Shifted(t.cross), t.cross.Times(-2), t.w23, col2, col2, col3, col3, t.w12)
	} else {
		add(false, t.b2.CenterCS.Shifted(t.cross), t.cross.Times(-2), t.w24, col2, col2, col4, col4, t.w12.Neg())
		add(false, t.b3.CenterCS.Shifted(t.cross), t.cross.Times(-2), t.w34, col3, col3, col4, col4, t.w13.Neg())
	}

	if !any {
		return false, TraceIntervals{}
	}

	mindir := mina1.CrossProd(mina2)
	if mindir.ScalarProd(minopp) > 0 {
		mindir = mindir.Times(-1)
	}

	maxdir := maxa1.CrossProd(maxa2)
	if maxdir.ScalarProd(maxopp) > 0 {
		maxdir = maxdir.Times(-1)
	}
	return true, TraceIntervals{
		TraceInterval{
			Start: TraceResult{minz, mindir, minc},
			End:   TraceResult{maxz, maxdir, maxc},
		},
	}
}

func (t *SteakTracer) GetBounds() image.Rectangle {
	return t.bounds
}
