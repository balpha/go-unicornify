package rendering

import (
	. "bitbucket.org/balpha/go-unicornify/unicornify/core"
)

type DirectionalLightTracer struct {
	GroupTracer
	LightDirectionUnit Vector
	Lighten, Darken    float64
}

func (t *DirectionalLightTracer) Trace(x, y float64) (bool, float64, Vector, Color) {
	ok, z, dir, col := t.GroupTracer.Trace(x, y)
	if !ok {
		return ok, z, dir, col
	}
	dirlen := dir.Length()
	if dirlen == 0 {
		return ok, z, dir, col
	}

	unit := dir.Times(1 / dirlen)
	sp := unit.ScalarProd(t.LightDirectionUnit)

	if sp >= 0 {
		col = Darken(col, uint8(sp*t.Darken))
	} else {
		col = Lighten(col, uint8(-sp*t.Lighten))
	}

	return ok, z, dir, col
}

func (t *DirectionalLightTracer) TraceDeep(x, y float64) (bool, TraceIntervals) {
	return DeepifyTrace(t, x, y)
}

func (t *DirectionalLightTracer) Add(ts ...Tracer) {
	t.GroupTracer.Add(ts...)
}
func (t *DirectionalLightTracer) SetLightDirection(dir Vector) {
	length := dir.Length()
	if length != 0 {
		dir = dir.Times(1 / length)
	}
	t.LightDirectionUnit = dir
}

func NewDirectionalLightTracer(lightDirection Vector, lighten, darken float64) *DirectionalLightTracer {
	result := &DirectionalLightTracer{Lighten: lighten, Darken: darken}
	result.SetLightDirection(lightDirection)
	return result
}
