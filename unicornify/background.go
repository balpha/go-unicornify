package unicornify

import (
	"bitbucket.org/balpha/gopyrand"
	"code.google.com/p/draw2d/draw2d"
	"image"
	"math"
)

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

func (d BackgroundData) Draw(im *image.RGBA) {
	size := im.Bounds().Dx()
	fsize := float64(size - 1)

	// sky

	horizonPixels := int(float64(size) * d.Horizon)
	for y := 0; y < horizonPixels; y++ {
		col := MixColors(d.Color("Sky", 60), d.Color("Sky", 10), float64(y)/fsize)
		for x := 0; x < size; x++ {
			im.Set(x, y, col)
		}
	}

	// ground

	land1 := d.Color("Land", d.LandLight)
	land2 := d.Color("Land", d.LandLight/2)

	for x := 0; x < size; x++ {
		col := MixColors(land1, land2, float64(x)/fsize)

		for y := horizonPixels; y < size; y++ {
			im.Set(x, y, col)
		}
	}

	// rainbow

	ctx := draw2d.NewGraphicContext(im)

	bandPixWidth := d.RainbowBandWidth * fsize
	rainbowCenterX := fsize * (d.RainbowFoot + d.RainbowDir*d.RainbowHeight)
	outerRadius := d.RainbowHeight * fsize

	// except for the innermost (purple) one, the bands are drawn a bit wider so they overlap
	// and thus we see no sub-pixel gap
	overlap := math.Min(2, bandPixWidth)
	ctx.SetLineWidth(d.RainbowBandWidth*fsize + overlap)
	for band := 0; band < 7; band++ {
		if band == 6 {
			ctx.SetLineWidth(d.RainbowBandWidth * fsize)
			overlap = 0
		}
		col := hsl2col(band*45, 100, 50)
		ctx.BeginPath()
		ctx.SetStrokeColor(col)
		rad := outerRadius - float64(band)*bandPixWidth - overlap/2
		ctx.ArcTo(rainbowCenterX, float64(horizonPixels), rad, rad, 0, -180*DEGREE)
		ctx.Stroke()
	}

	// clouds

	for i, pos := range d.CloudPositions {
		sizes := d.CloudSizes[i]
		drawCloud(ctx, fsize*pos[0], fsize*pos[1], fsize*sizes[0], fsize*sizes[0]*sizes[1], d.Color("Sky", d.CloudLightnesses[i]))
	}

}

func drawCloud(ctx draw2d.GraphicContext, x, y, size1, size2 float64, col Color) {
	ctx.SetFillColor(col)
	ctx.BeginPath()
	ctx.ArcTo(x-2*size1, y-size1, size1, size1, 0, deg360)
	ctx.Fill()
	ctx.BeginPath()
	ctx.ArcTo(x+2*size1, y-size1, size1, size1, 0, deg360)
	ctx.Fill()
	ctx.BeginPath()
	ctx.ArcTo(x, y-size1, size2, size2, 0, -180*DEGREE)
	ctx.Fill()
	ctx.BeginPath()
	ctx.MoveTo(x-2*size1, y-size1-1)
	ctx.LineTo(x+2*size1, y-size1-1)
	ctx.LineTo(x+2*size1, y)
	ctx.LineTo(x-2*size1, y)
	ctx.Close()
	ctx.Fill()

}
