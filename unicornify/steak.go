package unicornify

import (
	"image"
)
type Steak struct {
	Balls [3]*Ball
	FourCorners bool
	FourthColor Color // ignored if FourCorners is false
	Rounded bool
}

func (s *Steak) Project(wv WorldView) {
	for _, b := range s.Balls {
		b.Project(wv)
	}
}
func (s *Steak) Bounding() image.Rectangle {
	return s.GetTracer(s.Balls[0].Projection.WorldView).GetBounds()
}
func NewSteak(b1, b2, b3 *Ball) *Steak {
	return &Steak{Balls:[3]*Ball{b1, b2, b3}}
}

func (s *Steak) GetTracer(wv WorldView) Tracer {
	return NewSteakTracer(s.Balls[0], s.Balls[1], s.Balls[2], s.FourCorners, s.FourthColor, s.Rounded)
}

type SteakTracer struct {
	// b4 can be nil, in which case it's a triangle. the w*4 vectors are ignored then
	b1, b2, b3, b4                         *Ball
	w12, w13, w23, w24, w34 Point3d
	cross Point3d
}

func NewSteakTracer(b1, b2, b3 *Ball, fourCorners bool, fourthColor Color, rounded bool) Tracer {
	if b1.Radius != b2.Radius || b1.Radius != b3.Radius {
		panic("Steak with differing radii not implemented yet")
	}
	
	w12 := b2.Projection.CenterCS.Minus(b1.Projection.CenterCS)
	w13 := b3.Projection.CenterCS.Minus(b1.Projection.CenterCS)
	
	t := &SteakTracer{
		b1:b1, b2:b2, b3:b3,
		w12: w12,
		w13: w13,
		w23: b3.Projection.CenterCS.Minus(b2.Projection.CenterCS),
	}
	
	if fourCorners {
		t.b4 = NewBallP(t.b2.Center.Shifted(t.b3.Center.Minus(t.b1.Center)), t.b1.Radius, fourthColor)
		t.b4.Project(t.b1.Projection.WorldView)
		t.w24 = t.b4.Projection.CenterCS.Minus(b2.Projection.CenterCS)
		t.w34 = t.b4.Projection.CenterCS.Minus(b3.Projection.CenterCS)
	}
	
	t.cross = w12.CrossProd(w13).Unit().Times(b1.Radius)
	
	if rounded {
		sides := &Figure{}
		sides.Add(
			NewBone(b1, b2),
			NewBone(b1, b3),
		)
		if fourCorners {
			sides.Add(
				NewBone(b2, t.b4),
				NewBone(b3, t.b4),
			)
		} else {
			sides.Add(NewBone(b2, t.b3))
		}
		gt := NewGroupTracer()
		st := sides.GetTracer(b1.Projection.WorldView)
		gt.Add(st, t)
		return gt
	}
	
	return t
}

func (t *SteakTracer) TraceDeep(x, y float64) (bool, TraceIntervals) {
	return DeepifyTrace(t, x, y)
}

func (t *SteakTracer) Trace(x, y float64) (bool, float64, Point3d, Color) {	
	focalLength := t.b1.Projection.ProjectedCenterCS.Z()
	v := Point3d{x, y, focalLength}.Unit()
	cam := Point3d{0,0,0}
	c1 := t.b1.Projection.CenterCS
	var minz, maxz float64
	var minc, maxc Color
	var mina1, mina2, maxa1, maxa2, minopp, maxopp Point3d
	any := false
	
	add := func (tri bool, p0, ep1, ep2 Point3d, col1, col2, col3, col4 Color, perpOpposit Point3d) {
		ok, inter := IntersectionOfPlaneAndLine(p0, ep1, ep2, cam, v)
		if !ok {
			return
		}
		if inter[0] < 0 || inter[0] > 1 || inter[1] < 0 || inter[1] > 1 {
			return
		}
		if tri && inter[0] + inter[1] > 1 {
			return
		}
		z := inter[2]
		var col Color
		if tri {
			f1 := 1.0
			if inter[1] < 1 {
				f1 = inter[0]/(1-inter[1])
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
	if t.b4 != nil {
		col4 = t.b4.Color
	}
	add(t.b4==nil, c1.Shifted(t.cross), t.w12, t.w13, t.b1.Color, t.b2.Color, t.b3.Color, col4, t.cross.Neg())
	add(t.b4==nil, c1.Shifted(t.cross.Neg()), t.w12, t.w13, t.b1.Color, t.b2.Color, t.b3.Color, col4, t.cross)
	add(false, c1.Shifted(t.cross), t.cross.Times(-2), t.w12, t.b1.Color, t.b1.Color, t.b2.Color, t.b2.Color, t.w13)
	add(false, c1.Shifted(t.cross), t.cross.Times(-2), t.w13, t.b1.Color, t.b1.Color, t.b3.Color, t.b3.Color, t.w12)
	if t.b4 == nil {
		add(false, t.b2.Projection.CenterCS.Shifted(t.cross), t.cross.Times(-2), t.w23, t.b2.Color, t.b2.Color, t.b3.Color, t.b3.Color, t.w12)
	} else {
		add(false, t.b2.Projection.CenterCS.Shifted(t.cross), t.cross.Times(-2), t.w24, t.b2.Color, t.b2.Color, t.b4.Color, t.b4.Color, t.w12.Neg())
		add(false, t.b3.Projection.CenterCS.Shifted(t.cross), t.cross.Times(-2), t.w34, t.b3.Color, t.b3.Color, t.b4.Color, t.b4.Color, t.w13.Neg())
	}
	
	if (!any) {
		return false, 0, NoDirection, Color{}
	}
	
	dir := mina1.CrossProd(mina2)
	if dir.ScalarProd(minopp) > 0 {
		dir = dir.Times(-1)
	}
	
	return true, minz, dir, minc
}

func (t *SteakTracer) GetBounds() image.Rectangle {
	r := t.b1.Bounding().Union(t.b2.Bounding()).Union(t.b3.Bounding())
	if t.b4 != nil {
		r = r.Union(t.b4.Bounding())
	}
	return r
}