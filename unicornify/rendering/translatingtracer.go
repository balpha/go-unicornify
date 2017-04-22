package rendering

import (
	. "bitbucket.org/balpha/go-unicornify/unicornify/core"
)

type TranslatingTracer struct {
	Source         Tracer
	ShiftX, ShiftY float64
	bounds         Bounds
	wv             WorldView
}

func (t *TranslatingTracer) TraceDeep(x, y float64, ray Vector) (bool, TraceIntervals) {
	newX := x - t.ShiftX
	newY := y - t.ShiftY
	newRay := t.wv.Ray(newX, newY)

	return t.Source.TraceDeep(newX, newY, newRay)
}

func (t *TranslatingTracer) GetBounds() Bounds {
	return t.bounds
}

func (t *TranslatingTracer) Trace(x, y float64, ray Vector) (bool, float64, Vector, Color) {
	newX := x - t.ShiftX
	newY := y - t.ShiftY
	newRay := t.wv.Ray(newX, newY)
	return t.Source.Trace(newX, newY, newRay)
}

func NewTranslatingTracer(wv WorldView, source Tracer, dx, dy float64) *TranslatingTracer {
	b := source.GetBounds()
	if !b.Empty {
		b.XMin += dx
		b.XMax += dx
		b.YMin += dy
		b.YMax += dy
	}
	return &TranslatingTracer{source, dx, dy, b, wv}
}

func (t *TranslatingTracer) Pruned(rp RenderingParameters) Tracer {
	rpShifted := rp.Translated(t.ShiftX, t.ShiftY)
	prunedSource := t.Source.Pruned(rpShifted)
	if prunedSource == nil {
		return nil
	} else if prunedSource == t.Source {
		return t
	}
	return NewTranslatingTracer(t.wv, prunedSource, t.ShiftX, t.ShiftY)
}
