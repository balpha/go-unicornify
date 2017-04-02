package unicornify

import (
	. "bitbucket.org/balpha/go-unicornify/unicornify/core"
	"bitbucket.org/balpha/gopyrand"
	"image"
	"image/color"
	"math"
)

type GrassData struct {
	Seed                            uint32
	RowSeedAdd                      uint32
	Horizon                         float64
	BladeHeightFar, BladeHeightNear float64 // 0-1-based, relative to image width/height
	Wind                            float64
	Color1, Color2                  Color
	MinBottomY                      float64 // pixel
}

func (d *GrassData) Randomize(rand *pyrand.Random) {
	r := rand.RandBits(64)
	d.Seed = r[0]
	d.RowSeedAdd = r[1]
	d.Wind = 1.6*rand.Random() - 0.8
}

type BladeData struct {
	BottomX, BottomY, Height, BottomWidth, TopWidth, CurveStrength float64 // pixel-based
	CurveStart, CurveEnd                                           float64
	Color                                                          Color
	ConstrainImage                                                 *image.RGBA
}

func DrawGrass(img *image.RGBA, d GrassData, wv WorldView, shadowImage *image.RGBA) {
	bd := BladeData{}
	fsize := float64(img.Bounds().Dy())
	for row := uint32(0); bd.BottomY-bd.Height <= fsize; row++ {
		seed := d.Seed + row*d.RowSeedAdd
		rand := pyrand.NewRandom()
		rand.SeedFromUInt32(seed)

		rowf := float64(row) / 100.0

		distf := d.BladeHeightFar / d.BladeHeightNear

		y := (1-distf)*rowf*rowf + distf*rowf
		baseSize := d.BladeHeightFar + rowf*(d.BladeHeightNear-d.BladeHeightFar)
		colstep := 0.2 * baseSize

		bottomY := fsize * (d.Horizon + y*(1-d.Horizon))
		if bottomY < d.MinBottomY {
			continue
		}

		for col := 0.0; col <= 1; col += colstep {
			bd.BottomX = fsize * (col + baseSize*(rand.Random()*0.2-0.1))
			bd.BottomY = bottomY + fsize*baseSize*rand.Random()*0.3
			bd.Height = baseSize * fsize * (0.95 + rand.Random()*0.1)
			bd.BottomWidth = baseSize * fsize * (rand.Random()*0.04 + 0.1)
			bd.TopWidth = baseSize * fsize * (rand.Random() * 0.01)
			bd.CurveStrength = baseSize * fsize * (d.Wind + rand.Random()*0.2)
			bd.CurveStart = rand.Random() * 0.5
			bd.CurveEnd = 0.5 + rand.Random()*0.5
			bd.Color = MixColors(d.Color1, d.Color2, rand.Random())

			if shadowImage != nil {
				s := shadowImage.RGBAAt(Round(bd.BottomX), Round(bd.BottomY))
				if s.A == 255 && s.R < 128 {
					bd.Color = Darken(bd.Color, uint8(128-s.R))
				}
			}

			DrawGrassBlade(img, bd)
		}
	}

}

func DrawGrassBlade(img *image.RGBA, d BladeData) {

	for dy := 0; dy <= Round(d.Height); dy++ {
		f := float64(dy) / d.Height
		curveP := (d.CurveStart + f*(d.CurveEnd-d.CurveStart)) * math.Pi / 2
		curve := math.Sin(curveP) - curveP - (math.Sin(d.CurveStart*math.Pi/2) - d.CurveStart*math.Pi/2)
		width := d.BottomWidth + f*(d.TopWidth-d.BottomWidth)
		left := Round(d.BottomX + curve*d.CurveStrength - width/2)
		right := Round(d.BottomX + curve*d.CurveStrength + width/2)
		y := Round(d.BottomY) - dy
		for x := left; x <= right; x++ {
			if d.ConstrainImage != nil && d.ConstrainImage.At(x, y).(color.RGBA).A == 0 {
				continue
			}
			thiscol := d.Color
			if (d.CurveStrength < 0 && x >= left+(right-left)*2/3) || (d.CurveStrength >= 0 && x <= left+(right-left)*1/3) {
				thiscol = Darken(thiscol, 10)
			}
			img.SetRGBA(x, y, thiscol.ToRGBA())
		}
	}
}
