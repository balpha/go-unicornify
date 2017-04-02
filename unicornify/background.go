package unicornify

import (
	. "bitbucket.org/balpha/go-unicornify/unicornify/core"
	"bitbucket.org/balpha/gopyrand"
	"image"
	"image/color"
	"math"
)

type Point2d [2]float64

type BackgroundData struct {
	SkyHue           int
	SkySat           int
	LandHue          int
	LandSat          int
	Horizon          float64
	RainbowFoot      float64
	RainbowDir       float64 // +1 or -1
	RainbowHeight    float64
	RainbowBandWidth float64
	CloudPositions   []Point2d
	CloudSizes       []Point2d // not actually any kind of point
	CloudLightnesses []int
	LandLight        int
}

func (d BackgroundData) Color(name string, lightness int) Color {
	return ColorFromData(d, name, lightness)
}

func (d *BackgroundData) Randomize1(rand *pyrand.Random) {
	d.SkyHue = rand.RandInt(0, 359)
	d.SkySat = rand.RandInt(30, 70)
	d.LandHue = rand.RandInt(0, 359)
	d.LandSat = rand.RandInt(20, 60)
	d.Horizon = .5 + rand.Random()*.2
	d.RainbowFoot = .2 + rand.Random()*.6
	d.RainbowDir = float64(rand.Choice(2)*2 - 1)
	d.RainbowHeight = .5 + rand.Random()*1.5
	d.RainbowBandWidth = .01 + rand.Random()*.02
	d.LandLight = rand.RandInt(20, 50)
}

func (d *BackgroundData) Randomize2(rand *pyrand.Random) {
	cloudCount := rand.RandInt(1, 3)
	d.CloudPositions = make([]Point2d, cloudCount)
	d.CloudSizes = make([]Point2d, cloudCount)
	d.CloudLightnesses = make([]int, cloudCount)

	for i := 0; i < cloudCount; i++ {
		d.CloudPositions[i] = Point2d{
			rand.Random(),
			(.3 + rand.Random()*.6) * d.Horizon,
		}
	}
	for i := 0; i < cloudCount; i++ {
		d.CloudSizes[i] = Point2d{
			rand.Random()*.04 + .02,
			rand.Random()*.7 + 1.3,
		}
	}
	for i := 0; i < cloudCount; i++ {
		d.CloudLightnesses[i] = rand.RandInt(75, 90)
	}
}

func (d BackgroundData) Draw(im *image.RGBA, shading bool) {
	size := im.Bounds().Dx()
	fsize := float64(size - 1)

	// sky

	horizonPixels := int(float64(size) * d.Horizon)
	for y := 0; y < horizonPixels; y++ {
		col := MixColors(d.Color("Sky", 60), d.Color("Sky", 10), float64(y)/fsize)
		for x := 0; x < size; x++ {
			im.SetRGBA(x, y, col.ToRGBA())
		}
	}

	// ground

	land1 := d.Color("Land", d.LandLight)
	land2 := d.Color("Land", d.LandLight/2)

	for x := 0; x < size; x++ {
		col := MixColors(land1, land2, float64(x)/fsize)

		for y := horizonPixels; y < size; y++ {
			im.SetRGBA(x, y, col.ToRGBA())
		}
	}

	// rainbow

	bandPixWidth := d.RainbowBandWidth * fsize
	rainbowCenterX := fsize * (d.RainbowFoot + d.RainbowDir*d.RainbowHeight)
	outerRadius := d.RainbowHeight * fsize

	drawRainbow(im, int(rainbowCenterX+.5), horizonPixels, int(outerRadius+.5), bandPixWidth)

	// clouds

	for i, pos := range d.CloudPositions {

		sizes := d.CloudSizes[i]
		drawCloud(im, fsize*pos[0], fsize*pos[1], fsize*sizes[0], fsize*sizes[0]*sizes[1], d.Color("Sky", d.CloudLightnesses[i]), shading)
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

func drawRainbow(img *image.RGBA, cx, cy, r int, bandWidth float64) {
	size := img.Bounds().Dx()
	left := between(cx-r, 0, size-1)
	right := between(cx+r, 0, size-1)
	top := between(cy-r, 0, size-1)
	bottom := between(cy, 0, size-1)
	innerRadSquared := int(float64(r) - 7*bandWidth)

	bandCols := [7]color.RGBA{}
	for i := 0; i < 7; i++ {
		col := Hsl2col(i*45, 100, 50)
		bandCols[i] = color.RGBA{col.R, col.G, col.B, 255}
	}

	for x := left; x <= right; x++ {
		dx := x - cx
		for y := top; y <= bottom; y++ {
			dy := y - cy
			dsquared := dx*dx + dy*dy
			if dsquared < innerRadSquared {
				continue
			}
			d := math.Sqrt(float64(dsquared))
			band := (float64(r) - d) / bandWidth
			if band >= 7 || band < 0 {
				continue
			}

			img.SetRGBA(x, y, bandCols[int(band)])
		}
	}

}
func drawCloud(img *image.RGBA, x, y, size1, size2 float64, col Color, shaded bool) {
	shading := 0.0
	if shaded {
		shading = 0.25
	}
	cp := DefaultGradientWithShading(shading)
	CircleF(img, x-2*size1, y-size1, size1, col, cp)
	CircleF(img, x+2*size1, y-size1, size1, col, cp)
	TopHalfCircleF(img, x, y-size1, size2, col, cp)

	xi := int(x + .5)
	yi := int(y + .5)
	size1i := int(size1 + .5)
	right := xi + 2*size1i
	for py := yi - size1i - 1; py <= yi; py++ {
		for px := xi - 2*size1i; px <= right; px++ {
			thiscol := col
			if shaded {
				dy := float64(py - (yi - size1i - 1))
				thiscolr := CircleShadingRGBA(0, dy, size1, col.ToRGBA(), cp)
				img.SetRGBA(px, py, thiscolr)
			} else {
				img.SetRGBA(px, py, thiscol.ToRGBA())
			}

		}
	}
}
