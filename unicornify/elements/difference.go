package elements

import (
	. "github.com/balpha/go-unicornify/unicornify/core"
	. "github.com/balpha/go-unicornify/unicornify/rendering"
)

type Difference struct {
	Base       Thing
	Subtrahend Thing
}

func NewDifference(base, subtrahend Thing) *Difference {
	return &Difference{base, subtrahend}
}

func (i *Difference) GetTracer(wv WorldView) Tracer {
	return NewDifferenceTracer(i.Base.GetTracer(wv), i.Subtrahend.GetTracer(wv))
}
