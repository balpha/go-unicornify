package unicornify

import (
	. "bitbucket.org/balpha/go-unicornify/unicornify/core"
	. "bitbucket.org/balpha/go-unicornify/unicornify/elements"
	. "bitbucket.org/balpha/go-unicornify/unicornify/rendering"
	"bitbucket.org/balpha/gopyrand"
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

	if grass {
		shins := make(map[Thing]int)
		ymax := -99999.0
		ymaxproj := -99999.0
		for i, l := range uni.Legs {
			shins[l.Shin] = i
			if l.Hoof.Center.Y() > ymax {
				ymax = l.Hoof.Center.Y()
			}
			ymaxproj = math.Max(ymaxproj, ProjectBall(wv, l.Hoof).Y())
		}
		shiftedY := ymaxproj*Scale + Shift[1]
		hoofHorizonDist := (shiftedY/fsize - bgdata.Horizon) / (1 - bgdata.Horizon) // 0 = bottom foot at horizon
		if hoofHorizonDist < 0.5 {
			gf := 1 + (1-hoofHorizonDist/0.5)*2
			grassdata.BladeHeightFar *= gf
			grassdata.BladeHeightNear *= gf
		}

		var groundShadowImg *image.RGBA
		if shading {
			ground := NewSteak(
				NewBall(-400, ymax+uni.Legs[0].Hoof.Radius+1, -400, 1, Color{128, 128, 128}),
				NewBall(700, ymax+uni.Legs[0].Hoof.Radius+1, -400, 1, Color{128, 128, 128}),
				NewBall(-400, ymax+uni.Legs[0].Hoof.Radius+1, 400, 1, Color{128, 128, 128}),
			)
			ground.FourCorners = true
			ground.FourthColor = Color{128, 128, 128}
			ground.Rounded = false
			groundtr := ground.GetTracer(wv)
			groundtr = NewShadowCastingTracer(groundtr, wv, uni, uni.Head.Center.Minus(lightDirection.Times(1000)), uni.Head.Center, 0, 32)
			groundtr = scaleAndShift(groundtr)
			groundShadowImg = image.NewRGBA(image.Rect(0, 0, size, size))

			DrawTracerParallel(groundtr, wv, groundShadowImg, nil, 8)
		}

		for _, l := range uni.Legs {
			h := l.Hoof
			hproj := ProjectBall(wv, h)
			gimg := image.NewRGBA(image.Rect(0, 0, size, size))
			shiftedY := hproj.Y()*Scale + Shift[1]
			grassdata.MinBottomY = shiftedY + hproj.ProjectedRadius*Scale + (ymax-h.Center.Y())*hproj.ProjectedRadius/h.Radius*Scale
			DrawGrass(gimg, grassdata, wv, groundShadowImg)
			shinTracer := scaleAndShift(l.Shin.GetTracer(wv))
			z := func(x, y float64) (bool, float64) {
				ok, z, _, _ := shinTracer.Trace(x, y, wv.Ray(x, y))
				return ok, z - 1
			}
			tr := NewImageTracer(gimg, shinTracer.GetBounds(), z)
			grassTracers = append(grassTracers, tr)
		}
		grassdata.MinBottomY = 0
		DrawGrass(img, grassdata, wv, groundShadowImg)
	}

	tracer := uni.GetTracer(wv)

	if shading {
		p := Vector{0, 0, 1000}
		pp := wv.ProjectSphere(p, 0).CenterCS
		ldp := wv.ProjectSphere(p.Plus(lightDirection), 0).CenterCS.Minus(pp)
		lt := NewDirectionalLightTracer(tracer, ldp, 32, 80)

		sc := NewShadowCastingTracer(lt, wv, uni, uni.Head.Center.Minus(lightDirection.Times(1000)), uni.Head.Center, 16, 16)
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
