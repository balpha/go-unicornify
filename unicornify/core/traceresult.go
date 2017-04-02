package core

import (
	"math"
)

type TraceResult struct {
	Z         float64
	Direction Vector
	Color     Color
}

type TraceInterval struct {
	Start, End TraceResult
}

var EmptyInterval = TraceInterval{TraceResult{0, NoDirection, Color{}}, TraceResult{0, NoDirection, Color{}}}

type TraceIntervals []TraceInterval

var EmptyIntervals = TraceIntervals{}

func (first TraceInterval) Intersect(second TraceInterval) TraceInterval {
	var left, right TraceResult
	if second.Start.Z > first.Start.Z {
		left = second.Start
	} else {
		left = first.Start
	}
	if second.End.Z < first.End.Z {
		right = second.End
	} else {
		right = first.End
	}

	if right.Z <= left.Z {
		return EmptyInterval
	}

	return TraceInterval{left, right}
}

func (i TraceInterval) IsEmpty() bool {
	return i.Start.Z >= i.End.Z
}

func (first TraceIntervals) Intersect(second TraceIntervals) TraceIntervals {
	i1 := 0
	i2 := 0
	result := TraceIntervals{}
	for i1 < len(first) && i2 < len(second) {
		intersection := first[i1].Intersect(second[i2])
		if !intersection.IsEmpty() {
			result = append(result, intersection)
		}
		if i1+1 == len(first) {
			i2++
		} else if i2+1 == len(second) {
			i1++
		} else if first[i1+1].Start.Z <= second[i2+1].Start.Z {
			i1++
		} else {
			i2++
		}
	}
	return result
}

func (first TraceIntervals) Union(second TraceIntervals) TraceIntervals {
	first = first.Intersect(second.Invert())
	i1 := 0
	i2 := 0
	result := make(TraceIntervals, 0, len(first)+len(second))
	for i1 < len(first) || i2 < len(second) {
		if i2 == len(second) || (i1 < len(first) && first[i1].Start.Z <= second[i2].Start.Z) {
			result = append(result, first[i1])
			i1++
			continue
		}
		if i1 == len(first) || (i2 < len(second) && second[i2].Start.Z < first[i1].Start.Z) {
			result = append(result, second[i2])
			i2++
		}
	}
	return result
}

func (is TraceIntervals) Invert() TraceIntervals {
	//TODO: handle inf start&end
	if len(is) == 0 {
		return TraceIntervals{
			TraceInterval{
				Start: TraceResult{math.Inf(-1), NoDirection, Color{}},
				End:   TraceResult{math.Inf(1), NoDirection, Color{}},
			},
		}
	}
	result := make(TraceIntervals, len(is)+1)
	prev := TraceResult{math.Inf(-1), is[0].Start.Direction, is[0].Start.Color}
	for index, i := range is {
		n := TraceInterval{
			Start: prev,
			End:   TraceResult{i.Start.Z, i.Start.Direction.Neg(), i.Start.Color},
		}
		result[index] = n
		prev = TraceResult{i.End.Z, i.End.Direction.Neg(), i.End.Color}
	}
	result[len(is)] = TraceInterval{
		Start: prev,
		End:   TraceResult{math.Inf(1), prev.Direction.Neg(), prev.Color},
	}
	return result
}
