package rendering

import (
	. "bitbucket.org/balpha/go-unicornify/unicornify/core"
)

type TranslatingTracer struct {
	Source         Tracer
	ShiftX, ShiftY float64
	bounds         Bounds
}

func (t *TranslatingTracer) TraceDeep(x, y float64) (bool, TraceIntervals) {
	return t.Source.TraceDeep(x-t.ShiftX, y-t.ShiftY)
}

func (t *TranslatingTracer) GetBounds() Bounds {
	return t.bounds
}

func (t *TranslatingTracer) Trace(x, y float64) (bool, float64, Vector, Color) {
	return t.Source.Trace(x-t.ShiftX, y-t.ShiftY)
}

func NewTranslatingTracer(source Tracer, dx, dy float64) *TranslatingTracer {
	b := source.GetBounds()
	if !b.Empty {
		b.XMin += dx
		b.XMax += dx
		b.YMin += dy
		b.YMax += dy
	}
	return &TranslatingTracer{source, dx, dy, b}
}
