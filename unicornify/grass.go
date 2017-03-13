package unicornify

import (
	"bitbucket.org/balpha/gopyrand"
	"image"
	"math"
	"image/color"
)

type GrassData struct {
	Seed uint32
	RowSeedAdd uint32
	Horizon float64
	BladeHeightFar, BladeHeightNear float64 // 0-1-based, relative to image width/height
	Wind float64
	Color1, Color2 Color
	MinBottomY float64 // pixel
	ConstrainBone *Bone // pixel
}

func (d *GrassData) Randomize(rand *pyrand.Random) {
	r := rand.RandBits(64)
	d.Seed = r[0]
	d.RowSeedAdd = r[1]
	d.Wind = 1.6*rand.Random() - 0.8
}

type BladeData struct {
	BottomX, BottomY, Height, BottomWidth, TopWidth, CurveStrength float64 // pixel-based
	CurveStart, CurveEnd float64
	Color Color
	ConstrainImage *image.RGBA
}

func DrawGrass(img *image.RGBA, d GrassData, wv WorldView) {
	bd := BladeData{}
	if (d.ConstrainBone != nil) {
		mask := image.NewRGBA(img.Bounds())
		cx1 := d.ConstrainBone.Balls[0].Projection.X() + wv.Shift[0]
		cy1 := d.ConstrainBone.Balls[0].Projection.Y() + wv.Shift[1]
		r1 := d.ConstrainBone.Balls[0].Radius
		cx2 := d.ConstrainBone.Balls[1].Projection.X() + wv.Shift[0]
		cy2 := d.ConstrainBone.Balls[1].Projection.Y() + wv.Shift[1]
		r2 := d.ConstrainBone.Balls[1].Radius
		white := Color{255,255,255}
		ConnectCirclesF(mask, cx1, cy1,r1, white, cx2, cy2, r2, white, 0)
		bd.ConstrainImage = mask
	}
	fsize := float64(img.Bounds().Dy())
	for row := uint32(0); bd.BottomY - bd.Height <= fsize; row++ {
		seed := d.Seed + row * d.RowSeedAdd
		rand := pyrand.NewRandom()
		rand.SeedFromUInt32(seed)
		
		rowf := float64(row) / 100.0
		
		distf := d.BladeHeightFar / d.BladeHeightNear
		
		y := (1-distf)*rowf*rowf + distf * rowf
		baseSize := d.BladeHeightFar + rowf * (d.BladeHeightNear - d.BladeHeightFar)
		colstep := 0.2*baseSize
		
		bd.BottomY = fsize*(d.Horizon + y*(1- d.Horizon))
		if bd.BottomY < d.MinBottomY {
			continue
		}
		
		for col:=0.0; col <= 1; col+=colstep {
			bd.BottomX = fsize*(col + baseSize*(rand.Random()*0.2 - 0.1))
			bd.Height = baseSize*fsize*(0.95+rand.Random()*0.1)
			bd.BottomWidth = baseSize*fsize*(rand.Random()*0.04+0.1)
			bd.TopWidth = baseSize*fsize*(rand.Random()*0.01)
			bd.CurveStrength = baseSize*fsize*(d.Wind + rand.Random()*0.2)
			bd.CurveStart = rand.Random() * 0.5
			bd.CurveEnd = 0.5 + rand.Random()*0.5
			bd.Color = MixColors(d.Color1, d.Color2, rand.Random())
			
			DrawGrassBlade(img, bd)
		}
	}			
	
}

func DrawGrassBlade(img *image.RGBA, d BladeData) {
	
	for dy:=0; dy <= round(d.Height); dy++ {
		f := float64(dy) / d.Height
		curveP := (d.CurveStart + f*(d.CurveEnd - d.CurveStart)) * math.Pi / 2
		curve := math.Sin(curveP) - curveP - (math.Sin(d.CurveStart* math.Pi / 2) - d.CurveStart* math.Pi / 2)
		width := d.BottomWidth + f*(d.TopWidth - d.BottomWidth)
		left := round(d.BottomX + curve * d.CurveStrength - width/2)
		right := round(d.BottomX + curve * d.CurveStrength + width/2)
		y := round(d.BottomY) - dy
		for x:=left; x<=right; x++ {
			if d.ConstrainImage != nil && d.ConstrainImage.At(x, y).(color.RGBA).A == 0 {
				continue
			}
			thiscol := d.Color
			if x >= left +(right-left)*2/3 {
				thiscol = Darken(thiscol, 10)
			}
			img.Set(x, y, thiscol)
		}
	}
}