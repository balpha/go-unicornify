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

	p1 := b1.Projection
	p2 := b2.Projection

	r1 := b1.Projection.Radius
	r2 := b2.Projection.Radius

	c1 := b1.Color
	c2 := b2.Color

	if b.XFunc != nil || b.YFunc != nil {
		bounding := b.Bounding()
		parts := bounding.Dy()
		if bounding.Dx() > bounding.Dy() {
			parts = bounding.Dx()
		}
		parts = roundUp(float64(parts) * SCALE)

		var prevX, prevY, prevR, prevZ, prevFx, prevFy float64
		var prevCol Color

		prevX, prevY = p1.X(), p1.Y()
		prevR = r1
		prevCol = c1
		prevZ = p1.Z()
		result := NewGroupTracer()
		subgroup := NewGroupTracer()
		for i := 1; i <= parts; i++ {
			factor := float64(i) / float64(parts)
			col := MixColors(c1, c2, factor)
			fx, fy := factor, factor
			if f := b.XFunc; f != nil {
				fx = f(fx)
			}
			if f := b.YFunc; f != nil {
				fy = f(fy)
			}
			if i > 1 && i < parts && math.Abs((prevFx/prevFy)/(fx/fy)-1) < 0.02 {
				continue
			}
			prevFx = fx
			prevFy = fy
			x := MixFloats(p1.X(), p2.X(), fx)
			y := MixFloats(p1.Y(), p2.Y(), fy)

			z := MixFloats(p1.Z(), p2.Z(), factor)
			r := MixFloats(r1, r2, factor)
			tracer := NewConnectedSpheresTracer(wv, prevX, prevY, prevZ, prevR, prevCol, x, y, z, r, col /*, cp*/)
			subgroup.Add(tracer)

			if i%5 == 0 || i == parts {
				result.Add(subgroup)
				subgroup = NewGroupTracer()
			}

			prevX, prevY, prevZ, prevR, prevCol = x, y, z, r, col
		}

		return result

	} else {
		tracer := NewConnectedSpheresTracer(wv, p1.X(), p1.Y(), p1.Z(), r1, c1, p2.X(), p2.Y(), p2.Z(), r2, c2 /*, cp*/)
		return tracer
	}
	return nil
}

func (b Bone) Bounding() image.Rectangle {
	return b.Balls[0].Bounding().Union(b.Balls[1].Bounding())
}
