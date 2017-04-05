package elements

import (
	. "bitbucket.org/balpha/go-unicornify/unicornify/core"
	. "bitbucket.org/balpha/go-unicornify/unicornify/rendering"
)

type Steak struct {
	Balls       [3]*Ball
	FourCorners bool
	FourthColor Color // ignored if FourCorners is false
	Rounded     bool
}

func NewSteak(b1, b2, b3 *Ball) *Steak {
	return &Steak{Balls: [3]*Ball{b1, b2, b3}}
}

func (s *Steak) GetTracer(wv WorldView) Tracer {
	return NewSteakTracer(
		wv,
		s.Balls[0],
		s.Balls[1],
		s.Balls[2],
		s.FourCorners,
		s.FourthColor,
		s.Rounded,
	)
}

type SteakTracer struct {
	bounds Bounds
	Flats  []*FlatTracer
	wv     WorldView
}

func NewSteakTracer(wv WorldView, b1, b2, b3 *Ball, fourCorners bool, fourthColor Color, rounded bool) Tracer {
	t := &SteakTracer{
		wv:     wv,
		bounds: EmptyBounds,
	}

	col1 := b1.Color
	col2 := b2.Color
	col3 := b3.Color

	c1 := b1.Center
	w12 := b2.Center.Minus(b1.Center)
	w13 := b3.Center.Minus(b1.Center)

	cross := w12.CrossProd(w13).Unit().Times(b1.Radius)
	top1 := c1.Plus(cross) // not necessarily on the top in any non-arbitrary way
	bottom1 := c1.Plus(cross.Neg())
	add := func(tri bool, b1, b2, b3 *Ball, fourthColor Color, roughDirection Vector) {
		ft := NewFlatTracer(wv, b1, b2, b3, !tri, fourthColor, roughDirection)
		t.bounds = t.bounds.Union(ft.GetBounds())
		t.Flats = append(t.Flats, ft)
	}

	add(!fourCorners, NewBallP(top1, 0, col1), NewBallP(top1.Plus(w12), 0, col2), NewBallP(top1.Plus(w13), 0, col3), fourthColor, cross)
	add(!fourCorners, NewBallP(bottom1, 0, col1), NewBallP(bottom1.Plus(w12), 0, col2), NewBallP(bottom1.Plus(w13), 0, col3), fourthColor, cross.Neg())
	add(false, NewBallP(top1, 0, col1), NewBallP(bottom1, 0, col1), NewBallP(top1.Plus(w12), 0, col2), col2, w13.Neg())
	add(false, NewBallP(top1, 0, col1), NewBallP(bottom1, 0, col1), NewBallP(top1.Plus(w13), 0, col3), col3, w12.Neg())

	var w14 Vector

	if !fourCorners {
		add(false, NewBallP(top1.Plus(w12), 0, col2), NewBallP(bottom1.Plus(w12), 0, col2), NewBallP(top1.Plus(w13), 0, col3), col3, w12)
	} else {
		w14 = w12.Plus(w13)
		add(false, NewBallP(top1.Plus(w12), 0, col2), NewBallP(bottom1.Plus(w12), 0, col2), NewBallP(top1.Plus(w14), 0, fourthColor), fourthColor, w12)
		add(false, NewBallP(top1.Plus(w13), 0, col3), NewBallP(bottom1.Plus(w13), 0, col3), NewBallP(top1.Plus(w14), 0, fourthColor), fourthColor, w13)
	}

	if rounded {
		sides := &Figure{}
		sides.Add(
			NewBone(b1, b2),
			NewBone(b1, b3),
		)
		if fourCorners {
			b4 := NewBallP(b1.Center.Plus(w14), b1.Radius, fourthColor)
			sides.Add(
				NewBone(b2, b4),
				NewBone(b3, b4),
			)
		} else {
			sides.Add(NewBone(b2, b3))
		}
		gt := NewGroupTracer()
		st := sides.GetTracer(wv)
		gt.Add(st, t)
		return gt
	}
	return t
}

func (t *SteakTracer) TraceDeep(x, y float64) (bool, TraceIntervals) {
	ray := Vector{x, y, t.wv.FocalLength}.Unit()

	var minz, maxz float64
	var mindir, maxdir Vector
	var mincol, maxcol Color
	any := false

	for _, ft := range t.Flats {
		if !ft.GetBounds().ContainsXY(x, y) {
			continue
		}
		ok, z, dir, col := ft.TraceRay(ray)
		if !ok {
			continue
		}
		if !any || z > maxz {
			maxz, maxdir, maxcol = z, dir, col
		}
		if !any || z < minz {
			minz, mindir, mincol = z, dir, col
		}
		any = true
	}
	if !any {
		return false, TraceIntervals{}
	}

	return true, TraceIntervals{
		TraceInterval{
			Start: TraceResult{minz, mindir, mincol},
			End:   TraceResult{maxz, maxdir, maxcol},
		},
	}
}

func (t *SteakTracer) Trace(x, y float64) (bool, float64, Vector, Color) {
	return UnDeepifyTrace(t, x, y)
}

func (t *SteakTracer) GetBounds() Bounds {
	return t.bounds
}
