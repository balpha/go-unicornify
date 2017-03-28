package unicornify

import (
	"image"
	"image/color"
	"math"
)

type TraceResult struct {
	Z         float64
	Direction Point3d
	Color     Color
}

type TraceInterval struct {
	Start, End TraceResult
}

var EmptyInterval = TraceInterval{TraceResult{0, NoDirection, Color{}}, TraceResult{0, NoDirection, Color{}}}

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

type TraceIntervals []TraceInterval

var EmptyIntervals = TraceIntervals{}

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
	result := TraceIntervals{}
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
	result := TraceIntervals{}
	if len(is) == 0 {
		return TraceIntervals{
			TraceInterval{
				Start: TraceResult{math.Inf(-1), NoDirection, Color{}},
				End:   TraceResult{math.Inf(1), NoDirection, Color{}},
			},
		}
	}
	prev := TraceResult{math.Inf(-1), is[0].Start.Direction, is[0].Start.Color}
	for _, i := range is {
		n := TraceInterval{
			Start: prev,
			End:   TraceResult{i.Start.Z, i.Start.Direction.Neg(), i.Start.Color},
		}
		result = append(result, n)
		prev = TraceResult{i.End.Z, i.End.Direction.Neg(), i.End.Color}
	}
	result = append(result, TraceInterval{
		Start: prev,
		End:   TraceResult{math.Inf(1), prev.Direction.Neg(), prev.Color},
	})
	return result
}

type Tracer interface {
	Trace(x, y float64) (bool, float64, Point3d, Color)
	TraceDeep(x, y float64) (bool, TraceIntervals)
	GetBounds() image.Rectangle
}

func DeepifyTrace(t Tracer, x, y float64) (bool, TraceIntervals) {
	ok, z, dir, col := t.Trace(x, y)
	inter := TraceIntervals{TraceInterval{
		Start: TraceResult{z, dir, col},
		End:   TraceResult{math.Inf(1), dir, col},
	}}
	return ok, inter
}

func UnDeepifyTrace(t Tracer, x, y float64) (bool, float64, Point3d, Color) {
	ok, r := t.TraceDeep(x, y)
	if ok {
		first := r[0].Start
		return true, first.Z, first.Direction, first.Color
	}
	return false, 0, NoDirection, Color{}
}

type WrappingTracer interface {
	Tracer
	Add(tracers ...Tracer)
}

func DrawTracerPartial(t Tracer, img *image.RGBA, yCallback func(int), bounds image.Rectangle, c chan bool) {
	r := bounds.Intersect(t.GetBounds())
	for y := r.Min.Y; y <= r.Max.Y; y++ {
		for x := r.Min.X; x <= r.Max.X; x++ {
			any, _, _, col := t.Trace(float64(x), float64(y))
			if any {
				img.Set(x, y, col)
			}
		}
		if yCallback != nil {
			yCallback(y)
		}
	}
	if c != nil {
		c <- true
	}
}
func DrawTracer(t Tracer, img *image.RGBA, yCallback func(int)) {
	DrawTracerPartial(t, img, yCallback, img.Bounds(), nil)
}
func DrawTracerParallel(t Tracer, img *image.RGBA, yCallback func(int), partsRoot int) {
	full := img.Bounds()
	c := make(chan bool)
	parts := partsRoot * partsRoot
	partsLeft := parts
	for x := 0; x < partsRoot; x++ {
		for y := 0; y < partsRoot; y++ {
			r := image.Rect(full.Dx()*x/partsRoot, full.Dy()*y/partsRoot, full.Dx()*(x+1)/partsRoot-1, full.Dy()*(y+1)/partsRoot-1)
			go DrawTracerPartial(t, img, nil, r, c)
		}
	}
	for partsLeft > 0 {
		<-c
		partsLeft--
		yCallback(full.Dy() * (parts - partsLeft) / parts)
	}
}

// ------- GroupTracer -------

type GroupTracer struct {
	tracers       []Tracer
	bounds        image.Rectangle
	boundsCurrent bool
}

func NewGroupTracer() *GroupTracer {
	return &GroupTracer{}
}

func (gt *GroupTracer) Trace(x, y float64) (bool, float64, Point3d, Color) {
	any := false
	var minz float64 = 0.0
	var col Color = Black
	var dir Point3d
	xi := int(x)
	yi := int(y)
	for _, t := range gt.tracers {
		tr := t.GetBounds()
		if xi < tr.Min.X || xi >= tr.Max.X+1 || yi < tr.Min.Y || yi >= tr.Max.Y+1 {
			continue
		}
		ok, z, thisdir, thiscol := t.Trace(x, y)
		if ok {
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
	xi := int(x)
	yi := int(y)
	for _, t := range t.tracers {
		tr := t.GetBounds()
		if xi < tr.Min.X || xi >= tr.Max.X+1 || yi < tr.Min.Y || yi >= tr.Max.Y+1 {
			continue
		}
		ok, is := t.TraceDeep(x, y)
		if ok {
			any = true
			result = result.Union(is)
		}
	}
	return any, result
}

func (gt *GroupTracer) GetBounds() image.Rectangle {
	if !gt.boundsCurrent {
		if len(gt.tracers) == 0 {
			gt.bounds = image.Rect(-10, -10, -10, -10)
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

// ------- ImageTracer -------

type ImageTracer struct {
	img    *image.RGBA
	bounds image.Rectangle
	z      func(x, y float64) (bool, float64)
}

var NoDirection = Point3d{0, 0, 0}

func (t *ImageTracer) Trace(x, y float64) (bool, float64, Point3d, Color) {
	tr := t.bounds
	xi := int(x)
	yi := int(y)
	if xi < tr.Min.X || xi >= tr.Max.X+1 || yi < tr.Min.Y || yi >= tr.Max.Y+1 {
		return false, 0, NoDirection, Black
	}
	c := t.img.At(xi, yi).(color.RGBA)
	if c.A < 255 {
		return false, 0, NoDirection, Black
	}
	ok, z := t.z(x, y)

	return ok, z, NoDirection, Color{c.R, c.G, c.B}
}

func (t *ImageTracer) TraceDeep(x, y float64) (bool, TraceIntervals) {
	return DeepifyTrace(t, x, y)
}

func (t *ImageTracer) GetBounds() image.Rectangle {
	return t.bounds
}

// ------- DirectionalLightTracer -------

type DirectionalLightTracer struct {
	GroupTracer
	LightDirectionUnit Point3d
}

func (t *DirectionalLightTracer) Trace(x, y float64) (bool, float64, Point3d, Color) {
	ok, z, dir, col := t.GroupTracer.Trace(x, y)
	if !ok {
		return ok, z, dir, col
	}
	dirlen := dir.Length()
	if dirlen == 0 {
		return ok, z, dir, col
	}

	unit := dir.Times(1 / dirlen)
	sp := unit.ScalarProd(t.LightDirectionUnit)

	if sp >= 0 {
		col = Darken(col, uint8(sp*96))
	} else {
		col = Lighten(col, uint8(-sp*48))
	}

	return ok, z, dir, col
}

func (t *DirectionalLightTracer) TraceDeep(x, y float64) (bool, TraceIntervals) {
	return DeepifyTrace(t, x, y)
}

func (t *DirectionalLightTracer) Add(ts ...Tracer) {
	t.GroupTracer.Add(ts...)
}
func (t *DirectionalLightTracer) SetLightDirection(dir Point3d) {
	length := dir.Length()
	if length != 0 {
		dir = dir.Times(1 / length)
	}
	t.LightDirectionUnit = dir
}

func NewDirectionalLightTracer(lightDirection Point3d) *DirectionalLightTracer {
	result := &DirectionalLightTracer{}
	result.SetLightDirection(lightDirection)
	return result
}

// ------- PointLightTracer (experimental, unused) -------

type PointLightTracer struct {
	LightPositions []Point3d
	SourceTracer   Tracer
	HalfLifes      []float64
}

func (t *PointLightTracer) Trace(x, y float64) (bool, float64, Point3d, Color) {
	ok, z, dir, col := t.SourceTracer.Trace(x, y)
	if !ok {
		return ok, z, dir, col
	}
	dirlen := dir.Length()
	unit := Point3d{0, 0, 0}
	if dirlen > 0 {
		unit = dir.Times(1 / dirlen)
	} else {
		return ok, z, dir, col
	}

	lightsum := 0.0
	for i, lightposition := range t.LightPositions {
		lightdir := Point3d{float64(x), float64(y), z}.Shifted(lightposition.Neg())
		lightdirunit := lightdir.Times(1 / lightdir.Length())

		sp := -unit.ScalarProd(lightdirunit)
		if dirlen == 0 {
			sp = 0.5
		}
		if sp < 0 {
			continue
		}
		strength := math.Pow(0.5, lightdir.Length()/t.HalfLifes[i])
		sp = sp * strength
		lightsum += sp
	}

	if lightsum < 0 {
		col = Black
	} else {
		if lightsum <= 0.5 {
			col = Darken(col, uint8((0.5-lightsum)*2*255))
		} else {
			col = Lighten(col, uint8((lightsum-0.5)*96))
		}
	}

	return ok, z, dir, col
}

func (t *PointLightTracer) TraceDeep(x, y float64) (bool, TraceIntervals) {
	return DeepifyTrace(t, x, y)
}

func (t *PointLightTracer) GetBounds() image.Rectangle {
	return t.SourceTracer.GetBounds()
}

func NewPointLightTracer(source Tracer, lightPos ...Point3d) *PointLightTracer {
	result := &PointLightTracer{SourceTracer: source, LightPositions: lightPos}
	return result
}

// ------- DifferenceTracer -------

type DifferenceTracer struct {
	Base, Subtrahend Tracer
}

func (t *DifferenceTracer) TraceDeep(x, y float64) (bool, TraceIntervals) {
	ok1, i1 := t.Base.TraceDeep(x, y)
	ok2, i2 := t.Subtrahend.TraceDeep(x, y)
	if !ok1 {
		return false, EmptyIntervals
	}
	if !ok2 {
		return ok1, i1
	}
	res := i1.Intersect(i2.Invert())
	return len(res) > 0, res
}

func (t *DifferenceTracer) GetBounds() image.Rectangle {
	return t.Base.GetBounds()
}

func (t *DifferenceTracer) Trace(x, y float64) (bool, float64, Point3d, Color) {
	return UnDeepifyTrace(t, x, y)
}

func NewDifferenceTracer(base, subtrahend Tracer) *DifferenceTracer {
	return &DifferenceTracer{base, subtrahend}
}

// ------- IntersectionTracer -------

type IntersectionTracer struct {
	Base, Other Tracer
}

func (t *IntersectionTracer) TraceDeep(x, y float64) (bool, TraceIntervals) {
	ok1, i1 := t.Base.TraceDeep(x, y)
	ok2, i2 := t.Other.TraceDeep(x, y)
	if !ok1 || !ok2 {
		return false, EmptyIntervals
	}
	res := i1.Intersect(i2)
	return len(res) > 0, res
}

func (t *IntersectionTracer) GetBounds() image.Rectangle {
	return t.Base.GetBounds().Intersect(t.Other.GetBounds())
}

func (t *IntersectionTracer) Trace(x, y float64) (bool, float64, Point3d, Color) {
	return UnDeepifyTrace(t, x, y)
}

func NewIntersectionTracer(base, other Tracer) *IntersectionTracer {
	return &IntersectionTracer{base, other}
}

// ------- ScalingTracer -------

type ScalingTracer struct {
	Source Tracer
	Scale  float64
}

func (t *ScalingTracer) TraceDeep(x, y float64) (bool, TraceIntervals) {
	return t.Source.TraceDeep(x/t.Scale, y/t.Scale) // TODO: scale the result?
}

func (t *ScalingTracer) GetBounds() image.Rectangle {
	b := t.Source.GetBounds()
	s := t.Scale
	return rectFromFloats(float64(b.Min.X)*s, float64(b.Min.Y)*s, float64(b.Max.X)*s, float64(b.Max.Y)*s)
}

func (t *ScalingTracer) Trace(x, y float64) (bool, float64, Point3d, Color) {
	return t.Source.Trace(x/t.Scale, y/t.Scale) // TODO: scale the result?
}

func NewScalingTracer(source Tracer, scale float64) *ScalingTracer {
	return &ScalingTracer{source, scale}
}

// ------- TranslatingTracer -------

type TranslatingTracer struct {
	Source         Tracer
	ShiftX, ShiftY float64
}

func (t *TranslatingTracer) TraceDeep(x, y float64) (bool, TraceIntervals) {
	return t.Source.TraceDeep(x-t.ShiftX, y-t.ShiftY)
}

func (t *TranslatingTracer) GetBounds() image.Rectangle {
	b := t.Source.GetBounds()
	return rectFromFloats(float64(b.Min.X)+t.ShiftX, float64(b.Min.Y)+t.ShiftY, float64(b.Max.X)+t.ShiftX, float64(b.Max.Y)+t.ShiftY)
}

func (t *TranslatingTracer) Trace(x, y float64) (bool, float64, Point3d, Color) {
	return t.Source.Trace(x-t.ShiftX, y-t.ShiftY)
}

func NewTranslatingTracer(source Tracer, dx, dy float64) *TranslatingTracer {
	return &TranslatingTracer{source, dx, dy}
}
