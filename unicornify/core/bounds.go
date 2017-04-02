package core

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

func (b Bounds) ContainsXY(x, y float64) bool {
	return !b.Empty && x >= b.XMin && x <= b.XMax && y >= b.YMin && y <= b.YMax
}

func (b Bounds) ContainsPointsInFrontOfZ(z float64) bool {
	return !b.Empty && z > b.ZMin
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
	return image.Rect(RoundDown(b.XMin), RoundDown(b.YMin), RoundUp(b.XMax), RoundUp(b.YMax))
}
func (b Bounds) MidPoint() Vector {
	if b.Empty {
		return Vector{}
	}
	return Vector{
		(b.XMax + b.XMin) / 2,
		(b.YMax + b.YMin) / 2,
		(b.ZMax + b.ZMin) / 2,
	}
}
