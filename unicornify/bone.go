package unicornify

import (
	"image"
	"math"
)

type Bone struct {
	Balls        [2]*Ball
	XFunc, YFunc func(float64) float64 // may be nil
}

func NewBone(b1, b2 *Ball) *Bone {
	return NewNonLinBone(b1, b2, nil, nil)
}

func NewShadedBone(b1, b2 *Ball, shading float64) *Bone {
	return NewShadedNonLinBone(b1, b2, nil, nil, shading)
}

const defaultShading = 0.25

func NewNonLinBone(b1, b2 *Ball, xFunc, yFunc func(float64) float64) *Bone {
	return NewShadedNonLinBone(b1, b2, xFunc, yFunc, defaultShading)
}

func NewShadedNonLinBone(b1, b2 *Ball, xFunc, yFunc func(float64) float64, shading float64) *Bone {
	return &Bone{[2]*Ball{b1, b2}, xFunc, yFunc}
}

func reverse(f func(float64) float64) func(float64) float64 {
	return func(v float64) float64 {
		return 1 - f(1-v)
	}
}

func (b *Bone) Project(wv WorldView) {
	b.Balls[0].Project(wv)
	b.Balls[1].Project(wv)
}

func (b *Bone) GetTracer(wv WorldView) Tracer {
	b1 := b.Balls[0]
	b2 := b.Balls[1]

	if b.XFunc == nil && b.YFunc == nil {
		return NewBoneTracer(b1, b2)
	} else {
		c1 := b1.Color
		c2 := b2.Color

		v := b2.Center.Minus(b1.Center)
		length := v.Length()
		vx, vy := CrossAxes(v.Times(1 / length))

		bounding := b.Bounding()
		parts := bounding.Dy()
		if bounding.Dx() > bounding.Dy() {
			parts = bounding.Dx()
		}
		parts = roundUp(float64(parts) * SCALE)

		calcBall := func(factor float64) *Ball {
			col := MixColors(c1, c2, factor)
			fx, fy := factor, factor
			if f := b.XFunc; f != nil {
				fx = f(fx)
			}
			if f := b.YFunc; f != nil {
				fy = f(fy)
			}

			c := b1.Center.Shifted(v.Times(factor)).Shifted(vx.Times((fx - factor) * length)).Shifted(vy.Times((fy - factor) * length))
			r := MixFloats(b1.Radius, b2.Radius, factor)
			ball := NewBallP(c, r, col)
			ball.Project(b1.Projection.WorldView)
			return ball
		}

		prevBall := b1

		result := NewGroupTracer()
		subgroup := NewGroupTracer()

		nextBall := calcBall(1 / float64(parts))

		for i := 1; i <= parts; i++ {
			curBall := nextBall

			if i < parts {
				nextBall = calcBall(float64(i+1) / float64(parts))
				seg1 := curBall.Center.Minus(prevBall.Center)
				seg2 := nextBall.Center.Minus(curBall.Center)
				if seg1.ScalarProd(seg2)/(seg1.Length()*seg2.Length()) > 0.999848 { // cosine of 1°
					continue
				}
			}

			tracer := NewBoneTracer(prevBall, curBall)
			subgroup.Add(tracer)

			if i%5 == 0 || i == parts {
				result.Add(subgroup)
				subgroup = NewGroupTracer()
			}

			prevBall = curBall
		}

		return result

	}
}

func (b Bone) Bounding() image.Rectangle {
	return b.Balls[0].Bounding().Union(b.Balls[1].Bounding())
}

type BoneTracer struct {
	xmin, xmax, ymin, ymax         int
	v1, v2, v3, a1, a2, a3, ra, dr float64
	b1, b2                         *Ball
}

func (t *BoneTracer) GetBounds() image.Rectangle {
	return image.Rect(t.xmin, t.ymin, t.xmax, t.ymax)
}

func (t *BoneTracer) Trace(x, y float64) (bool, float64, Point3d, Color) {
	return t.traceImpl(x, y, false)
}
func (t *BoneTracer) traceImpl(x, y float64, backside bool) (bool, float64, Point3d, Color) {
	p1 := x
	p2 := y

	s1 := p1 - t.a1
	s2 := p2 - t.a2

	c1 := sqr(t.v1) + sqr(t.v2) + sqr(t.v3) - sqr(t.dr)
	c2 := -2*s1*t.v1 - 2*s2*t.v2 - 2*t.ra*t.dr
	c3 := sqr(s1) + sqr(s2) - sqr(t.ra)

	ks := 4*sqr(t.v3) - 4*c1
	ms := -4 * t.v3 * c2
	ns := sqr(c2) - 4*c1*c3

	var f, s3 float64

	if c1 <= 0 {
		if t.dr > 0 {
			f = 1
		} else {
			f = 0
		}
	} else if ks == 0 {
		f = 1
	} else {
		ps := ms / ks
		qs := ns / ks

		disc := sqr(ps)/4 - qs
		if disc < 0 {
			return false, 0, NoDirection, Black
		}

		if backside {
			s3 = -ps/2 + math.Sqrt(disc)
		} else {
			s3 = -ps/2 - math.Sqrt(disc)
		}

		kf := c1
		mf := -2*s3*t.v3 + c2
		nf := sqr(s3) + c3 //unnecessary

		pf := mf / kf
		qf := nf / kf
		_ = qf

		f = math.Min(1, math.Max(0, -pf/2))

		if t.ra+(-pf/2)*t.dr < 0 {
			f = 1 - f
		}
	}

	m1 := t.a1 + f*t.v1
	m2 := t.a2 + f*t.v2
	m3 := t.a3 + f*t.v3
	r := t.ra + f*t.dr

	var p3 float64

	if f <= 0 || f >= 1 {
		pm := -2 * m3
		qm := sqr(p1-m1) + sqr(p2-m2) + sqr(m3) - sqr(r)

		discm := sqr(pm)/4 - qm
		if discm < 0 {
			return false, 0, NoDirection, Black
		}
		if backside {
			p3 = -pm/2 + math.Sqrt(discm)
		} else {
			p3 = -pm/2 - math.Sqrt(discm)
		}

	} else {
		p3 = s3 + t.a3
	}

	d1 := p1 - m1
	d2 := p2 - m2
	d3 := p3 - m3

	return true, p3, Point3d{d1, d2, d3}, MixColors(t.b1.Color, t.b2.Color, math.Max(0, math.Min(1, f)))
}

func (t *BoneTracer) TraceDeep(x, y float64) (bool, TraceIntervals) {
	ok1, z1, dir1, col1 := t.traceImpl(x, y, false)
	ok2, z2, dir2, col2 := t.traceImpl(x, y, true)
	if ok1 {
		if !ok2 {
			panic("Huh? That makes no sense")
		}
		return true, TraceIntervals{
			TraceInterval{
				Start: TraceResult{z1, dir1, col1},
				End:   TraceResult{z2, dir2, col2},
			},
		}
	}
	return false, TraceIntervals{}
}

func NewBoneTracer(b1, b2 *Ball) *BoneTracer {

	t := &BoneTracer{b1: b1, b2: b2}

	cx1, cy1, cz1, r1 := b1.Projection.X(), b1.Projection.Y(), b1.Projection.Z(), b1.Projection.Radius
	cx2, cy2, cz2, r2 := b2.Projection.X(), b2.Projection.Y(), b2.Projection.Z(), b2.Projection.Radius

	t.xmin = roundDown(math.Min(cx1-r1, cx2-r2))
	t.xmax = roundUp(math.Max(cx1+r1, cx2+r2))
	t.ymin = roundDown(math.Min(cy1-r1, cy2-r2))
	t.ymax = roundUp(math.Max(cy1+r1, cy2+r2))

	t.v1 = float64(cx2 - cx1)
	t.v2 = float64(cy2 - cy1)
	t.v3 = float64(cz2 - cz1)

	t.a1 = float64(cx1)
	t.a2 = float64(cy1)
	t.a3 = float64(cz1)

	t.ra = float64(r1)
	t.dr = float64(r2 - r1)

	return t
}
