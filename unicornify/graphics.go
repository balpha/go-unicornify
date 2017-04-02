package unicornify

import (
	. "bitbucket.org/balpha/go-unicornify/unicornify/core"
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
