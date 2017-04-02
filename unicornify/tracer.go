package unicornify

import (
	"image"
	"image/color"
	"math"
	"sort"
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

type Tracer interface {
	Trace(x, y float64) (bool, float64, Point3d, Color)
	TraceDeep(x, y float64) (bool, TraceIntervals)
	GetBounds() Bounds
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
	r := bounds.Intersect(t.GetBounds().ToRect())
	for y := r.Min.Y; y <= r.Max.Y; y++ {
		for x := r.Min.X; x <= r.Max.X; x++ {
			any, _, _, col := t.Trace(float64(x), float64(y))
			if any {
				img.SetRGBA(x, y, col.ToRGBA())
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
		if yCallback != nil {
			yCallback(full.Dy() * (parts - partsLeft) / parts)
		}
	}
}

// ------- GroupTracer -------

type GroupTracer struct {
	tracers       []Tracer
	bounds        Bounds
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
	for _, t := range gt.tracers {
		b := t.GetBounds()
		if !b.ContainsXY(x, y) {
			continue
		}
		if any && !b.ContainsPointsInFrontOfZ(minz) {
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
	for _, t := range t.tracers {
		if !t.GetBounds().ContainsXY(x, y) {
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
		bucketx := round(float64(n-1) * (mid.X() - b.XMin) / b.Dx())
		buckety := round(float64(n-1) * (mid.Y() - b.YMin) / b.Dy())
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
				//_=sst
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

// ------- ImageTracer -------

type ImageTracer struct {
	img    *image.RGBA
	bounds Bounds
	z      func(x, y float64) (bool, float64)
}

var NoDirection = Point3d{0, 0, 0}

func (t *ImageTracer) Trace(x, y float64) (bool, float64, Point3d, Color) {
	if !t.bounds.ContainsXY(x, y) {
		return false, 0, NoDirection, Black
	}
	c := t.img.At(round(x), round(y)).(color.RGBA)
	if c.A < 255 {
		return false, 0, NoDirection, Black
	}

	ok, z := t.z(x, y)

	return ok, z, NoDirection, Color{c.R, c.G, c.B}
}

func (t *ImageTracer) TraceDeep(x, y float64) (bool, TraceIntervals) {
	return DeepifyTrace(t, x, y)
}

func (t *ImageTracer) GetBounds() Bounds {
	return t.bounds
}

// ------- DirectionalLightTracer -------

type DirectionalLightTracer struct {
	GroupTracer
	LightDirectionUnit Point3d
	Lighten, Darken    float64
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
		col = Darken(col, uint8(sp*t.Darken))
	} else {
		col = Lighten(col, uint8(-sp*t.Lighten))
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

func NewDirectionalLightTracer(lightDirection Point3d, lighten, darken float64) *DirectionalLightTracer {
	result := &DirectionalLightTracer{Lighten: lighten, Darken: darken}
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

func (t *PointLightTracer) GetBounds() Bounds {
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

func (t *DifferenceTracer) GetBounds() Bounds {
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
	bounds      Bounds
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

func (t *IntersectionTracer) GetBounds() Bounds {
	return t.bounds
}

func (t *IntersectionTracer) Trace(x, y float64) (bool, float64, Point3d, Color) {
	return UnDeepifyTrace(t, x, y)
}

func NewIntersectionTracer(base, other Tracer) *IntersectionTracer {
	return &IntersectionTracer{base, other, base.GetBounds().Intersect(other.GetBounds())}
}

// ------- ScalingTracer -------

type ScalingTracer struct {
	Source Tracer
	Scale  float64
	bounds Bounds
}

func (t *ScalingTracer) TraceDeep(x, y float64) (bool, TraceIntervals) {
	return t.Source.TraceDeep(x/t.Scale, y/t.Scale) // TODO: scale the result?
}

func (t *ScalingTracer) GetBounds() Bounds {
	return t.bounds
}

func (t *ScalingTracer) Trace(x, y float64) (bool, float64, Point3d, Color) {
	ok, z, dir, col := t.Source.Trace(x/t.Scale, y/t.Scale) // TODO: scale the result?
	return ok, z * t.Scale, dir, col
}

func NewScalingTracer(source Tracer, scale float64) *ScalingTracer {
	b := source.GetBounds()
	if !b.Empty {
		b.XMin *= scale
		b.XMax *= scale
		b.YMin *= scale
		b.YMax *= scale
	}
	return &ScalingTracer{source, scale, b}
}

// ------- TranslatingTracer -------

type TranslatingTracer struct {
	Source         Tracer
	ShiftX, ShiftY float64
	bounds         Bounds
}

func (t *TranslatingTracer) TraceDeep(x, y float64) (bool, TraceIntervals) {
	return t.Source.TraceDeep(x-t.ShiftX, y-t.ShiftY)
}

func (t *TranslatingTracer) GetBounds() Bounds {
	return t.bounds
}

func (t *TranslatingTracer) Trace(x, y float64) (bool, float64, Point3d, Color) {
	return t.Source.Trace(x-t.ShiftX, y-t.ShiftY)
}

func NewTranslatingTracer(source Tracer, dx, dy float64) *TranslatingTracer {
	b := source.GetBounds()
	if !b.Empty {
		b.XMin += dx
		b.XMax += dx
		b.YMin += dy
		b.YMax += dy
	}
	return &TranslatingTracer{source, dx, dy, b}
}

// ------- ShadowCastingTracer -------

type ShadowCastingTracer struct {
	WorldView, LightView      WorldView
	SourceTracer, LightTracer Tracer
	LightProjection           BallProjection
	Lighten, Darken           float64
}

func (t *ShadowCastingTracer) Trace(x, y float64) (bool, float64, Point3d, Color) {
	ok, z, dir, col := t.SourceTracer.Trace(x, y)
	if !ok {
		return ok, z, dir, col
	}
	origPoint := t.WorldView.UnProject(Point3d{x, y, z})
	lp := t.LightView.ProjectBall(NewBallP(origPoint, 0, Color{}))

	lok, lz, ldir, _ := t.LightTracer.Trace(lp.X(), lp.Y())

	seeing := !lok || lz >= origPoint.Minus(t.LightView.CameraPosition).Length()-0.01

	if !seeing {
		col = Darken(col, uint8(t.Darken))
	} else {
		rayUnit := Point3d{lp.X(), lp.Y(), t.LightView.FocalLength}.Unit()
		sp := ldir.Unit().ScalarProd(rayUnit)

		if sp > 0 { // Given a completely realistic world with no rounding errors, this wouldn't happen.
			col = Darken(col, uint8((1-sp)*t.Darken))
		} else if sp < 0 {
			sp = -sp
			if sp < 0.5 {
				col = Darken(col, uint8((0.5-sp)*t.Darken*2))
			} else {
				col = Lighten(col, uint8((sp-0.5)*t.Lighten*2))
			}

		}
	}
	return ok, z, dir, col
}

func (t *ShadowCastingTracer) TraceDeep(x, y float64) (bool, TraceIntervals) {
	return DeepifyTrace(t, x, y)
}

func (t *ShadowCastingTracer) GetBounds() Bounds {
	return t.SourceTracer.GetBounds()
}

func NewShadowCastingTracer(source Tracer, worldView WorldView, shadowCaster Thing, lightPos, lightTarget Point3d, lighten, darken float64) *ShadowCastingTracer {
	lightView := WorldView{
		CameraPosition: lightPos,
		LookAtPoint:    lightTarget,
		FocalLength:    1, // doesn't matter
	}
	lightView.Init()
	lightTracer := shadowCaster.GetTracer(lightView)
	lightProjection := worldView.ProjectBall(NewBallP(lightPos, 1, Color{}))

	result := &ShadowCastingTracer{
		SourceTracer:    source,
		LightTracer:     lightTracer,
		WorldView:       worldView,
		LightView:       lightView,
		LightProjection: lightProjection,
		Lighten:         lighten,
		Darken:          darken,
	}
	return result
}
