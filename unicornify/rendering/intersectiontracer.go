package rendering

import (
	. "bitbucket.org/balpha/go-unicornify/unicornify/core"
)

type IntersectionTracer struct {
	Base, Other Tracer
	bounds      Bounds
}

func (t *IntersectionTracer) TraceDeep(x, y float64, ray Vector) (bool, TraceIntervals) {
	ok1, i1 := t.Base.TraceDeep(x, y, ray)
	ok2, i2 := t.Other.TraceDeep(x, y, ray)
	if !ok1 || !ok2 {
		return false, EmptyIntervals
	}
	res := i1.Intersect(i2)
	return len(res) > 0, res
}

func (t *IntersectionTracer) GetBounds() Bounds {
	return t.bounds
}

func (t *IntersectionTracer) Pruned(rp RenderingParameters) Tracer {
	return SimplyPruned(t, rp) //FIXME
}

func (t *IntersectionTracer) Trace(x, y float64, ray Vector) (bool, float64, Vector, Color) {
	return UnDeepifyTrace(t, x, y, ray)
}

func NewIntersectionTracer(base, other Tracer) *IntersectionTracer {
	return &IntersectionTracer{base, other, base.GetBounds().Intersect(other.GetBounds())}
}
