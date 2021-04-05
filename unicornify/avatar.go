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

	_ = rand.Random()

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

	scaleAndShift := func(t Tracer) Tracer {
		t = NewScalingTracer(wv, t, Scale)
		return NewTranslatingTracer(wv, t, Shift[0], Shift[1])
	}

	uniAndMaybeGrass := &Figure{}
	uniAndMaybeGrass.Add(uni)

	if grass {
		ymaxhoof := -99999.0
		for _, l := range uni.Legs {
			if l.Hoof.Center.Y() > ymaxhoof {
				ymaxhoof = l.Hoof.Center.Y()
			}
		}
		floory := ymaxhoof + uni.Legs[0].Hoof.Radius

		grassSandwich := GrassSandwich(floory, bgdata, grassdata, Shift, Scale, size, wv)
		uniAndMaybeGrass.Add(grassSandwich)
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
