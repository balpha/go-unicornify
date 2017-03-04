package unicornify

import (
	"image"
	"math"
)

type Thing interface {
	Draw(img *image.RGBA, wv WorldView, shading bool)
	Project(wv WorldView)
	Bounding() image.Rectangle
	Sort(wv WorldView)
}

type Figure struct {
	things []Thing
}

func (f *Figure) Add(things ...Thing) {

	f.things = append(f.things, things...)
}

func (f *Figure) Project(wv WorldView) {
	for i := 0; i < len(f.things); i++ {
		f.things[i].Project(wv)
	}
}

func (f *Figure) Draw(img *image.RGBA, wv WorldView, shading bool) {
	for i := 0; i < len(f.things); i++ {
		f.things[i].Draw(img, wv, shading)
	}
}

func (f *Figure) Sort(wv WorldView) {
	// values of this map are slices of all things that have
	// to be drawn before the thing with the key index in f.things
	drawAfter := map[int]map[int]bool{}
	for i := 0; i < len(f.things); i++ {
		drawAfter[i] = map[int]bool{}
	}
	for first := 0; first < len(f.things); first++ {
		for second := first + 1; second < len(f.things); second++ {
			if !drawAfter[first][second] && !drawAfter[second][first] {
				firstT, secondT := f.things[first], f.things[second]
				if firstT.Bounding().Overlaps(secondT.Bounding()) {
					c := Compare(wv, firstT, secondT)
					switch {
					case c < 0: // first is in front of second
						drawAfter[first][second] = true
					case c > 0:
						drawAfter[second][first] = true
					}
				}
			}
		}
	}

	// this is pretty much the algorithm from http://stackoverflow.com/questions/952302/
	sortedThings := make([]int, 0, len(f.things))
	queue := make([]int, 0, len(f.things))
	for i, deps := range drawAfter {
		if len(deps) == 0 {
			queue = append(queue, i)
			delete(drawAfter, i) // according to the spec, deleting while iterating over the map shouldn't cause issues
		}
	}
	for len(drawAfter) > 0 {
		for len(queue) > 0 {
			popped := queue[len(queue)-1]
			queue = queue[:len(queue)-1]
			sortedThings = append(sortedThings, popped)
			for i, deps := range drawAfter {
				if deps[popped] {
					delete(deps, popped)
					if len(deps) == 0 {
						queue = append(queue, i)
						delete(drawAfter, i)
					}
				}
			}
		}
		// if the sorting couldn't fullfill all "draw after" contraints,
		// we remove the ball / bone which lies farthest in the back
		// and try again
		if len(drawAfter) > 0 {
			leastEvil := 0
			leastEvilZ := float64(-0xffffff)
			for i, _ := range drawAfter {
				z := maxZ(f.things[i])
				if z > leastEvilZ {
					leastEvilZ = z
					leastEvil = i
				}
			}
			delete(drawAfter, leastEvil)
			sortedThings = append(sortedThings, leastEvil)
			for i, deps := range drawAfter {
				if deps[leastEvil] {
					delete(deps, leastEvil)
					if len(deps) == 0 {
						queue = append(queue, i)
						delete(drawAfter, i)
					}
				}
			}

		}
	}
	newThings := make([]Thing, len(f.things))
	for ni, oi := range sortedThings {
		newThings[ni] = f.things[oi]
	}
	f.things = newThings
	for _, t := range f.things {
		t.Sort(wv)
	}
}

func (f *Figure) Bounding() image.Rectangle {
	result := image.Rect(0, 0, 0, 0)
	for _, t := range f.things {
		result = result.Union(t.Bounding())
	}
	return result
}

func (f *Figure) Scale(factor float64) {
	for b := range f.BallSet() {
		b.Radius *= factor
		b.Center = b.Center.Times(factor)
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
	default:
		panic("unhandled thing type")
	}
	if outer {
		close(ch)
	}
}

func maxZ(t Thing) float64 {
	switch t := t.(type) {
	case *Ball:
		return t.Projection.Z()
	case *Bone:
		return math.Max(t.Balls[0].Projection.Z(), t.Balls[1].Projection.Z())
	case *Figure:
		res := maxZ(t.things[0])
		for _, s := range t.things[1:] {
			res = math.Max(res, maxZ(s))
		}
		return res
	default:
		panic("maxZ doesn't handle this")
	}
}

// much simpler than the version in python, but (at least for the limited
// perspectives in the avatar generator) just as okay results, sometimes
// even better (see c55de3e9a7424bd984e6e56cedba5be8)
func Compare(wv WorldView, first, second Thing) int {
	z1 := maxZ(first)
	z2 := maxZ(second)
	switch {
	case z1 < z2:
		return -1
	case z1 > z2:
		return 1
	default:
		return 0
	}
}
