package elements

import (
	. "bitbucket.org/balpha/go-unicornify/unicornify/core"
	. "bitbucket.org/balpha/go-unicornify/unicornify/rendering"
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

func (b *Bone) GetTracer(wv WorldView) Tracer {
	b1 := b.Balls[0]
	b2 := b.Balls[1]
	proj1 := ProjectBall(wv, b1)
	proj2 := ProjectBall(wv, b2)

	if b.XFunc == nil && b.YFunc == nil {
		return NewBoneTracer(proj1, proj2)
	} else {
		c1 := b1.Color
		c2 := b2.Color

		v := b2.Center.Minus(b1.Center)
		length := v.Length()
		vx, vy := CrossAxes(v.Times(1 / length))

		parts := 255

		calcBall := func(factor float64) BallProjection {
			col := MixColors(c1, c2, factor)
			fx, fy := factor, factor
			if f := b.XFunc; f != nil {
				fx = f(fx)
			}
			if f := b.YFunc; f != nil {
				fy = f(fy)
			}

			c := b1.Center.Plus(v.Times(factor)).Plus(vx.Times((fx - factor) * length)).Plus(vy.Times((fy - factor) * length))
			r := MixFloats(b1.Radius, b2.Radius, factor)
			ballp := ProjectBall(wv, NewBallP(c, r, col))
			return ballp
		}

		prevBall := proj1

		result := NewGroupTracer()

		nextBall := calcBall(1 / float64(parts))

		for i := 1; i <= parts; i++ {
			curBall := nextBall

			if i < parts {
				nextBall = calcBall(float64(i+1) / float64(parts))
				seg1 := curBall.BaseBall.Center.Minus(prevBall.BaseBall.Center)
				seg2 := nextBall.BaseBall.Center.Minus(curBall.BaseBall.Center)
				if seg1.ScalarProd(seg2)/(seg1.Length()*seg2.Length()) > 0.999848 { // cosine of 1°
					continue
				}
			}

			tracer := NewBoneTracer(prevBall, curBall)

			result.Add(tracer)

			prevBall = curBall
		}
		return result
	}
}

type BoneTracer struct {
	w1, w2, w3, a1, a2, a3, ra, dr    float64
	c2, c4, c6, c8, c9, c11, c14, c2i float64
	b1, b2                            BallProjection
	bounds                            Bounds
}

func (t *BoneTracer) GetBounds() Bounds {
	return t.bounds
}

func (t *BoneTracer) Trace(x, y float64, ray Vector) (bool, float64, Vector, Color) {
	return t.traceImpl(x, y, ray, false)
}
func (t *BoneTracer) traceImpl(x, y float64, ray Vector, backside bool) (bool, float64, Vector, Color) {
	v1, v2, v3 := ray.Decompose()

	c3 := -2 * (v1*t.w1 + v2*t.w2 + v3*t.w3)
	c5 := -2 * (v1*t.a1 + v2*t.a2 + v3*t.a3)

	var z, f float64

	if t.c2 == 0 {
		if t.dr > 0 {
			f = 1
		} else {
			f = 0
		}
	} else {
		c7 := c3 * t.c2i
		c10 := c5 * t.c2i
		c12i := 1 / (Sqr(c7)/4 - t.c9)
		c13 := c7*t.c8/2 - c10

		pz := c13 * c12i
		qz := t.c14 * c12i
		discz := Sqr(pz)/4 - qz

		if discz < 0 {
			return false, 0, NoDirection, Color{}
		}

		rdiscz := math.Sqrt(discz)
		z1 := -pz/2 + rdiscz
		z2 := -pz/2 - rdiscz
		f1 := -(c3*z1 + t.c4) / (2 * t.c2)
		f2 := -(c3*z2 + t.c4) / (2 * t.c2)

		g1 := t.ra+f1*t.dr >= 0
		g2 := t.ra+f2*t.dr >= 0

		if !g1 {
			z1, f1 = z2, f2
			if t.dr > 0 {
				f2 = 1
			} else {
				f2 = 0
			}
		}
		if !g2 {
			z2, f2 = z1, f1
			if t.dr > 0 {
				f1 = 1
			} else {
				f1 = 0
			}
		}

		if backside {
			z = z1
			f = f1
		} else {
			z = z2
			f = f2
		}
	}

	if f <= 0 || f >= 1 {
		f = math.Min(1, math.Max(0, f))
		pz := c3*f + c5
		qz := t.c2*Sqr(f) + t.c4*f + t.c6
		discz := Sqr(pz)/4 - qz
		if discz < 0 {
			f = 1 - f
			pz = c3*f + c5
			qz = t.c2*Sqr(f) + t.c4*f + t.c6
			discz = Sqr(pz)/4 - qz

			if discz < 0 {
				return false, 0, NoDirection, Color{}
			}
		}

		if backside {
			z = -pz/2 + math.Sqrt(discz)
		} else {
			z = -pz/2 - math.Sqrt(discz)
		}

	}

	m1 := t.a1 + f*t.w1
	m2 := t.a2 + f*t.w2
	m3 := t.a3 + f*t.w3

	p := Vector{v1, v2, v3}.Times(z)
	dir := p.Minus(Vector{m1, m2, m3})
	return true, z, dir, MixColors(t.b1.BaseBall.Color, t.b2.BaseBall.Color, f)

}

func (t *BoneTracer) TraceDeep(x, y float64, ray Vector) (bool, TraceIntervals) {
	ok1, z1, dir1, col1 := t.traceImpl(x, y, ray, false)
	ok2, z2, dir2, col2 := t.traceImpl(x, y, ray, true)
	if ok1 {
		if !ok2 { // this can happen because of rounding errors
			return false, TraceIntervals{}
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

func (t *BoneTracer) Pruned(rp RenderingParameters) Tracer {
	return SimplyPruned(t, rp)
}

func NewBoneTracer(b1, b2 BallProjection) *BoneTracer {

	t := &BoneTracer{b1: b1, b2: b2}

	cx1, cy1, cz1, r1 := b1.CenterCS.X(), b1.CenterCS.Y(), b1.CenterCS.Z(), b1.BaseBall.Radius
	cx2, cy2, cz2, r2 := b2.CenterCS.X(), b2.CenterCS.Y(), b2.CenterCS.Z(), b2.BaseBall.Radius

	t.bounds = RenderingBoundsForBalls(b1, b2)

	t.w1 = float64(cx2 - cx1)
	t.w2 = float64(cy2 - cy1)
	t.w3 = float64(cz2 - cz1)

	t.a1 = float64(cx1)
	t.a2 = float64(cy1)
	t.a3 = float64(cz1)

	t.ra = float64(r1)
	t.dr = float64(r2 - r1)

	t.c2 = -Sqr(t.dr) + Sqr(t.w1) + Sqr(t.w2) + Sqr(t.w3)
	if t.c2 != 0 {
		t.c2i = 1 / t.c2
	}

	t.c4 = -2*t.ra*t.dr + 2*(t.a1*t.w1+t.a2*t.w2+t.a3*t.w3)
	t.c6 = -Sqr(t.ra) + Sqr(t.a1) + Sqr(t.a2) + Sqr(t.a3)
	t.c8 = t.c4 / t.c2
	t.c9 = 1 / t.c2
	t.c11 = t.c6 / t.c2
	t.c14 = Sqr(t.c8)/4 - t.c11

	return t
}
