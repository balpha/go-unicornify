package unicornify

import (
	"math"

	. "github.com/balpha/go-unicornify/unicornify/core"
	. "github.com/balpha/go-unicornify/unicornify/elements"
)

type Unicorn struct {
	Figure
	Head, Snout, Shoulder, Butt, HornOnset, HornTip *Ball
	EyeLeft, EyeRight, PupilLeft, PupilRight        *Ball
	TailStart, TailEnd                              *Ball
	BrowLeftInner, BrowLeftMiddle, BrowLeftOuter    *Ball
	BrowRightInner, BrowRightMiddle, BrowRightOuter *Ball
	Tail                                            *Bone
	Legs                                            [4]Leg
	Hairs                                           *Figure
	EarLeft, EarRight                               *Figure
}

var red = Color{255, 0, 0}
var blue = Color{0, 0, 255}

func NewUnicorn(data UnicornData) *Unicorn {
	u := &Unicorn{}

	u.Head = NewBall(0, 0, 0, data.HeadSize, data.Color("Body", 60))
	u.Snout = NewBall(-25, 60, 0, data.SnoutSize, data.Color("Body", 80))
	u.Snout.SetDistance(data.SnoutLength, *u.Head)
	u.Shoulder = NewBall(80, 120, 0, data.ShoulderSize, data.Color("Body", 40))
	u.Butt = NewBall(235, 155, 0, data.ButtSize, data.Color("Body", 40))
	u.HornOnset = NewBall(-22, -10, 0, data.HornOnsetSize, data.Color("Horn", 70))
	u.HornOnset.MoveToSphere(*u.Head)

	tipPos := u.HornOnset.Center.Plus(Vector{-10, 0, 0})
	u.HornTip = NewBall(tipPos[0], tipPos[1], tipPos[2], data.HornTipSize, data.Color("Horn", 90))
	u.HornTip.SetDistance(data.HornLength, *u.HornOnset)
	u.HornTip.RotateAround(*u.HornOnset, data.HornAngle, 2)

	u.makeEyes(data)
	u.makeEars(data)
	u.makeLegs(data)
	data.PoseKind(u, data.PosePhase)
	for _, leg := range u.Legs {
		leg.MoveHoofTo(leg.Hoof.Center)
	}

	u.makeMane(data)

	u.TailStart = NewBallP(u.Butt.Center.Plus(Vector{10, -10, 0}), data.TailStartSize, data.Color("Hair", 80))
	u.TailStart.MoveToSphere(*u.Butt)
	u.TailEnd = NewBallP(u.TailStart.Center.Plus(Vector{10, 0, 0}), data.TailEndSize, data.Color("Hair", 60))
	u.TailEnd.SetDistance(data.TailLength, *u.TailStart)
	u.TailEnd.RotateAround(*u.TailStart, data.TailAngle, 2)
	u.Tail = NewNonLinBone(u.TailStart, u.TailEnd, nil, gammaFuncTimes(data.TailGamma, 0.3))

	eyecurve := gammaFunc(1.5)

	u.Add(
		NewBone(u.Snout, u.Head),
		NewBone(u.HornOnset, u.HornTip),
		u.EyeLeft, u.EyeRight,
		u.PupilLeft, u.PupilRight,
		NewNonLinBone(u.BrowLeftInner, u.BrowLeftMiddle, nil, eyecurve),
		NewNonLinBone(u.BrowLeftMiddle, u.BrowLeftOuter, nil, eyecurve),
		NewNonLinBone(u.BrowRightInner, u.BrowRightMiddle, nil, eyecurve),
		NewNonLinBone(u.BrowRightMiddle, u.BrowRightOuter, nil, eyecurve),
		u.EarLeft, u.EarRight,
	)

	for b := range u.BallSet() {
		b.RotateAround(*u.Head, data.FaceTilt, 0)
	}

	u.Add(NewBone(u.Head, u.Shoulder))
	u.Add(u.Hairs)

	for b := range u.BallSet() {
		b.RotateAround(*u.Shoulder, data.NeckTilt, 1)
	}

	u.Add(
		NewBone(u.Shoulder, u.Butt),
		u.Tail,
	)

	for _, l := range u.Legs {
		u.Add(l.Calf, l.Shin)
	}
	return u
}

func (u *Unicorn) makeEyes(data UnicornData) {
	u.EyeLeft = NewBall(-10, 3, -5, data.EyeSize, Color{255, 255, 255})
	u.EyeLeft.SetGap(5, *u.Head)
	u.EyeRight = NewBall(-10, 3, 5, data.EyeSize, Color{255, 255, 255})
	u.EyeRight.SetGap(5, *u.Head)

	u.PupilLeft = NewBallP(u.EyeLeft.Center.Plus(Vector{-1, 0, 0}), data.PupilSize, Color{0, 0, 0})
	u.PupilLeft.MoveToSphere(*u.EyeLeft)
	u.PupilRight = NewBallP(u.EyeRight.Center.Plus(Vector{-1, 0, 0}), data.PupilSize, Color{0, 0, 0})
	u.PupilRight.MoveToSphere(*u.EyeRight)

	moodDelta := data.BrowMood * 3

	u.BrowLeftInner = NewBallP(u.EyeLeft.Center.Plus(Vector{0, -10, data.BrowLength}), data.BrowSize, data.Color("Hair", 50))
	u.BrowLeftInner.SetGap(5+moodDelta, *u.EyeLeft)
	u.BrowLeftMiddle = NewBallP(u.EyeLeft.Center.Plus(Vector{0, -10, 0}), data.BrowSize, data.Color("Hair", 70))
	u.BrowLeftMiddle.SetGap(5+data.BrowLength, *u.EyeLeft)
	u.BrowLeftOuter = NewBallP(u.EyeLeft.Center.Plus(Vector{0, -10, -data.BrowLength}), data.BrowSize, data.Color("Hair", 60))
	u.BrowLeftOuter.SetGap(5-moodDelta, *u.EyeLeft)

	u.BrowRightInner = NewBallP(u.EyeRight.Center.Plus(Vector{0, -10, -data.BrowLength}), data.BrowSize, data.Color("Hair", 50))
	u.BrowRightInner.SetGap(5+moodDelta, *u.EyeRight)
	u.BrowRightMiddle = NewBallP(u.EyeRight.Center.Plus(Vector{0, -10, 0}), data.BrowSize, data.Color("Hair", 70))
	u.BrowRightMiddle.SetGap(5+data.BrowLength, *u.EyeRight)
	u.BrowRightOuter = NewBallP(u.EyeRight.Center.Plus(Vector{0, -10, data.BrowLength}), data.BrowSize, data.Color("Hair", 60))
	u.BrowRightOuter.SetGap(5-moodDelta, *u.EyeRight)
}

func (u *Unicorn) makeEar(data UnicornData, side float64) *Figure {
	color := data.Color("Body", 60)
	baseRadius := data.HeadSize * 0.35
	base := NewBall(0, -10, 10*side, baseRadius, data.Color("Body", 50))
	base.SetGap(-baseRadius, *u.Head)
	tip := NewBallP(base.Center.Plus(Vector{0, -data.EarLength, 0}), 4, data.Color("Body", 70))

	baseInner := NewBallP(base.Center, base.Radius-2, base.Color)
	tipInner := NewBallP(tip.Center.Plus(Vector{-tip.Radius, 0, 0}), tip.Radius-2, tip.Color)

	box := NewBone(
		NewBallP(base.Center.Plus((Vector{2 * baseRadius, 2 * baseRadius, 0})), 2*baseRadius, color),
		NewBallP(tip.Center.Plus((Vector{2 * baseRadius, -2 * baseRadius, 0})), 2*baseRadius, color),
	)
	result := &Figure{}
	result.Add(NewIntersection(
		NewDifference(
			NewBone(base, tip),
			NewBone(baseInner, tipInner),
		),
		box,
	))
	for b := range result.BallSet() {
		frontBack := -45.0 - 45*Sqr(data.BrowMood)
		upDown := -20 - 10*data.BrowMood
		b.RotateAround(*base, frontBack*side*DEGREE, 1)
		b.RotateAround(*base, upDown*side*DEGREE, 0)
	}
	return result
}

func (u *Unicorn) makeEars(data UnicornData) {
	u.EarLeft = u.makeEar(data, -1)

	u.EarRight = u.makeEar(data, 1)
}

func (u *Unicorn) makeMane(data UnicornData) {
	u.Hairs = &Figure{}

	hairTop := NewBallP(u.Head.Center.Plus(Vector{10, -5, 0}), 5, Color{})
	hairTop.MoveToSphere(*u.Head)
	hairBottom := NewBallP(u.Shoulder.Center.Plus(Vector{10, -15, 0}), 5, Color{})
	hairBottom.MoveToSphere(*u.Shoulder)
	hairSpan := hairBottom.Center.Plus(hairTop.Center.Neg())
	for i := 0; i < len(data.HairStarts); i++ {
		p := hairTop.Center.Plus(hairSpan.Times(data.HairStarts[i] / 100.0))
		hairStart := NewBallP(p, 5, data.Color("Hair", 60))
		endPoint := Vector{
			p.X() + data.HairLengths[i],
			p.Y(),
			p.Z() + data.HairStraightnesses[i],
		}
		hairEnd := NewBallP(endPoint, 2, data.Color("Hair", data.HairTipLightnesses[i]))
		hairEnd.RotateAround(*hairStart, -data.HairAngles[i], 2)
		hair := NewNonLinBone(hairStart, hairEnd, gammaFuncTimes(data.HairGammas[i], 0.2), gammaFuncTimes(1/data.HairGammas[i], 0.2))
		u.Hairs.Add(hair)
	}
}

func gammaFunc(gamma float64) func(float64) float64 {
	return gammaFuncTimes(gamma, 1)
}

func gammaFuncTimes(gamma, t float64) func(float64) float64 {
	return func(x float64) float64 {
		return t*math.Pow(x, gamma) + (1-t)*x
	}
}

func (u *Unicorn) makeLegs(data UnicornData) {
	// front
	for i, z := range [2]float64{-25, 25} {
		hip := NewBall(55, 160, z, 25, data.Color("Body", 40))
		knee := NewBall(35, 254, z, 9, data.Color("Body", 70))
		hoof := NewBall(55, 310, z, 11, data.Color("Body", 45))
		hip.MoveToSphere(*u.Shoulder)
		u.Legs[i] = NewLeg(hip, knee, hoof, true, z < 0)
	}
	// back
	for i, z := range [2]float64{-25, 25} {
		hip := NewBall(225, 190, z, 25, data.Color("Body", 40))
		knee := NewBall(230, 265, z, 9, data.Color("Body", 70))
		hoof := NewBall(220, 310, z, 11, data.Color("Body", 45))
		hip.MoveToSphere(*u.Butt)
		u.Legs[i+2] = NewLeg(hip, knee, hoof, false, z < 0)
	}
}

type Leg struct {
	Hip        *Ball
	Knee       *Ball
	Hoof       *Ball
	Calf       *Bone
	Shin       *Bone
	CalfLength float64
	ShinLength float64
	IsFront    bool
	IsLeft     bool
}

func (l *Leg) MoveHoofTo(hoofPos Vector) {
	var dir Vector
	delta := hoofPos.Minus(l.Hip.Center)
	var sideFactor float64
	if l.IsLeft {
		sideFactor = -1
	} else {
		sideFactor = 1
	}

	if l.IsFront {
		if delta.X() >= 0 {
			//hoof towards back
			dir = Vector{-1, 1, 0}
		} else {
			// hoof towards front
			slope := math.Abs(delta.Y()) / math.Abs(delta.X())
			if slope < 0.3 {
				// more up than outward
				dir = Vector{0, 0, sideFactor}
			} else if slope > 0.6 {
				dir = Vector{-1, -1, 0}
			} else {
				f := (slope - 0.3) / (0.6 - 0.3)
				dir = Vector{-1 * f, -1 * f, sideFactor * (1 - f)}
			}
		}
	} else { // back
		if delta.X() <= 0 {
			//hoof towards front
			dir = Vector{1, 1, 0}
		} else {
			// hoof towards front
			slope := math.Abs(delta.Y()) / math.Abs(delta.X())
			if slope < 0.3 {
				// more up than outward
				dir = Vector{0, 0, sideFactor}
			} else if slope > 0.6 {
				dir = Vector{1, -1, 0}
			} else {
				f := (slope - 0.3) / (0.6 - 0.3)
				dir = Vector{1 * f, -1 * f, sideFactor * (1 - f)}
			}
		}
	}
	kneePos := l.Hip.Center.Plus(Joint(hoofPos.Minus(l.Hip.Center), l.CalfLength, l.ShinLength, dir))
	l.Hoof.Center = hoofPos
	l.Knee.Center = kneePos
}

func NewLeg(hip, knee, hoof *Ball, isFront, isLeft bool) Leg {
	return Leg{
		Hip:        hip,
		Knee:       knee,
		Hoof:       hoof,
		Calf:       NewBone(hip, knee),
		CalfLength: knee.Center.Minus(hip.Center).Length(),
		Shin:       NewBone(knee, hoof),
		ShinLength: hoof.Center.Minus(knee.Center).Length(),
		IsFront:    isFront,
		IsLeft:     isLeft,
	}
}
