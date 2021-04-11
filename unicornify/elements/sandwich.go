package elements

import (
	. "github.com/balpha/go-unicornify/unicornify/core"
)

type SandwichFunction = func(float64, float64, bool, float64, float64, float64, bool, float64, float64, float64) (bool, TraceIntervals)

type Sandwich struct {
	Balls     [3]*Ball
	Extrusion Vector
	F         SandwichFunction
}

func NewSandwich(b1, b2, b3 *Ball, extrusion Vector, f SandwichFunction) *Sandwich {
	return &Sandwich{Balls: [3]*Ball{b1, b2, b3}, Extrusion: extrusion, F: f}
}

type SandwichTracer struct {
	bottomTracer, topTracer *FlatTracer
	bounds                  Bounds
	F                       SandwichFunction
}

func (s *Sandwich) GetTracer(wv WorldView) Tracer {
	bottom := NewFlatTracer(wv, s.Balls[0], s.Balls[1], s.Balls[2], true, Color{}, s.Extrusion)
	top := NewFlatTracer(wv, s.Balls[0].Shifted(s.Extrusion), s.Balls[1].Shifted(s.Extrusion), s.Balls[2].Shifted(s.Extrusion), true, Color{}, s.Extrusion)

	return &SandwichTracer{
		bottomTracer: bottom,
		topTracer:    top,
		bounds:       bottom.GetBounds().Union(top.GetBounds()),
		F:            s.F,
	}
}

func (t *SandwichTracer) Trace(x, y float64, ray Vector) (bool, float64, Vector, Color) {
	return UnDeepifyTrace(t, x, y, ray)
}

func (t *SandwichTracer) Pruned(rp RenderingParameters) Tracer {
	return SimplyPruned(t, rp) //FIXME
}

func (t *SandwichTracer) TraceDeep(x, y float64, ray Vector) (bool, TraceIntervals) {
	bOk, bV, bW, bZ := t.bottomTracer.TraceToIntersection(x, y, ray)
	tOk, tV, tW, tZ := t.topTracer.TraceToIntersection(x, y, ray)

	if !bOk && !tOk {
		return false, TraceIntervals{}
	}
	return t.F(x, y, bOk, bV, bW, bZ, tOk, tV, tW, tZ)
}

func (t *SandwichTracer) GetBounds() Bounds {
	return t.bounds
}
