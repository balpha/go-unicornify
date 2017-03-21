package unicornify

import (
	"image"
	"image/color"
)

type Tracer interface {
	Trace(x, y int) (bool, float64, Color)
	GetBounds() image.Rectangle
}

func DrawTracer(t Tracer, img *image.RGBA, yCallback func(int)) {
	r := img.Bounds().Intersect(t.GetBounds())
	for y := r.Min.Y; y <= r.Max.Y; y++ {
		for x := r.Min.X; x <= r.Max.X; x++ {
			any, _, col := t.Trace(x, y)
			if any {
				img.Set(x, y, col)
			}
		}
		if yCallback != nil {
			yCallback(y)
		}
	}

}

// ------- GroupTracer -------

type GroupTracer struct {
	tracers       []Tracer
	bounds        image.Rectangle
	boundsCurrent bool
}

func NewGroupTracer() *GroupTracer {
	return &GroupTracer{}
}

func (gt *GroupTracer) Trace(x, y int) (bool, float64, Color) {
	any := false
	var minz float64 = 0.0
	var col Color = Black
	for _, t := range gt.tracers {
		tr := t.GetBounds()
		if x < tr.Min.X || x > tr.Max.X || y < tr.Min.Y || y > tr.Max.Y {
			continue
		}
		ok, z, thiscol := t.Trace(x, y)
		if ok {
			if !any || z < minz {
				col = thiscol
				minz = z
				any = true
			}
		}
	}
	return any, minz, col
}

func (gt *GroupTracer) GetBounds() image.Rectangle {
	if !gt.boundsCurrent {
		if len(gt.tracers) == 0 {
			gt.bounds = image.Rect(-10, -10, -10, -10)
		} else {
			r := gt.tracers[0].GetBounds()
			for _, t := range gt.tracers[1:] {
				r = r.Union(t.GetBounds())
			}
			gt.bounds = r
		}
		gt.boundsCurrent = true
	}
	return gt.bounds
}

func (gt *GroupTracer) Add(ts ...Tracer) {
	for _, t := range ts {
		gt.tracers = append(gt.tracers, t)
	}
	gt.boundsCurrent = false
}

// ------- ImageTracer -------

type ImageTracer struct {
	img    *image.RGBA
	bounds image.Rectangle
	z      float64
}

func (t *ImageTracer) Trace(x, y int) (bool, float64, Color) {
	tr := t.bounds
	if x < tr.Min.X || x > tr.Max.X || y < tr.Min.Y || y > tr.Max.Y {
		return false, 0, Black
	}
	c := t.img.At(x, y).(color.RGBA)
	if c.A < 255 {
		return false, 0, Black
	}
	return true, t.z, Color{c.R, c.G, c.B}
}

func (t *ImageTracer) GetBounds() image.Rectangle {
	return t.bounds
}
