package rendering

import (
	. "bitbucket.org/balpha/go-unicornify/unicornify/core"
)

// ------- ShadowCastingTracer -------

type ShadowCastingTracer struct {
	WorldView, LightView      WorldView
	SourceTracer, LightTracer Tracer
	LightProjection           SphereProjection
	Lighten, Darken           float64
}

func (t *ShadowCastingTracer) Trace(x, y float64) (bool, float64, Vector, Color) {
	ok, z, dir, col := t.SourceTracer.Trace(x, y)
	if !ok {
		return ok, z, dir, col
	}
	origPoint := t.WorldView.UnProject(Vector{x, y, z})
	lp := t.LightView.ProjectSphere(origPoint, 0)

	lok, lz, ldir, _ := t.LightTracer.Trace(lp.X(), lp.Y())

	seeing := !lok || lz >= origPoint.Minus(t.LightView.CameraPosition).Length()-0.01

	if !seeing {
		col = Darken(col, uint8(t.Darken))
	} else {
		rayUnit := Vector{lp.X(), lp.Y(), t.LightView.FocalLength}.Unit()
		sp := ldir.Unit().ScalarProd(rayUnit)

		if sp > 0 { // Given a completely realistic world with no rounding errors, this wouldn't happen.
			col = Darken(col, uint8((1-sp)*t.Darken))
		} else if sp < 0 {
			sp = -sp
			if sp < 0.5 {
				col = Darken(col, uint8((0.5-sp)*t.Darken*2))
			} else {
				col = Lighten(col, uint8((sp-0.5)*t.Lighten*2))
			}

		}
	}
	return ok, z, dir, col
}

func (t *ShadowCastingTracer) TraceDeep(x, y float64) (bool, TraceIntervals) {
	return DeepifyTrace(t, x, y)
}

func (t *ShadowCastingTracer) GetBounds() Bounds {
	return t.SourceTracer.GetBounds()
}

func NewShadowCastingTracer(source Tracer, worldView WorldView, shadowCaster Thing, lightPos, lightTarget Vector, lighten, darken float64) *ShadowCastingTracer {
	lightView := WorldView{
		CameraPosition: lightPos,
		LookAtPoint:    lightTarget,
		FocalLength:    1, // doesn't matter
	}
	lightView.Init()
	lightTracer := shadowCaster.GetTracer(lightView)
	lightProjection := worldView.ProjectSphere(lightPos, 0)

	result := &ShadowCastingTracer{
		SourceTracer:    source,
		LightTracer:     lightTracer,
		WorldView:       worldView,
		LightView:       lightView,
		LightProjection: lightProjection,
		Lighten:         lighten,
		Darken:          darken,
	}
	return result
}
