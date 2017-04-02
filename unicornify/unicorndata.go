package unicornify

import (
	. "bitbucket.org/balpha/go-unicornify/unicornify/core"
	"bitbucket.org/balpha/gopyrand"
)

type UnicornData struct {
	HeadSize           float64
	SnoutSize          float64
	ShoulderSize       float64
	SnoutLength        float64
	ButtSize           float64
	BodyHue            int
	BodySat            int
	HornHue            int
	HornSat            int
	HornOnsetSize      float64
	HornTipSize        float64
	HornLength         float64
	HornAngle          float64 // 0 means straight in x-direction, >0 means upwards
	EyeSize            float64
	IrisSize           float64 // no longer used
	IrisHue            int     // no longer used
	IrisSat            int     // no longer used
	PupilSize          float64
	HairHue            int
	HairSat            int
	HairStarts         []float64
	HairGammas         []float64
	HairLengths        []float64
	HairAngles         []float64
	HairStraightnesses []float64 // for lack of a better word -- this is just the z offsets of the tip
	HairTipLightnesses []int
	TailStartSize      float64
	TailEndSize        float64
	TailLength         float64
	TailAngle          float64
	TailGamma          float64
	BrowSize           float64
	BrowLength         float64
	BrowMood           float64 // from -1 (angry) to 1 (astonished)

	PoseKind      func(*Unicorn, float64)
	PoseKindIndex int
	PosePhase     float64

	NeckTilt float64
	FaceTilt float64
}

func (d UnicornData) Color(name string, lightness int) Color {
	return ColorFromData(d, name, lightness)
}

func (d *UnicornData) Randomize1(rand *pyrand.Random) {
	d.BodyHue = rand.RandInt(0, 359)
	d.BodySat = rand.RandInt(50, 100)
	d.HornHue = (d.BodyHue + rand.RandInt(60, 300)) % 360
	d.HornSat = rand.RandInt(50, 100)
	d.SnoutSize = float64(rand.RandInt(8, 30))
	d.SnoutLength = float64(rand.RandInt(70, 110))
	d.HeadSize = float64(rand.RandInt(25, 40))
	d.ShoulderSize = float64(rand.RandInt(40, 60))
	d.ButtSize = float64(rand.RandInt(30, 60))
	d.HornOnsetSize = float64(rand.RandInt(6, 12))
	d.HornTipSize = float64(rand.RandInt(3, 6))
	d.HornLength = float64(rand.RandInt(50, 100))
	d.HornAngle = float64(rand.RandInt(10, 60)) * DEGREE
	d.EyeSize = float64(rand.RandInt(8, 12))
	d.IrisSize = float64(rand.RandInt(3, 6))
	d.IrisHue = rand.RandInt(70, 270)
	d.IrisSat = rand.RandInt(40, 70)
	d.PupilSize = float64(rand.RandInt(2, 5))
	_ = rand.RandInt(0, 60)
	d.HairHue = (d.BodyHue + rand.RandInt(60, 300)) % 360
	d.HairSat = rand.RandInt(60, 100)
	hairCount := rand.RandInt(12, 30) * 2
	d.HairStarts = make([]float64, hairCount)
	d.HairGammas = make([]float64, hairCount)
	d.HairLengths = make([]float64, hairCount)
	d.HairAngles = make([]float64, hairCount)
	d.HairTipLightnesses = make([]int, hairCount)
	d.HairStraightnesses = make([]float64, hairCount)
	d.MakeHair1(rand, 0, hairCount/2)

}

func (d *UnicornData) Randomize2(rand *pyrand.Random) {
	hairCount := len(d.HairStarts)
	d.MakeHair2(rand, 0, hairCount/2)

	d.TailStartSize = float64(rand.RandInt(4, 10))
	d.TailEndSize = float64(rand.RandInt(10, 20))
	d.TailLength = float64(rand.RandInt(100, 150))
	d.TailAngle = float64(rand.RandInt(-20, 45)) * DEGREE
	d.TailGamma = .1 + rand.Random()*6
	d.BrowSize = float64(rand.RandInt(2, 4))
	d.BrowLength = 2 + rand.Random()*3
	d.BrowMood = 2*rand.Random() - 1
	intNeckTilt := rand.RandInt(-30, 30)
	d.NeckTilt = float64(intNeckTilt) * DEGREE
	a, b := intNeckTilt/3, intNeckTilt/4
	if a > b {
		a, b = b, a
	}
	d.FaceTilt = float64(rand.RandInt(a, b)) * DEGREE

}

func (d *UnicornData) Randomize3(rand *pyrand.Random) {
	d.PoseKindIndex = rand.Choice(len(Poses))
	d.PoseKind = Poses[d.PoseKindIndex]
	d.PosePhase = rand.Random()
}
func (d *UnicornData) Randomize4(rand *pyrand.Random) {
	halfCount := len(d.HairStarts) / 2
	d.MakeHair1(rand, halfCount, halfCount)
	d.MakeHair2(rand, halfCount, halfCount)
}

func (d *UnicornData) MakeHair1(rand *pyrand.Random, start, count int) {
	for i := start; i < start+count; i++ {
		d.HairStarts[i] = float64(rand.RandInt(-20, 100))
	}
	for i := start; i < start+count; i++ {
		d.HairGammas[i] = .3 + rand.Random()*3
	}
	for i := start; i < start+count; i++ {
		d.HairLengths[i] = float64(rand.RandInt(80, 150))
	}
	for i := start; i < start+count; i++ {
		d.HairAngles[i] = float64(rand.RandInt(0, 60)) * DEGREE
	}
}

func (d *UnicornData) MakeHair2(rand *pyrand.Random, start, count int) {
	for i := start; i < start+count; i++ {
		d.HairTipLightnesses[i] = rand.RandInt(40, 85)
	}
	for i := start; i < start+count; i++ {
		d.HairStraightnesses[i] = float64(rand.RandInt(-40, 40))
	}
}
