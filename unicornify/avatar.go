package unicornify

import (
	"bitbucket.org/balpha/gopyrand"
	"image"
	"math"
)

var SCALE float64 // FIXME temp for non-linear bones, figure out how to handle it correctly

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

	lightDirection := Point3d{rand.Random()*16 - 8, 10, rand.Random() * 3} // note this is based on projected, not original, coordinate system

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

	fsize := float64(size)

	// factor = 1 means center the head at (1/2, 1/3); factor = 0 means
	// center the shoulder at (1/2, 1/2)
	factor := math.Sqrt((unicornScaleFactor - .5) / 2.5)

	lookAtPoint := uni.Shoulder.Center.Shifted(uni.Head.Center.Shifted(uni.Shoulder.Center.Neg()).Times(factor))
	cp := lookAtPoint.Shifted(Point3d{0, 0, -3 * focalLength}).RotatedAround(uni.Head.Center, -xAngle, 0).RotatedAround(uni.Head.Center, -yAngle, 1)
	wv := PerspectiveWorldView{
		CameraPosition: cp,
		LookAtPoint:    lookAtPoint,
		FocalLength:    focalLength,
	}
	Shift := Point2d{0.5 * fsize, factor*fsize/3 + (1-factor)*fsize/2}
	Scale := ((unicornScaleFactor-0.5)/2.5*2 + 0.5) * fsize / 140.0

	uni.Project(wv)

	img := image.NewRGBA(image.Rect(0, 0, size, size))
	if withBackground {
		bgdata.Draw(img, shading)
	}

	grassTracers := make([]Tracer, 0)

	scaleAndShift := func(t Tracer) Tracer {
		t = NewScalingTracer(t, Scale)
		return NewTranslatingTracer(t, Shift[0], Shift[1])
	}

	if grass {
		shins := make(map[Thing]int)
		ymax := -99999.0
		ymax2 := -99999.0
		ymaxproj := -99999.0
		for i, l := range uni.Legs {
			shins[l.Shin] = i
			if l.Hoof.Center.Y() > ymax {
				ymax2 = ymax
				ymax = l.Hoof.Center.Y()
			} else if l.Hoof.Center.Y() > ymax2 {
				ymax2 = l.Hoof.Center.Y()
			}
			ymaxproj = math.Max(ymaxproj, l.Hoof.Projection.Y())
		}
		shiftedY := ymaxproj*Scale + Shift[1]
		hoofHorizonDist := (shiftedY/fsize - bgdata.Horizon) / (1 - bgdata.Horizon) // 0 = bottom foot at horizon
		if hoofHorizonDist < 0.5 {
			gf := 1 + (1-hoofHorizonDist/0.5)*2
			grassdata.BladeHeightFar *= gf
			grassdata.BladeHeightNear *= gf
		}

		isGroundHoof := func(h *Ball, s *Bone) bool {
			r := s.Bounding()
			if r.Dx()*2 > r.Dy()*3 {
				return false
			}
			if xAngle >= -3*DEGREE {
				yground := math.Min(ymax-h.Radius, ymax2)
				return yground-h.Center.Y() <= 0
			} else {
				return math.Abs(ymaxproj-h.Projection.Y()) <= h.Projection.Radius
			}
		}

		groundHoofs := make([]Leg, 0)
		for _, l := range uni.Legs {
			hoof := l.Hoof
			if !isGroundHoof(hoof, l.Shin) {
				continue
			}
			var i int
			for i = 0; i < len(groundHoofs) && groundHoofs[i].Hoof.Projection.Z() < hoof.Projection.Z(); i++ {
			}
			head := append(groundHoofs[:i], l)
			groundHoofs = append(head, groundHoofs[i:]...)
		}
		for _, l := range groundHoofs {
			h := l.Hoof
			gimg := image.NewRGBA(image.Rect(0, 0, size, size))
			shiftedY := h.Projection.Y()*Scale + Shift[1]
			grassdata.MinBottomY = shiftedY + h.Projection.Radius*Scale
			DrawGrass(gimg, grassdata, wv)
			shinTracer := scaleAndShift(l.Shin.GetTracer(wv))
			z := func(x, y float64) (bool, float64) {
				ok, z, _, _ := shinTracer.Trace(x, y)
				return ok, z - 1
			}
			tr := &ImageTracer{gimg, shinTracer.GetBounds(), z}
			grassTracers = append(grassTracers, tr)
		}
		grassdata.MinBottomY = 0
		DrawGrass(img, grassdata, wv)
	}

	SCALE = Scale

	tracer := uni.GetTracer(wv)

	if shading {
		lt := NewDirectionalLightTracer(lightDirection)
		lt.Add(tracer)
		tracer = lt
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
		DrawTracerParallel(tracer, img, yCallback, parts)
	} else {
		DrawTracer(tracer, img, yCallback)
	}

	return nil, img
}
