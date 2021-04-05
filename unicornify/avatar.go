package unicornify

import (
	. "github.com/balpha/go-unicornify/unicornify/core"
	. "github.com/balpha/go-unicornify/unicornify/elements"
	. "github.com/balpha/go-unicornify/unicornify/rendering"
	"github.com/balpha/gopyrand"
	"image"
	"math"
)

func MakeAvatar(hash string, size int, withBackground bool, zoomOut bool, shading bool, grass bool, parallelize bool, yCallback func(int)) (error, *image.RGBA) {
	rand := pyrand.NewRandom()
	err := rand.SeedFromHexString(hash)
	if err != nil {
		return err, nil
	}

	data := UnicornData{}
	bgdata := BackgroundData{}
	grassdata := GrassData{}

	// begin randomization
	// To keep consistency of unicorns between versions,
	// new Random() calls should always be added at the end
	data.Randomize1(rand)
	bgdata.Randomize1(rand)

	unicornScaleFactor := .5 + math.Pow(rand.Random(), 2)*2.5
	if zoomOut {
		unicornScaleFactor = .5
	}

	sign := rand.Choice(2)*2 - 1
	abs := rand.RandInt(10, 75)
	yAngle := float64(90+sign*abs) * DEGREE
	xAngle := float64(rand.RandInt(-20, 20)) * DEGREE

	data.Randomize2(rand)
	bgdata.Randomize2(rand)
	data.Randomize3(rand)
	grassdata.Randomize(rand)

	grassSlope := 2 + 4*(20-xAngle/DEGREE)/40
	grassScale := 1 + (unicornScaleFactor-0.5)/2.5
	grassdata.BladeHeightNear = (0.02 + 0.02*rand.Random()) * grassScale
	grassdata.BladeHeightFar = grassdata.BladeHeightNear / grassSlope

	focalLength := 250 + rand.Random()*250

	data.Randomize4(rand)

	lightDirection := Vector{rand.Random()*16 - 8, 10, rand.Random() * 3}

	lightDirection = Vector{lightDirection.Z(), lightDirection.Y(), -lightDirection.X()}

	// end randomization

	grassdata.Horizon = bgdata.Horizon
	grassdata.Color1 = bgdata.Color("Land", bgdata.LandLight)
	grassdata.Color2 = bgdata.Color("Land", bgdata.LandLight/2)

	if (yAngle-90*DEGREE)*data.NeckTilt > 0 {
		// The unicorn should look at the camera.
		data.NeckTilt *= -1
		data.FaceTilt *= -1
	}
	uni := NewUnicorn(data)

	if data.PoseKindIndex == 1 /*Walk*/ {
		lowFront := uni.Legs[0].Hoof.Center
		if uni.Legs[1].Hoof.Center.Y() > uni.Legs[0].Hoof.Center.Y() {
			lowFront = uni.Legs[1].Hoof.Center
		}
		lowBack := uni.Legs[2].Hoof.Center
		if uni.Legs[3].Hoof.Center.Y() > uni.Legs[2].Hoof.Center.Y() {
			lowBack = uni.Legs[3].Hoof.Center
		}
		angle := math.Atan((lowBack.Y() - lowFront.Y()) / (lowBack.X() - lowFront.X()))
		for b := range uni.BallSet() {
			b.RotateAround(*uni.Shoulder, -angle, 2)
		}
	}

	if xAngle < 0 {
		for b := range uni.BallSet() {
			b.RotateAround(*uni.Shoulder, yAngle, 1)
			b.RotateAround(*uni.Shoulder, xAngle, 0)
			b.RotateAround(*uni.Shoulder, -yAngle, 1)
		}
		xAngle = 0
	}

	fsize := float64(size)

	// factor = 1 means center the head at (1/2, 1/3); factor = 0 means
	// center the shoulder at (1/2, 1/2)
	factor := math.Sqrt((unicornScaleFactor - .5) / 2.5)

	lookAtPoint := uni.Shoulder.Center.Plus(uni.Head.Center.Plus(uni.Shoulder.Center.Neg()).Times(factor))
	cp := lookAtPoint.Plus(Vector{0, 0, -3 * focalLength}).RotatedAround(uni.Head.Center, -xAngle, 0).RotatedAround(uni.Head.Center, -yAngle, 1)

	wv := WorldView{
		CameraPosition: cp,
		LookAtPoint:    lookAtPoint,
		FocalLength:    focalLength,
	}
	Shift := Point2d{0.5 * fsize, factor*fsize/3 + (1-factor)*fsize/2}
	Scale := ((unicornScaleFactor-0.5)/2.5*2 + 0.5) * fsize / 140.0

	wv.Init()

	img := image.NewRGBA(image.Rect(0, 0, size, size))
	if withBackground {
		bgdata.Draw(img, shading)
	}

	grassTracers := make([]Tracer, 0)

	scaleAndShift := func(t Tracer) Tracer {
		t = NewScalingTracer(wv, t, Scale)
		return NewTranslatingTracer(wv, t, Shift[0], Shift[1])
	}

	uniAndMaybeGrass := &Figure{}
	uniAndMaybeGrass.Add(uni)

	if grass {
		var grassSize float64 = 20000
		var bladeDistance float64 = 3
		var bladeDiameter float64 = 2

		ymaxhoof := -99999.0
		for _, l := range uni.Legs {
			if l.Hoof.Center.Y() > ymaxhoof {
				ymaxhoof = l.Hoof.Center.Y()
			}
		}
		floory := ymaxhoof + uni.Legs[0].Hoof.Radius

		fb1 := NewBall(-grassSize/2, floory, -grassSize/2, 1, Color{255, 0, 0})
		fb2 := NewBall(grassSize/2, floory, -grassSize/2, 1, Color{0, 255, 0})
		fb3 := NewBall(-grassSize/2, floory, grassSize/2, 1, Color{0, 0, 255})

		hx := (0 - Shift[0]) / Scale
		hy := (bgdata.Horizon*float64(size) - Shift[1]) / Scale
		var hdist float64 = 100
		for wv.UnProject(Vector{hx, hy, hdist}).Y() < floory {
			hdist += 100
		}

		swf := func(x, y float64, bOk bool, bV, bW, bZ float64, tOk bool, tV, tW, tZ float64) (bool, float64, Vector, Color) {

			if !bOk || !tOk || tZ > hdist {
				return false, 0, NoDirection, Color{}
			}

			I := Vector{tV * grassSize, 0, tW * grassSize}
			O := Vector{bV * grassSize, 15, bW * grassSize}
			C := O.Minus(I)

			sqrCX := Sqr(C.X())
			sqrCY := Sqr(C.Y())
			sqrCZ := Sqr(C.Z())

			sqrIY := Sqr(I.Y())

			CXCY := C.X() * C.Y()
			CZCY := C.Z() * C.Y()

			CXIY := C.X() * I.Y()
			CZIY := C.Z() * I.Y()

			twoIYCY := 2 * I.Y() * C.Y()

			r0 := bladeDiameter
			sqrr0 := Sqr(r0)

			crossingCells := 5 * float64(RoundUp(math.Abs(C.X())/bladeDistance)+RoundUp(math.Abs(C.Z())/bladeDistance))
			d := C.Times(1.0 / crossingCells)

			prevX := -999999999
			prevY := -999999999

			for n := float64(0); n <= crossingCells; n++ {

				p := I.Plus(d.Times(n))
				cX := float64(RoundDown(p.X()/bladeDistance)) * bladeDistance
				cY := float64(RoundDown(p.Z()/bladeDistance)) * bladeDistance
				cx := int(cX)
				cy := int(cY)
				if cx == prevX && cy == prevY {
					continue
				}
				prevX = cx
				prevY = cy

				randomish1 := float64(QuickRand2(cx, cy)) / 2147483648.0
				randomish2 := float64(QuickRand2(cy, cx)) / 2147483648.0

				B := Vector{cX + bladeDiameter + randomish1*(bladeDistance-2*bladeDiameter), 15, cY + randomish2*bladeDistance}
				T := Vector{cX + randomish2*bladeDistance, 0, cY + randomish1*bladeDistance}
				D := B.Minus(T)

				sqrDX := Sqr(D.X())
				sqrDY := Sqr(D.Y())
				sqrDZ := Sqr(D.Z())

				twoDYDX := 2 * D.Y() * D.X()
				twoDYDZ := 2 * D.Y() * D.Z()

				Q := I.Minus(T)

				m1 := sqrDY*sqrCX - 2*D.Y()*D.X()*CXCY + sqrDX*sqrCY +
					sqrDY*sqrCZ - 2*D.Y()*D.Z()*CZCY + sqrDZ*sqrCY -
					sqrr0*sqrCY

				m2 := 2*Q.X()*C.X()*sqrDY - twoDYDX*(Q.X()*C.Y()+CXIY) + sqrDX*twoIYCY +
					2*Q.Z()*C.Z()*sqrDY - twoDYDZ*(Q.Z()*C.Y()+CZIY) + sqrDZ*twoIYCY -
					sqrr0*twoIYCY

				m3 := sqrDY*Sqr(Q.X()) - twoDYDX*Q.X()*I.Y() + sqrDX*sqrIY +
					sqrDY*Sqr(Q.Z()) - twoDYDZ*Q.Z()*I.Y() + sqrDZ*sqrIY -
					sqrr0*sqrIY

				ep := m2 / m1 //FIXME: zero
				eq := m3 / m1

				disc := Sqr(ep)/4 - eq

				if disc >= 0 {
					t := -ep/2 - math.Sqrt(disc)
					k := (I.Y() + t*C.Y()) / D.Y()

					if k >= 0 && k <= 1 {
						z := tZ + t*(bZ-tZ)

						if z > hdist {
							return false, 0, NoDirection, Color{}
						}

						dir := I.Plus(C.Times(t)).Minus(T.Plus(D.Times(k)))
						dir[1] = -0.1

						return true, z, dir, MixColors(grassdata.Color1, grassdata.Color2, k)

					}
				}

			}
			if bZ <= hdist {
				landColor := MixColors(bgdata.Color("Land", bgdata.LandLight), bgdata.Color("Land", bgdata.LandLight/2), (x*Scale+Shift[0])/float64(size))
				return true, bZ, Vector{0, -1, 0}, landColor
			}
			return false, 0, NoDirection, Color{}
		}

		sw := NewSandwich(fb1, fb2, fb3, Vector{0, -15, 0}, swf)
		uniAndMaybeGrass.Add(sw)
	}

	tracer := uniAndMaybeGrass.GetTracer(wv)

	if shading {
		p := Vector{0, 0, 1000}
		pp := wv.ProjectSphere(p, 0).CenterCS
		ldp := wv.ProjectSphere(p.Plus(lightDirection), 0).CenterCS.Minus(pp)
		lt := NewDirectionalLightTracer(tracer, ldp, 32, 80)

		sc := NewShadowCastingTracer(lt, wv, uniAndMaybeGrass, uni.Head.Center.Minus(lightDirection.Times(1000)), uni.Head.Center, 16, 16)
		tracer = sc
	}

	tracer = scaleAndShift(tracer)

	if grass {
		gt := NewGroupTracer()
		gt.Add(tracer)
		gt.Add(grassTracers...)
		tracer = gt
	}

	if parallelize {
		parts := size / 128
		if parts < 8 {
			parts = 8
		}
		DrawTracerParallel(tracer, wv, img, yCallback, parts)
	} else {
		DrawTracer(tracer, wv, img, yCallback)
	}

	return nil, img
}
