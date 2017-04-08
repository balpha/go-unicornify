package rendering

import (
	. "bitbucket.org/balpha/go-unicornify/unicornify/core"
	"sort"
)

type GroupTracer struct {
	tracers       []Tracer
	bounds        Bounds
	boundsCurrent bool
}

func NewGroupTracer() *GroupTracer {
	return &GroupTracer{}
}

func (gt *GroupTracer) Trace(x, y float64) (bool, float64, Vector, Color) {
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
			continue
		}
		ok, z, thisdir, thiscol := t.Trace(x, y)
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

func (t *GroupTracer) TraceDeep(x, y float64) (bool, TraceIntervals) {
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
		ok, is := t.TraceDeep(x, y)
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
}

func (gt *GroupTracer) SubdivideAndSort() {
	const n = 2
	const min = 5

	subs := make([]*GroupTracer, n*n)
	b := gt.GetBounds()
	if b.Empty {
		return
	}
	count := 0
	for _, t := range gt.tracers {
		mid := t.GetBounds().MidPoint()
		bucketx := Round(float64(n-1) * (mid.X() - b.XMin) / b.Dx())
		buckety := Round(float64(n-1) * (mid.Y() - b.YMin) / b.Dy())
		bucket := bucketx + n*buckety
		sub := subs[bucket]
		if sub == nil {
			sub = NewGroupTracer()
			subs[bucket] = sub
			count++
		}
		sub.Add(t)
	}
	if count == 1 {
		sort.Sort(gt)
		return
	}
	gt.tracers = make([]Tracer, 0, count)
	for _, sub := range subs {
		if sub == nil {
			continue
		}
		if len(sub.tracers) < min {
			for _, sst := range sub.tracers {
				gt.Add(sst)
			}
		} else {
			sub.SubdivideAndSort()
			gt.Add(sub)
		}
	}
	sort.Sort(gt)
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
