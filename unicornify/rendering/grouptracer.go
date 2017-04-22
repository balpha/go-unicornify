package rendering

import (
	. "bitbucket.org/balpha/go-unicornify/unicornify/core"
	"math"
	"sort"
)

type GroupTracer struct {
	tracers       []Tracer
	bounds        Bounds
	boundsCurrent bool
	isSorted      bool
}

func NewGroupTracer() *GroupTracer {
	return &GroupTracer{}
}

func (gt *GroupTracer) Trace(x, y float64, ray Vector) (bool, float64, Vector, Color) {
	any := false
	var minz float64 = 0.0
	var col Color = Black
	var dir Vector
	for _, t := range gt.tracers {
		b := t.GetBounds()
		if !b.ContainsXY(x, y) {
			continue
		}
		if b.ZMax <= 0 {
			continue
		}
		if any && !b.ContainsPointsInFrontOfZ(minz) {
			if gt.isSorted {
				break
			}
			continue
		}
		ok, z, thisdir, thiscol := t.Trace(x, y, ray)
		if ok && z > 0 {
			if !any || z < minz {
				col = thiscol
				minz = z
				dir = thisdir
				any = true
			}
		}
	}
	return any, minz, dir, col
}

func (t *GroupTracer) TraceDeep(x, y float64, ray Vector) (bool, TraceIntervals) {
	result := TraceIntervals{}
	any := false
	for _, t := range t.tracers {
		b := t.GetBounds()
		if !b.ContainsXY(x, y) {
			continue
		}
		if b.ZMax <= 0 {
			continue
		}
		ok, is := t.TraceDeep(x, y, ray)
		if ok && is[0].Start.Z > 0 {
			any = true
			result = result.Union(is)
		}
	}
	return any, result
}

func (gt *GroupTracer) GetBounds() Bounds {
	if !gt.boundsCurrent {
		if len(gt.tracers) == 0 {
			gt.bounds = EmptyBounds
		} else {
			r := gt.tracers[0].GetBounds()
			for _, t := range gt.tracers[1:] {
				r = r.Union(t.GetBounds())
			}
			gt.bounds = r
		}
		gt.boundsCurrent = true
	}
	return gt.bounds
}

func (gt *GroupTracer) Add(ts ...Tracer) {
	for _, t := range ts {
		gt.tracers = append(gt.tracers, t)
	}
	gt.boundsCurrent = false
	gt.isSorted = false
}

func (gt *GroupTracer) flattenPrunedIntoFacets(rp RenderingParameters, ft *FacetTracer) {
	if !rp.Contains(gt.GetBounds()) {
		return
	}
	for _, t := range gt.tracers {
		asGt, ok := t.(*GroupTracer)
		if ok {
			asGt.flattenPrunedIntoFacets(rp, ft)
		} else {
			pruned := t.Pruned(rp)
			if pruned != nil {
				prunedAsGt, ok := pruned.(*GroupTracer)
				if ok {
					prunedAsGt.flattenPrunedIntoFacets(rp, ft)
				} else {
					ft.Add(pruned)
				}
			}
		}
	}
}

func (gt *GroupTracer) Pruned(rp RenderingParameters) Tracer {
	if !rp.Contains(gt.GetBounds()) {
		return nil
	}
	var bounds Bounds
	if math.IsInf(rp.XMin, 0) || math.IsInf(rp.XMax, 0) || math.IsInf(rp.YMin, 0) || math.IsInf(rp.YMax, 0) {
		bounds = gt.GetBounds()
	} else {
		bounds = Bounds{
			XMin: rp.XMin,
			XMax: rp.XMax,
			YMin: rp.YMin,
			YMax: rp.YMax,
			ZMin: gt.GetBounds().ZMin,
			ZMax: gt.GetBounds().ZMax,
		}
	}
	result := NewFacetTracer(bounds, 16)
	gt.flattenPrunedIntoFacets(rp, result)
	if result.IsEmpty() {
		return nil
	}
	result.Sort()
	return result
}

func (gt *GroupTracer) Sort() {
	sort.Sort(gt)
	gt.isSorted = true
}

func (gt *GroupTracer) Len() int {
	return len(gt.tracers)
}

func (gt *GroupTracer) Less(i, j int) bool {
	return gt.tracers[i].GetBounds().ZMin < gt.tracers[j].GetBounds().ZMin
}

func (gt *GroupTracer) Swap(i, j int) {
	gt.tracers[i], gt.tracers[j] = gt.tracers[j], gt.tracers[i]
}
