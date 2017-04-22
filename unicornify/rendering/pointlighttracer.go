package rendering

import (
	. "bitbucket.org/balpha/go-unicornify/unicornify/core"
	"math"
)

// (experimental, unused)

type PointLightTracer struct {
	LightPositions []Vector
	SourceTracer   Tracer
	HalfLifes      []float64
}

func (t *PointLightTracer) Trace(x, y float64, ray Vector) (bool, float64, Vector, Color) {
	ok, z, dir, col := t.SourceTracer.Trace(x, y, ray)
	if !ok {
		return ok, z, dir, col
	}
	dirlen := dir.Length()
	unit := Vector{0, 0, 0}
	if dirlen > 0 {
		unit = dir.Times(1 / dirlen)
	} else {
		return ok, z, dir, col
	}

	lightsum := 0.0
	for i, lightposition := range t.LightPositions {
		lightdir := Vector{float64(x), float64(y), z}.Plus(lightposition.Neg())
		lightdirunit := lightdir.Times(1 / lightdir.Length())

		sp := -unit.ScalarProd(lightdirunit)
		if dirlen == 0 {
			sp = 0.5
		}
		if sp < 0 {
			continue
		}
		strength := math.Pow(0.5, lightdir.Length()/t.HalfLifes[i])
		sp = sp * strength
		lightsum += sp
	}

	if lightsum < 0 {
		col = Black
	} else {
		if lightsum <= 0.5 {
			col = Darken(col, uint8((0.5-lightsum)*2*255))
		} else {
			col = Lighten(col, uint8((lightsum-0.5)*96))
		}
	}

	return ok, z, dir, col
}

func (t *PointLightTracer) TraceDeep(x, y float64, ray Vector) (bool, TraceIntervals) {
	return DeepifyTrace(t, x, y, ray)
}

func (t *PointLightTracer) GetBounds() Bounds {
	return t.SourceTracer.GetBounds()
}

func (t *PointLightTracer) Pruned(rp RenderingParameters) Tracer {
	return SimplyPruned(t, rp) //FIXME
}

func NewPointLightTracer(source Tracer, lightPos ...Vector) *PointLightTracer {
	result := &PointLightTracer{SourceTracer: source, LightPositions: lightPos}
	return result
}
