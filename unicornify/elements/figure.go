package elements

import (
	. "bitbucket.org/balpha/go-unicornify/unicornify/core"
	. "bitbucket.org/balpha/go-unicornify/unicornify/rendering"
)

type Figure struct {
	things []Thing
}

func (f *Figure) Add(things ...Thing) {
	f.things = append(f.things, things...)
}

func (f *Figure) GetTracer(wv WorldView) Tracer {
	gt := NewGroupTracer()

	for _, th := range f.things {
		gt.Add(th.GetTracer(wv))
	}
	return gt
}

func (f *Figure) Scale(factor float64) {
	for b := range f.BallSet() {
		b.Radius *= factor
		b.Center = b.Center.Times(factor)
	}
}

func (f *Figure) Shift(delta Vector) {
	for b := range f.BallSet() {
		b.Shift(delta)
	}
}

func (f *Figure) BallSet() <-chan *Ball {
	seen := make(map[*Ball]bool)
	ch := make(chan *Ball)
	go ballSetImpl(f, seen, ch, true)
	return ch
}

func ballSetImpl(t Thing, seen map[*Ball]bool, ch chan *Ball, outer bool) {
	switch t := t.(type) {
	case *Ball:
		if !seen[t] {
			ch <- t
			seen[t] = true
		}
	case *Bone:
		ballSetImpl(t.Balls[0], seen, ch, false)
		ballSetImpl(t.Balls[1], seen, ch, false)
	case *Figure:
		for _, s := range t.things {
			ballSetImpl(s, seen, ch, false)
		}
	case *Steak:
		for _, s := range t.Balls {
			ballSetImpl(s, seen, ch, false)
		}
	default:
		panic("unhandled thing type")
	}
	if outer {
		close(ch)
	}
}
