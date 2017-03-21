package unicornify

import (
	"image"
	"image/color"
	"math"
)

const (
	CirclyGradient   = iota
	DistanceGradient = iota
)

type ColoringParameters struct {
	Shading  float64
	Gradient int
}

func DefaultGradientWithShading(shading float64) ColoringParameters {
	return ColoringParameters{shading, CirclyGradient}
}

func CircleShadingRGBA(x, y, r float64, col color.RGBA, coloring ColoringParameters) color.RGBA {
	if coloring.Shading == 0 || y == 0 {
		return col
	}
	var sh float64
	lighten := 128.0
	switch coloring.Gradient {
	case CirclyGradient:
		sh1 := 1 - math.Sqrt(1-math.Min(1, y*y/(r*r)))
		d := math.Sqrt(x*x+y*y) / r
		sh2 := math.Abs(y) / r
		sh = (1-d)*sh1 + d*sh2
	case DistanceGradient:
		sh = math.Abs(y / r)
		lighten = 255
	default:
		panic("unknown gradient")
	}

	if y > 0 {
		return DarkenRGBA(col, uint8(255*sh*coloring.Shading))
	} else {
		return LightenRGBA(col, uint8(lighten*sh*coloring.Shading))
	}
}

func TopHalfCircleF(img *image.RGBA, cx, cy, r float64, col Color, coloring ColoringParameters) {
	circleImpl(img, int(cx+.5), int(cy+.5), int(r+.5), col, true, coloring)
}

func CircleF(img *image.RGBA, cx, cy, r float64, col Color, coloring ColoringParameters) {
	Circle(img, int(cx+.5), int(cy+.5), int(r+.5), col, coloring)
}

func Circle(img *image.RGBA, cx, cy, r int, col Color, coloring ColoringParameters) {
	circleImpl(img, cx, cy, r, col, false, coloring)
}

func circleImpl(img *image.RGBA, cx, cy, r int, col Color, topHalfOnly bool, coloring ColoringParameters) {
	colrgba := color.RGBA{col.R, col.G, col.B, 255}
	imgsize := img.Bounds().Dx()
	if cx < -r || cy < -r || cx-r > imgsize || cy-r > imgsize {
		return
	}
	f := 1 - r
	ddF_x := 1
	ddF_y := -2 * r
	x := 0
	y := r

	fill := func(left, right, y int) {
		left += cx
		right += cx

		y += cy
		if left < 0 {
			left = 0
		}
		if right >= imgsize {
			right = imgsize - 1
		}

		for x := left; x <= right; x++ {
			thiscol := CircleShadingRGBA(float64(x-cx), float64(y-cy), float64(r), colrgba, coloring)
			img.SetRGBA(x, y, thiscol)
		}
	}

	fill(-r, r, 0)

	for x < y {
		if f >= 0 {
			y--
			ddF_y += 2
			f += ddF_y
		}
		x++
		ddF_x += 2
		f += ddF_x
		fill(-x, x, -y)
		fill(-y, y, -x)
		if !topHalfOnly {
			fill(-x, x, y)
			fill(-y, y, x)
		}
	}
}

func between(v, min, max int) int {
	if min > max {
		min, max = max, min
	}
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}
func round(v float64) int {
	return int(v + .5)
}

func sqr(x float64) float64 {
	return x * x
}

type ConnectedSpheresTracer struct {
	xmin, xmax, ymin, ymax         int
	v1, v2, v3, a1, a2, a3, ra, dr float64
	col1, col2                     Color
}

func (t *ConnectedSpheresTracer) GetBounds() image.Rectangle {
	return image.Rect(t.xmin, t.ymin, t.xmax, t.ymax)
}

func (t *ConnectedSpheresTracer) Trace(x, y int) (bool, float64, Point3d, Color) {
	return t.traceImpl(x, y, false)
}
func (t *ConnectedSpheresTracer) traceImpl(x, y int, backside bool) (bool, float64, Point3d, Color) {
	p1 := float64(x)
	p2 := float64(y)

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

	return true, p3, Point3d{d1, d2, d3}, MixColors(t.col1, t.col2, math.Max(0, math.Min(1, f)))
}

func (t *ConnectedSpheresTracer) TraceDeep(x, y int) (bool, TraceIntervals) {
	ok1, z1, dir1, col1 := t.traceImpl(x, y, false)
	ok2, z2, dir2, col2 := t.traceImpl(x, y, true)
	if ok1 {
		if !ok2 {
			panic("Huh? That makes no sense")
		}
		return true, TraceIntervals{
			TraceInterval{
				Start: TraceResult{z1,dir1,col1},
				End: TraceResult{z2,dir2,col2},
			},
		}
	}
	return false, TraceIntervals{}
}

func NewConnectedSpheresTracer(img *image.RGBA, wv WorldView, cx1, cy1, cz1, r1 float64, col1 Color, cx2, cy2, cz2, r2 float64, col2 Color) *ConnectedSpheresTracer {

	t := &ConnectedSpheresTracer{}

	size := img.Bounds().Dx()
	t.xmin = between(round(cx1-r1), 0, round(cx2-r2))
	t.xmax = between(round(cx1+r1), round(cx2+r2), size)
	t.ymin = between(round(cy1-r1), 0, round(cy2-r2))
	t.ymax = between(round(cy1+r1), round(cy2+r2), size)

	t.v1 = float64(cx2 - cx1)
	t.v2 = float64(cy2 - cy1)
	t.v3 = float64(cz2 - cz1)

	t.a1 = float64(cx1)
	t.a2 = float64(cy1)
	t.a3 = float64(cz1)

	t.ra = float64(r1)
	t.dr = float64(r2 - r1)

	t.col1 = col1
	t.col2 = col2

	return t

}
