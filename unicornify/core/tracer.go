package core

import (
	"image"
	"math"
)

var NoDirection = Vector{0, 0, 0}

type Tracer interface {
	Trace(x, y float64) (bool, float64, Vector, Color)
	TraceDeep(x, y float64) (bool, TraceIntervals)
	GetBounds() Bounds
}

func DeepifyTrace(t Tracer, x, y float64) (bool, TraceIntervals) {
	ok, z, dir, col := t.Trace(x, y)
	inter := TraceIntervals{TraceInterval{
		Start: TraceResult{z, dir, col},
		End:   TraceResult{math.Inf(1), dir, col},
	}}
	return ok, inter
}

func UnDeepifyTrace(t Tracer, x, y float64) (bool, float64, Vector, Color) {
	ok, r := t.TraceDeep(x, y)
	if ok {
		first := r[0].Start
		return true, first.Z, first.Direction, first.Color
	}
	return false, 0, NoDirection, Color{}
}

func DrawTracerPartial(t Tracer, img *image.RGBA, yCallback func(int), bounds image.Rectangle, c chan bool) {
	r := bounds.Intersect(t.GetBounds().ToRect())
	for y := r.Min.Y; y <= r.Max.Y; y++ {
		for x := r.Min.X; x <= r.Max.X; x++ {
			any, _, _, col := t.Trace(float64(x), float64(y))
			if any {
				img.SetRGBA(x, y, col.ToRGBA())
			}
		}
		if yCallback != nil {
			yCallback(y)
		}
	}
	if c != nil {
		c <- true
	}
}
func DrawTracer(t Tracer, img *image.RGBA, yCallback func(int)) {
	DrawTracerPartial(t, img, yCallback, img.Bounds(), nil)
}
func DrawTracerParallel(t Tracer, img *image.RGBA, yCallback func(int), partsRoot int) {
	full := img.Bounds()
	c := make(chan bool)
	parts := partsRoot * partsRoot
	partsLeft := parts
	for x := 0; x < partsRoot; x++ {
		for y := 0; y < partsRoot; y++ {
			r := image.Rect(full.Dx()*x/partsRoot, full.Dy()*y/partsRoot, full.Dx()*(x+1)/partsRoot-1, full.Dy()*(y+1)/partsRoot-1)
			go DrawTracerPartial(t, img, nil, r, c)
		}
	}
	for partsLeft > 0 {
		<-c
		partsLeft--
		if yCallback != nil {
			yCallback(full.Dy() * (parts - partsLeft) / parts)
		}
	}
}
