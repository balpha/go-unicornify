package unicornify

import (
	"image"
	"math"
)

type Bone struct {
	Balls        [2]*Ball
	XFunc, YFunc func(float64) float64 // may be nil
	Shading      float64
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
	return &Bone{[2]*Ball{b1, b2}, xFunc, yFunc, shading}
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

func (b Bone) Draw(img *image.RGBA, wv WorldView, shading bool) {
	b1 := b.Balls[0]
	b2 := b.Balls[1]

	p1 := b1.Projection
	p2 := b2.Projection

	r1 := b1.Radius
	r2 := b2.Radius

	c1 := b1.Color
	c2 := b2.Color

	sh := b.Shading
	if !shading {
		sh = 0
	}
	cp := DefaultGradientWithShading(sh)
	if b.XFunc == nil && b.YFunc == nil {
		ConnectCirclesF(img, p1.X()+wv.Shift[0], p1.Y()+wv.Shift[1], r1, c1, p2.X()+wv.Shift[0], p2.Y()+wv.Shift[1], r2, c2, cp)
		return
	}

	steps := math.Max(math.Abs(p2.X()-p1.X()), math.Abs(p2.Y()-p1.Y()))

	// The centers might be very close, but the radii my be more apart,
	// hence the following step. without it, the eye/iris gradient sometimes
	// only has two or three steps
	steps, _ = math.Modf(math.Max(steps, math.Abs(r2-r1)) + 1)
	var prevX, prevY, prevR float64
	var prevCol Color
	nonlin := b.XFunc != nil || b.YFunc != nil
	for step := float64(0); step <= steps; step++ {
		factor := step / steps
		col := MixColors(c1, c2, factor)
		fx, fy := factor, factor
		if f := b.XFunc; f != nil {
			fx = f(fx)
		}
		if f := b.YFunc; f != nil {
			fy = f(fy)
		}
		x := MixFloats(p1.X(), p2.X(), fx)
		y := MixFloats(p1.Y(), p2.Y(), fy)
		r := MixFloats(r1, r2, factor)

		if nonlin && step > 0 && (math.Abs(x-prevX) > 1.1 || math.Abs(y-prevY) > 1.1) {
			sb1 := &Ball{Projection: Point3d{prevX, prevY, 0}, Radius: prevR, Color: prevCol}
			sb2 := &Ball{Projection: Point3d{x, y, 0}, Radius: r, Color: col}
			NewShadedBone(sb1, sb2, b.Shading).Draw(img, wv, shading)
		} else {
			Circle(img, int(x+wv.Shift[0]+.5), int(y+wv.Shift[1]+.5), int(r+.5), col, cp)
		}
		prevX, prevY, prevR, prevCol = x, y, r, col
	}
}

func (b Bone) Bounding() image.Rectangle {
	return b.Balls[0].Bounding().Union(b.Balls[1].Bounding())
}

func (b *Bone) Sort(wv WorldView) {
	z1 := b.Balls[0].Projection.Z()
	z2 := b.Balls[1].Projection.Z()
	if z1 < z2 {
		b.Balls[0], b.Balls[1] = b.Balls[1], b.Balls[0]
		if b.XFunc != nil {
			b.XFunc = reverse(b.XFunc)
		}
		if b.YFunc != nil {
			b.YFunc = reverse(b.YFunc)
		}
	}

}
