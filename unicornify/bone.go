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
	w1, w2, w3, a1, a2, a3, ra, dr float64
	b1, b2                         *Ball
}

func (t *BoneTracer) GetBounds() image.Rectangle {
	return image.Rect(t.xmin, t.ymin, t.xmax, t.ymax)
}

func (t *BoneTracer) Trace(x, y float64) (bool, float64, Point3d, Color) {
	return t.traceImpl(x, y, false)
}
func (t *BoneTracer) traceImpl(x, y float64, backside bool) (bool, float64, Point3d, Color) {
	focalLength := t.b1.Projection.ProjectedCenterCS.Z()
	v1, v2, v3 := Point3d{x, y, focalLength}.Unit().Decompose()
	

	c1 := sqr(v1)+sqr(v2)+sqr(v3) // note: always 1
	c2 := -sqr(t.dr)+sqr(t.w1)+sqr(t.w2)+sqr(t.w3)
	c3 := -2*(v1*t.w1+v2*t.w2+v3*t.w3)
	c4 := -2*t.ra*t.dr+2*(t.a1*t.w1+t.a2*t.w2+t.a3*t.w3)
	c5 := -2*(v1*t.a1+v2*t.a2+v3*t.a3)
	c6 := -sqr(t.ra)+sqr(t.a1)+sqr(t.a2)+sqr(t.a3)
	
	
	var z,f float64
	
	if c2 == 0 {
		if t.dr > 0 {
			f = 1
		} else {
			f = 0
		}
	} else {
		c7 := c3/c2
		c8 := c4/c2
		c9 := c1/c2
		c10 := c5/c2
		c11 := c6/c2
		c12 := sqr(c7)/4-c9
		c13 := c7*c8/2-c10
		c14:= sqr(c8)/4-c11
		
		pz := c13/c12
		qz := c14/c12
		discz := sqr(pz)/4-qz
		
		
		if discz < 0 {
			return false, 0, NoDirection, Color{}
		}
		
		rdiscz := math.Sqrt(discz)
		z1 := -pz/2 + rdiscz
		z2 := -pz/2 - rdiscz
		f1 := -(c3*z1+c4)/(2*c2)
		f2 := -(c3*z2+c4)/(2*c2)
		
		g1 := t.ra+f1*t.dr>=0
		g2 := t.ra+f2*t.dr>=0
		
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
	
	
	if f<=0 || f>=1 {
		f = math.Min(1, math.Max(0, f))
		pz := (c3*f+c5)/c1
		qz := (c2*sqr(f)+c4*f+c6)/c1
		discz := sqr(pz)/4-qz
		if discz < 0 {
			f = 1 - f
			pz = (c3*f+c5)/c1
			qz = (c2*sqr(f)+c4*f+c6)/c1
			discz = sqr(pz)/4-qz

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
	
	p := Point3d{v1, v2, v3}.Times(z)
	dir := p.Minus(Point3d{m1,m2,m3})
	return true, z, dir, MixColors(t.b1.Color, t.b2.Color, f)
	
}

func (t *BoneTracer) TraceDeep(x, y float64) (bool, TraceIntervals) {
	ok1, z1, dir1, col1 := t.traceImpl(x, y, false)
	ok2, z2, dir2, col2 := t.traceImpl(x, y, true)
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

func NewBoneTracer(b1, b2 *Ball) *BoneTracer {

	t := &BoneTracer{b1: b1, b2: b2}

	cx1, cy1, cz1, r1 := b1.Projection.CenterCS.X(), b1.Projection.CenterCS.Y(), b1.Projection.CenterCS.Z(), b1.Radius
	cx2, cy2, cz2, r2 := b2.Projection.CenterCS.X(), b2.Projection.CenterCS.Y(), b2.Projection.CenterCS.Z(), b2.Radius

	pcx1, pcy1, pr1 := b1.Projection.X(), b1.Projection.Y(), b1.Projection.Radius
	pcx2, pcy2, pr2 := b2.Projection.X(), b2.Projection.Y(), b2.Projection.Radius
	
	t.xmin = roundDown(math.Min(pcx1-pr1, pcx2-pr2))
	t.xmax = roundUp(math.Max(pcx1+pr1, pcx2+pr2))
	t.ymin = roundDown(math.Min(pcy1-pr1, pcy2-pr2))
	t.ymax = roundUp(math.Max(pcy1+pr1, pcy2+pr2))

	t.w1 = float64(cx2 - cx1)
	t.w2 = float64(cy2 - cy1)
	t.w3 = float64(cz2 - cz1)

	t.a1 = float64(cx1)
	t.a2 = float64(cy1)
	t.a3 = float64(cz1)

	t.ra = float64(r1)
	t.dr = float64(r2 - r1)

	return t
}
