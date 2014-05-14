package unicornify

import (
	"bitbucket.org/balpha/gopyrand"
	"image"
	"math"
)

func MakeAvatar(hash string, size int, withBackground bool, zoomOut bool) (error, *image.RGBA) {

	rand := pyrand.NewRandom()
	err := rand.SeedFromHexString(hash)
	if err != nil {
		return err, nil
	}

	data := UnicornData{}
	bgdata := BackgroundData{}

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
	// end randomization

	if (yAngle-90*DEGREE)*data.NeckTilt > 0 {
		// The unicorn should look at the camera.
		data.NeckTilt *= -1
		data.FaceTilt *= -1
	}
	uni := NewUnicorn(data)

	wv := WorldView{
		AngleX:         xAngle,
		AngleY:         yAngle,
		Shift:          Point2d{100, 100},
		RotationCenter: Point3d{150, 0, 0},
	}

	fsize := float64(size)

	uni.Scale(unicornScaleFactor * fsize / 400.0)

	uni.Project(wv)
	uni.Sort(wv)

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

	// factor = 1 means center the head at (1/2, 1/3); factor = 0 means
	// center the shoulder at (1/2, 1/2)
	factor := math.Sqrt((unicornScaleFactor - .5) / 2.5)

	// shift can be changed without reprojecting
	wv.Shift = shoulderShift.Shifted(headShift.Shifted(shoulderShift.Neg()).Times(factor)).DiscardZ()

	img := image.NewRGBA(image.Rect(0, 0, size, size))
	if withBackground {
		bgdata.Draw(img)
	}
	uni.Draw(img, wv)

	return nil, img
}
