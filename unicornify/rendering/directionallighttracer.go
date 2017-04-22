package rendering

import (
	. "bitbucket.org/balpha/go-unicornify/unicornify/core"
)

type DirectionalLightTracer struct {
	SourceTracer       Tracer
	LightDirectionUnit Vector
	Lighten, Darken    float64
}

func (t *DirectionalLightTracer) Trace(x, y float64, ray Vector) (bool, float64, Vector, Color) {
	ok, z, dir, col := t.SourceTracer.Trace(x, y, ray)
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

func (t *DirectionalLightTracer) TraceDeep(x, y float64, ray Vector) (bool, TraceIntervals) {
	return DeepifyTrace(t, x, y, ray)
}

func (t *DirectionalLightTracer) GetBounds() Bounds {
	return t.SourceTracer.GetBounds()
}

func (t *DirectionalLightTracer) Pruned(rp RenderingParameters) Tracer {
	prunedSource := t.SourceTracer.Pruned(rp)
	if prunedSource == nil {
		return nil
	} else if prunedSource == t.SourceTracer {
		return t
	} else {
		return &DirectionalLightTracer{
			prunedSource,
			t.LightDirectionUnit,
			t.Lighten,
			t.Darken,
		}
	}
}

func (t *DirectionalLightTracer) SetLightDirection(dir Vector) {
	length := dir.Length()
	if length != 0 {
		dir = dir.Times(1 / length)
	}
	t.LightDirectionUnit = dir
}

func NewDirectionalLightTracer(source Tracer, lightDirection Vector, lighten, darken float64) *DirectionalLightTracer {
	result := &DirectionalLightTracer{SourceTracer: source, Lighten: lighten, Darken: darken}
	result.SetLightDirection(lightDirection)
	return result
}
