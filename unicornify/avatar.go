package unicornify

import (
	"bitbucket.org/balpha/gopyrand"
	"image"
	"math"
)

func MakeAvatar(hash string, size int, withBackground bool, zoomOut bool, shading bool, grass bool, persp bool) (error, *image.RGBA) {

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

	focalLength := 250 + rand.Random()*250 // only used for perspective camera

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

	var wv WorldView

	if persp {
		lookAtPoint := uni.Shoulder.Center.Shifted(uni.Head.Center.Shifted(uni.Shoulder.Center.Neg()).Times(factor))
		cp := lookAtPoint.Shifted(Point3d{0, 0, -focalLength}).RotatedAround(uni.Head.Center, -xAngle, 0).RotatedAround(uni.Head.Center, -yAngle, 1)
		wv = PerspectiveWorldView{
			CameraPosition: cp,
			LookAtPoint:    lookAtPoint,
			Shift:          Point2d{0.5 * fsize, factor*fsize/3 + (1-factor)*fsize/2},
			Scale:          ((unicornScaleFactor-0.5)/2.5*2 + 0.5) * fsize / 200.0,
			FocalLength:    focalLength,
		}
	} else {
		wv = ParallelWorldView{
			AngleX:         xAngle,
			AngleY:         yAngle,
			Shift:          Point2d{100, 100},
			RotationCenter: Point3d{150, 0, 0},
			Scale:          unicornScaleFactor * fsize / 400.0,
		}
	}

	uni.Project(wv)
	uni.Sort(wv)

	if !persp {

		headPos := uni.Head.Projection
		shoulderPos := uni.Shoulder.Projection
		// ignoring Z for the following two
		headShift := Point3d{
			fsize/2 - headPos.X(),
			fsize/3 - headPos.Y(),
			0,
		}
		shoulderShift := Point3d{
			fsize/2 - shoulderPos.X(),
			fsize/2 - shoulderPos.Y(),
			0,
		}

		parwv := wv.(ParallelWorldView)

		// shift can be changed without reprojecting
		parwv.Shift = shoulderShift.Shifted(headShift.Shifted(shoulderShift.Neg()).Times(factor)).DiscardZ()
		wv = parwv
	}
	img := image.NewRGBA(image.Rect(0, 0, size, size))
	if withBackground {
		bgdata.Draw(img, shading)
	}

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
	_, shiftedY := wv.Shifted(0, ymaxproj)
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
			return math.Abs(ymaxproj-h.Projection.Y()) <= h.ProjectionRadius
		}
	}

	unidrawer := uni.NewDrawer(img, wv, shading)

	if grass {
		unidrawer.OnAfterDrawThing = func(t Thing, d *Drawer) {
			i, ok := shins[t]
			if ok {
				shin := t.(*Bone)
				hoof := uni.Legs[i].Hoof
				if !isGroundHoof(hoof, shin) {
					return
				}
				_, grassdata.MinBottomY = wv.Shifted(0, hoof.Projection.Y()+hoof.ProjectionRadius)
				grassdata.ConstrainBone = shin
				DrawGrass(img, grassdata, wv)
			}
		}
		DrawGrass(img, grassdata, wv)
	}

	unidrawer.Draw()

	return nil, img
}
