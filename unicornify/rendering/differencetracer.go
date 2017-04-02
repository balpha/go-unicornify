package rendering

import (
	. "bitbucket.org/balpha/go-unicornify/unicornify/core"
)

type DifferenceTracer struct {
	Base, Subtrahend Tracer
}

func (t *DifferenceTracer) TraceDeep(x, y float64) (bool, TraceIntervals) {
	ok1, i1 := t.Base.TraceDeep(x, y)
	ok2, i2 := t.Subtrahend.TraceDeep(x, y)
	if !ok1 {
		return false, EmptyIntervals
	}
	if !ok2 {
		return ok1, i1
	}
	res := i1.Intersect(i2.Invert())
	return len(res) > 0, res
}

func (t *DifferenceTracer) GetBounds() Bounds {
	return t.Base.GetBounds()
}

func (t *DifferenceTracer) Trace(x, y float64) (bool, float64, Vector, Color) {
	return UnDeepifyTrace(t, x, y)
}

func NewDifferenceTracer(base, subtrahend Tracer) *DifferenceTracer {
	return &DifferenceTracer{base, subtrahend}
}
