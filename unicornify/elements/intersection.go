package elements

import (
	. "github.com/balpha/go-unicornify/unicornify/core"
	. "github.com/balpha/go-unicornify/unicornify/rendering"
)

type Intersection struct {
	Base  Thing
	Other Thing
}

func NewIntersection(base, other Thing) *Intersection {
	return &Intersection{base, other}
}

func (i *Intersection) GetTracer(wv WorldView) Tracer {
	return NewIntersectionTracer(i.Base.GetTracer(wv), i.Other.GetTracer(wv))
}
