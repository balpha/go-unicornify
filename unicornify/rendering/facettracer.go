package rendering

import (
	. "bitbucket.org/balpha/go-unicornify/unicornify/core"
	"math"
)

type FacetTracer struct {
	countRoot  int
	countRootF float64
	facets     []*GroupTracer
	bounds     Bounds
	isEmpty    bool
}

func NewFacetTracer(bounds Bounds, countRoot int) *FacetTracer {
	return &FacetTracer{
		countRoot:  countRoot,
		countRootF: float64(countRoot),
		facets:     make([]*GroupTracer, countRoot*countRoot),
		bounds:     bounds,
		isEmpty:    true,
	}
}

func (t *FacetTracer) IsEmpty() bool {
	return t.isEmpty
}

func (t *FacetTracer) Trace(x, y float64, ray Vector) (bool, float64, Vector, Color) {
	facet := t.facets[t.facetNum(x, y)]
	if facet == nil {
		return false, 0, NoDirection, Color{}
	}
	return facet.Trace(x, y, ray)
}

func (t *FacetTracer) Add(ts ...Tracer) {
	for _, nt := range ts {
		t.isEmpty = false
		b := nt.GetBounds()
		minx, miny := t.facetCoords(b.XMin, b.YMin)
		maxx, maxy := t.facetCoords(b.XMax, b.YMax)

		for y := miny; y <= maxy; y++ {
			for x := minx; x <= maxx; x++ {
				n := y*t.countRoot + x
				facet := t.facets[n]
				if facet == nil {
					facet = NewGroupTracer()
					t.facets[n] = facet
				}
				facet.Add(nt)
			}
		}
	}
}

func (t *FacetTracer) TraceDeep(x, y float64, ray Vector) (bool, TraceIntervals) {
	facet := t.facets[t.facetNum(x, y)]
	if facet == nil {
		return false, EmptyIntervals
	}
	return facet.TraceDeep(x, y, ray)
}

func (t *FacetTracer) facetNum(x, y float64) int {
	fx, fy := t.facetCoords(x, y)
	return fy*t.countRoot + fx
}
func (t *FacetTracer) facetCoords(x, y float64) (fx, fy int) {
	b := t.bounds
	facetx := math.Min(t.countRootF-1, math.Max(0, t.countRootF*(x-b.XMin)/b.Dx()))
	facety := math.Min(t.countRootF-1, math.Max(0, t.countRootF*(y-b.YMin)/b.Dy()))

	return int(facetx), int(facety)
}

func (t *FacetTracer) GetBounds() Bounds {
	return t.bounds
}

func (t *FacetTracer) Sort() {
	for _, f := range t.facets {
		if f != nil {
			f.Sort()
		}
	}
}

func (t *FacetTracer) Pruned(rp RenderingParameters) Tracer {
	return SimplyPruned(t, rp) // FIXME maybe? facet tracers are the *result* of pruning, so may be fine
}
