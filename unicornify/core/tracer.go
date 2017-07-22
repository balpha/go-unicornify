package core

import (
	"image"
	"math"
)

var NoDirection = Vector{0, 0, 0}

type Tracer interface {
	Trace(x, y float64, ray Vector) (bool, float64, Vector, Color)
	TraceDeep(x, y float64, ray Vector) (bool, TraceIntervals)
	GetBounds() Bounds
	Pruned(rp RenderingParameters) Tracer // okay to return nil or self
}

func DeepifyTrace(t Tracer, x, y float64, ray Vector) (bool, TraceIntervals) {
	ok, z, dir, col := t.Trace(x, y, ray)
	inter := TraceIntervals{TraceInterval{
		Start: TraceResult{z, dir, col},
		End:   TraceResult{math.Inf(1), dir, col},
	}}
	return ok, inter
}

func UnDeepifyTrace(t Tracer, x, y float64, ray Vector) (bool, float64, Vector, Color) {
	ok, r := t.TraceDeep(x, y, ray)
	if ok {
		first := r[0].Start
		return true, first.Z, first.Direction, first.Color
	}
	return false, 0, NoDirection, Color{}
}
func SimplyPruned(t Tracer, rp RenderingParameters) Tracer {
	if rp.Contains(t.GetBounds()) {
		return t
	}
	return nil
}

func DrawTracerPartial(t Tracer, wv WorldView, img *image.RGBA, yCallback func(int), bounds image.Rectangle, c chan bool) {
	r := bounds.Intersect(t.GetBounds().ToRect())
	rp := RenderingParameters{
		1,
		float64(r.Min.X - 1), float64(r.Max.X + 1),
		float64(r.Min.Y - 1), float64(r.Max.Y + 1),
	}
	pruned := t.Pruned(rp)
	if pruned != nil {
		for y := r.Min.Y; y <= r.Max.Y; y++ {
			for x := r.Min.X; x <= r.Max.X; x++ {
				fx, fy := float64(x), float64(y)
				any, _, _, col := pruned.Trace(fx, fy, wv.Ray(fx, fy))
				if any {
					img.SetRGBA(x, y, col.ToRGBA())
				}
			}
			if yCallback != nil {
				yCallback(y)
			}
		}
	}
	if c != nil {
		c <- true
	}
}
func DrawTracer(t Tracer, wv WorldView, img *image.RGBA, yCallback func(int)) {
	DrawTracerPartial(t, wv, img, yCallback, img.Bounds(), nil)
}
func DrawTracerParallel(t Tracer, wv WorldView, img *image.RGBA, yCallback func(int), partsRoot int) {
	full := img.Bounds()
	c := make(chan bool)
	parts := partsRoot * partsRoot
	partsLeft := parts
	for x := 0; x < partsRoot; x++ {
		for y := 0; y < partsRoot; y++ {
			r := image.Rect(full.Dx()*x/partsRoot, full.Dy()*y/partsRoot, full.Dx()*(x+1)/partsRoot-1, full.Dy()*(y+1)/partsRoot-1)
			go DrawTracerPartial(t, wv, img, nil, r, c)
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
