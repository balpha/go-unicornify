package unicornify

import (
	"bitbucket.org/balpha/gopyrand"
)

type BirdData struct {
	HeadSize      float64
	ChestSize     float64
	ButtSize      float64
	TailTipSize   float64
	TailLength    float64
	BeakOnsetSize float64
	BeakTipSize   float64
	BeakLength    float64
	WingLength    float64
	WingAngle     float64
	EyeSize       float64
	TailAngle     float64
	BodyAngle     float64
	BodyHue       int
	BodySat       int
	HeadHue       int
	HeadSat       int
	BeakHue       int
	BeakSat       int
}

var BaseBird = BirdData{
	HeadSize:      24,
	ChestSize:     33,
	ButtSize:      25,
	TailTipSize:   9,
	TailLength:    120,
	BeakOnsetSize: 6,
	BeakTipSize:   3,
	BeakLength:    18,
	WingLength:    128,
	WingAngle:     0 * DEGREE,
	EyeSize:       5,
	TailAngle:     0 * DEGREE,
	BodyAngle:     0 * DEGREE,
}

func (d *BirdData) Randomize(rand *pyrand.Random) {
	d.HeadSize = 15 + rand.Random()*15
	d.ChestSize = d.HeadSize + rand.Random()*15 + 5
	d.ButtSize = 20 + rand.Random()*15
	d.TailTipSize = 5 + rand.Random()*10
	d.TailLength = 40 + rand.Random()*100
	d.BeakOnsetSize = d.HeadSize * (0.1 + 0.3*rand.Random())
	d.BeakTipSize = d.BeakTipSize * (0.2 + 0.6*rand.Random())
	d.BeakLength = 10 + rand.Random()*20
	d.WingLength = 80 + rand.Random()*120
	d.EyeSize = d.HeadSize * (0.1 + rand.Random()*0.1)
	d.WingAngle = (20 + rand.Random()*50) * DEGREE
	d.TailAngle = rand.Random() * 40 * DEGREE
	d.BodyAngle = rand.Random() * 45 * DEGREE

	d.BodyHue = rand.RandInt(0, 359)
	d.BodySat = rand.RandInt(30, 100)
	d.HeadHue = rand.RandInt(0, 359)
	d.HeadSat = rand.RandInt(30, 100)
	d.BeakHue = rand.RandInt(0, 359)
	d.BeakSat = rand.RandInt(20, 40)

}

func (d BirdData) Color(name string, lightness int) Color {
	return ColorFromData(d, name, lightness)
}
