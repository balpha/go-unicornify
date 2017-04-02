package unicornify

/*
type Bird struct {
	Figure
}

var yellow = Color{255, 255, 0}
var green = Color{0, 255, 0}

func NewBird(data BirdData) *Bird {
	b := &Bird{}

	head := NewBall(0, 0, 0, data.HeadSize, data.Color("Head", 60))
	chest := NewBall(-26, 46, 0, data.ChestSize, data.Color("Body", 70))
	butt := NewBall(-70, 79, 0, data.ButtSize, data.Color("Body", 40))
	tailtip := NewBall(-156, 162, 0, data.TailTipSize, data.Color("Body", 50))
	tailtip.SetDistance(data.TailLength, *butt)
	tailtip.RotateAround(*butt, data.TailAngle, 2)

	tailtip.RotateAround(*chest, data.BodyAngle, 2)
	butt.RotateAround(*chest, data.BodyAngle, 2)

	beakOnset := NewBall(24, 0, 0, data.BeakOnsetSize, data.Color("Beak", 30))
	beakOnset.MoveToSphere(*head)
	beakTip := NewBallP(beakOnset.Center.Shifted(Vector{10, 0, 0}), data.BeakTipSize, data.Color("Beak", 60))
	beakTip.SetDistance(data.BeakLength, *beakOnset)

	for _, zs := range [2]float64{-1, 1} {
		wingAttachFront := NewBall(-6, 34, zs*23, 9, data.Color("Body", 30))
		wingAttachFront.MoveToSphere(*chest)
		wingAttachBack := NewBall(-41, 32, zs*10, 9, data.Color("Head", 30))
		wingMiddle := NewBallP(wingAttachBack.Center.Shifted(Vector{0, 40, zs * 200}), 6, data.Color("Body", 40))
		wingAttachBack.MoveToSphere(*chest)
		wingMiddle.MoveToSphere(*chest)
		wingTip := NewBallP(butt.Center.Shifted(Vector{0, 0, zs}), 6, data.Color("Head", 50))
		wingTip.MoveToSphere(*butt)
		wingTip.SetDistance(data.WingLength, *wingAttachFront)
		wingTip.RotateAround(*wingAttachBack, data.WingAngle, 2)
		wingTip.RotateAround(*wingAttachBack, -zs*data.WingAngle, 1)
		wingMiddle.RotateAround(*wingAttachFront, data.WingAngle, 2)
		wingMiddle.RotateAround(*wingAttachFront, -zs*data.WingAngle, 1)
		b.Add(NewQuad(wingAttachFront, wingMiddle, wingTip, wingAttachBack))

		eye := NewBall(1.5, -1, zs, data.EyeSize, Color{64, 64, 64})
		eye.MoveToSphere(*head)
		b.Add(eye)
	}
	b.Add(
		NewBone(chest, head),
		NewBone(chest, butt),
		NewBone(butt, tailtip),
		NewBone(beakOnset, beakTip),
	)

	return b
}
*/
