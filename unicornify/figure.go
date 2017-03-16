package unicornify

import (
	"image"
	"math"
)

type DrawingCallback func (t Thing, d *Drawer)

type Thing interface {
	Draw(img *image.RGBA, wv WorldView, shading bool)
	Project(wv WorldView)
	Bounding() image.Rectangle
	Sort(wv WorldView)
}

type Figure struct {
	things []Thing
}

type Drawer struct {
	Figure *Figure
	Image *image.RGBA
	Wv WorldView
	Shading bool
	OnBeforeDrawThing DrawingCallback
	OnAfterDrawThing DrawingCallback
}

func (f *Figure) Add(things ...Thing) {
	f.things = append(f.things, things...)
}

func (f *Figure) Project(wv WorldView) {
	for i := 0; i < len(f.things); i++ {
		f.things[i].Project(wv)
	}
}
func (f *Figure) NewDrawer(img *image.RGBA, wv WorldView, shading bool) *Drawer {
	return &Drawer{
		f,
		img,
		wv,
		shading,
		nil,
		nil,
	}
}
func (f *Figure) Draw(img *image.RGBA, wv WorldView, shading bool) {
	f.NewDrawer(img, wv, shading).Draw()
}

func (d *Drawer) Draw() {
	before := d.OnBeforeDrawThing
	after := d.OnAfterDrawThing
	for _, t := range d.Figure.things {
		if (before != nil) {
			before(t, d)
		}
		t.Draw(d.Image, d.Wv, d.Shading)
		if (after != nil) {
			after(t, d)
		}
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
					inter := firstT.Bounding().Intersect(secondT.Bounding())
					c := Compare(wv, firstT, secondT, float64(inter.Min.X + inter.Dx()/2), float64(inter.Min.Y + inter.Dy()/2))
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
	first := true
	for first || len(drawAfter) > 0 {
		first = false
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

func Zat(t Thing, x, y float64) float64 {
	r := t.Bounding()
	var c float64
	isX := r.Dx() > r.Dy()
	if isX {
		c = x
	} else {
		c = y
	}
	switch t := t.(type) {
	case *Ball:
		return t.Projection.Z()
	case *Bone:
		p1 := t.Balls[0].Projection
		p2 := t.Balls[1].Projection
		var c1, c2 float64
		if isX {
			c1 = p1.X()
			c2 = p2.X()
		} else {
			c1 = p1.Y()
			c2 = p2.Y()
		}
		if c < c1 && c < c2 {
			if c1 < c2 {
				return p1.Z()
			} else  {
				return p2.Z()
			}
		}
		if c > c1 && c > c2 {
			if c1 > c2 {
				return p1.Z()
			} else  {
				return p2.Z()
			}
		}
		if c1 == c2 {
			return p1.Z() + (p2.Z() - p1.Z()) / 2
		}
		return p1.Z() + (p2.Z() - p1.Z()) * (c - c1) / (c2 - c1)
	case *Figure:
		res := Zat(t.things[0], x, y)
		for _, s := range t.things[1:] {
			res = math.Max(res, Zat(s, x, y))
		}
		return res
	default:
		panic("Zat doesn't handle this")
	}
}

func Compare(wv WorldView, first, second Thing, x, y float64) int {
	
	// special case: if we have two bones that share a ball, compare
	// the two non-shared balls instead

	b1, b1_ok := first.(*Bone)
	b2, b2_ok := second.(*Bone)
	
	if b1_ok && b2_ok {
		for i:=0; i<=3; i++ {
			if b1.Balls[i&1] == b2.Balls[(i&2)>>1] {
				return Compare(wv, b1.Balls[1-(i&1)], b2.Balls[1-((i&2)>>1)], x, 1)
			}
		}
	}
	
	z1 := Zat(first, x, y)
	z2 := Zat(second, x, y)
	switch {
	case z1 < z2:
		return -1
	case z1 > z2:
		return 1
	default:
		return 0
	}
}
