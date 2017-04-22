package rendering

import (
	. "bitbucket.org/balpha/go-unicornify/unicornify/core"
)

type ScalingTracer struct {
	Source Tracer
	Scale  float64
	bounds Bounds
	wv     WorldView
}

func (t *ScalingTracer) TraceDeep(x, y float64, ray Vector) (bool, TraceIntervals) {
	newX := x / t.Scale
	newY := y / t.Scale
	newRay := t.wv.Ray(newX, newY)
	return t.Source.TraceDeep(newX, newY, newRay)
}

func (t *ScalingTracer) GetBounds() Bounds {
	return t.bounds
}

func (t *ScalingTracer) Trace(x, y float64, ray Vector) (bool, float64, Vector, Color) {
	newX := x / t.Scale
	newY := y / t.Scale
	newRay := t.wv.Ray(newX, newY)
	ok, z, dir, col := t.Source.Trace(newX, newY, newRay)
	return ok, z * t.Scale, dir, col
}

func NewScalingTracer(wv WorldView, source Tracer, scale float64) *ScalingTracer {
	b := source.GetBounds()
	if !b.Empty {
		b.XMin *= scale
		b.XMax *= scale
		b.YMin *= scale
		b.YMax *= scale
	}
	return &ScalingTracer{source, scale, b, wv}
}

func (t *ScalingTracer) Pruned(rp RenderingParameters) Tracer {
	rpScaled := rp.Scaled(t.Scale)
	prunedSource := t.Source.Pruned(rpScaled)
	if prunedSource == nil {
		return nil
	} else if prunedSource == t.Source {
		return t
	}
	return NewScalingTracer(t.wv, prunedSource, t.Scale)
}
