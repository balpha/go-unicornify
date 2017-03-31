package unicornify

import (
	"image"
	"math"
)

type Bounds struct {
	XMin, XMax, YMin, YMax, ZMin, ZMax float64
	Empty                              bool
}

var EmptyBounds = Bounds{0, 0, 0, 0, 0, 0, true}

func (b Bounds) Union(o Bounds) Bounds {
	if b.Empty && o.Empty {
		return EmptyBounds
	}
	if b.Empty {
		return o
	}
	if o.Empty {
		return b
	}
	return Bounds{
		XMin:  math.Min(b.XMin, o.XMin),
		XMax:  math.Max(b.XMax, o.XMax),
		YMin:  math.Min(b.YMin, o.YMin),
		YMax:  math.Max(b.YMax, o.YMax),
		ZMin:  math.Min(b.ZMin, o.ZMin),
		ZMax:  math.Max(b.ZMax, o.ZMax),
		Empty: false,
	}
}

func (b Bounds) Intersect(o Bounds) Bounds {
	if b.Empty || o.Empty {
		return EmptyBounds
	}
	res := Bounds{
		XMin:  math.Max(b.XMin, o.XMin),
		XMax:  math.Min(b.XMax, o.XMax),
		YMin:  math.Max(b.YMin, o.YMin),
		YMax:  math.Min(b.YMax, o.YMax),
		ZMin:  math.Max(b.ZMin, o.ZMin),
		ZMax:  math.Min(b.ZMax, o.ZMax),
		Empty: false,
	}
	if res.XMin > res.XMax || res.YMin > res.YMax || res.ZMin > res.ZMax {
		return EmptyBounds
	}
	return res
}

func RenderingBoundsForBalls(bps ...BallProjection) Bounds {
	res := EmptyBounds
	for _, bp := range bps {
		r := Bounds{
			XMin:  bp.X() - bp.ProjectedRadius,
			XMax:  bp.X() + bp.ProjectedRadius,
			YMin:  bp.Y() - bp.ProjectedRadius,
			YMax:  bp.Y() + bp.ProjectedRadius,
			ZMin:  bp.Z() - bp.BaseBall.Radius,
			ZMax:  bp.Z() + bp.BaseBall.Radius,
			Empty: false,
		}
		res = res.Union(r)
	}
	return res
}

func (b Bounds) ContainsXY(x, y float64) bool {
	return !b.Empty && x >= b.XMin && x <= b.XMax && y >= b.YMin && y <= b.YMax
}

func (b Bounds) Dx() float64 {
	if b.Empty {
		return 0
	}
	return b.XMax - b.XMin
}
func (b Bounds) Dy() float64 {
	if b.Empty {
		return 0
	}
	return b.YMax - b.YMin
}
func (b Bounds) ToRect() image.Rectangle {
	if b.Empty {
		return image.Rect(0, 0, 0, 0)
	}
	return image.Rect(roundDown(b.XMin), roundDown(b.YMin), roundUp(b.XMax), roundUp(b.YMax))
}
