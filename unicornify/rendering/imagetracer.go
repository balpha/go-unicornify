package rendering

import (
	. "bitbucket.org/balpha/go-unicornify/unicornify/core"
	"image"
	"image/color"
)

type ImageTracer struct {
	img    *image.RGBA
	bounds Bounds
	z      func(x, y float64) (bool, float64)
}

func (t *ImageTracer) Trace(x, y float64, ray Vector) (bool, float64, Vector, Color) {
	if !t.bounds.ContainsXY(x, y) {
		return false, 0, NoDirection, Black
	}
	c := t.img.At(Round(x), Round(y)).(color.RGBA)
	if c.A < 255 {
		return false, 0, NoDirection, Black
	}

	ok, z := t.z(x, y)

	return ok, z, NoDirection, Color{c.R, c.G, c.B}
}

func (t *ImageTracer) TraceDeep(x, y float64, ray Vector) (bool, TraceIntervals) {
	return DeepifyTrace(t, x, y, ray)
}

func (t *ImageTracer) Pruned(rp RenderingParameters) Tracer {
	return SimplyPruned(t, rp)
}

func (t *ImageTracer) GetBounds() Bounds {
	return t.bounds
}

func NewImageTracer(img *image.RGBA, bounds Bounds, z func(x, y float64) (bool, float64)) *ImageTracer {
	return &ImageTracer{
		img,
		bounds,
		z,
	}
}
