package unicornify

import (
	"image"
	"math"
)

type Quad struct {
	Balls   [4]*Ball
	Shading float64
}

func NewQuad(b1, b2, b3, b4 *Ball) *Quad {
	return NewShadedQuad(b1, b2, b3, b4, defaultShading)
}

func NewShadedQuad(b1, b2, b3, b4 *Ball, shading float64) *Quad {
	return &Quad{[4]*Ball{b1, b2, b3, b4}, shading}
}

func (q *Quad) Project(wv WorldView) {
	for _, b := range q.Balls {
		b.Project(wv)
	}
}

var Foo = true

func (q Quad) Draw(img *image.RGBA, wv WorldView, shading bool) {

	sh := q.Shading
	if !shading {
		sh = 0
	}
	b1 := q.Balls[0]
	b2 := q.Balls[1]
	b3 := q.Balls[2]
	b4 := q.Balls[3]

	steps12 := math.Max(math.Abs(b2.Projection.X()-b1.Projection.X()), math.Abs(b2.Projection.Y()-b1.Projection.Y())) + 1
	steps23 := math.Max(math.Abs(b3.Projection.X()-b2.Projection.X()), math.Abs(b3.Projection.Y()-b2.Projection.Y())) + 1
	steps34 := math.Max(math.Abs(b4.Projection.X()-b3.Projection.X()), math.Abs(b4.Projection.Y()-b3.Projection.Y())) + 1
	steps41 := math.Max(math.Abs(b1.Projection.X()-b4.Projection.X()), math.Abs(b1.Projection.Y()-b4.Projection.Y())) + 1

	stepsU := math.Max(steps12, steps34)
	stepsV := math.Max(steps23, steps41)

	var fromBalls, toBalls [2]*Ball

	if stepsU > stepsV {
		fromBalls = [2]*Ball{b1, b4}
		toBalls = [2]*Ball{b2, b3}
	} else {
		fromBalls = [2]*Ball{b1, b2}
		toBalls = [2]*Ball{b4, b3}
	}

	steps, _ := math.Modf(math.Max(stepsU, stepsV))

	cp := ColoringParameters{sh, DistanceGradient}

	for step := float64(0); step <= steps; step++ {
		factor := step / steps
		col1 := MixColors(fromBalls[0].Color, fromBalls[1].Color, factor)
		col2 := MixColors(toBalls[0].Color, toBalls[1].Color, factor)
		x1 := MixFloats(fromBalls[0].Projection.X(), fromBalls[1].Projection.X(), factor)
		x2 := MixFloats(toBalls[0].Projection.X(), toBalls[1].Projection.X(), factor)
		y1 := MixFloats(fromBalls[0].Projection.Y(), fromBalls[1].Projection.Y(), factor)
		y2 := MixFloats(toBalls[0].Projection.Y(), toBalls[1].Projection.Y(), factor)
		r1 := MixFloats(fromBalls[0].ProjectionRadius, fromBalls[1].ProjectionRadius, factor)
		r2 := MixFloats(toBalls[0].ProjectionRadius, toBalls[1].ProjectionRadius, factor)
		ConnectCirclesF(img, x1+wv.Shift[0], y1+wv.Shift[1], r1, col1, x2+wv.Shift[0], y2+wv.Shift[1], r2, col2, cp)
	}
}

func (q Quad) Bounding() image.Rectangle {
	result := q.Balls[0].Bounding()
	for _, b := range q.Balls[1:] {
		result = result.Union(b.Bounding())
	}
	return result
}

func (q *Quad) Sort(wv WorldView) {
	maxZi := -1
	maxZ := 0.0
	for i, b := range q.Balls {
		z := b.Projection.Z()
		if maxZi == -1 || z > maxZ {
			maxZi = i
			maxZ = z
		}
	}
	i := 0
	oldBalls := q.Balls
	q.Balls = [4]*Ball{}
	for i < 4 {
		q.Balls[i] = oldBalls[(i+maxZi)%4]
		i++
	}
}
