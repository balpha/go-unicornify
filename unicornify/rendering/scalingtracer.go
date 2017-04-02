package rendering

import (
	. "bitbucket.org/balpha/go-unicornify/unicornify/core"
)

type ScalingTracer struct {
	Source Tracer
	Scale  float64
	bounds Bounds
}

func (t *ScalingTracer) TraceDeep(x, y float64) (bool, TraceIntervals) {
	return t.Source.TraceDeep(x/t.Scale, y/t.Scale) // TODO: scale the result?
}

func (t *ScalingTracer) GetBounds() Bounds {
	return t.bounds
}

func (t *ScalingTracer) Trace(x, y float64) (bool, float64, Vector, Color) {
	ok, z, dir, col := t.Source.Trace(x/t.Scale, y/t.Scale) // TODO: scale the result?
	return ok, z * t.Scale, dir, col
}

func NewScalingTracer(source Tracer, scale float64) *ScalingTracer {
	b := source.GetBounds()
	if !b.Empty {
		b.XMin *= scale
		b.XMax *= scale
		b.YMin *= scale
		b.YMax *= scale
	}
	return &ScalingTracer{source, scale, b}
}
