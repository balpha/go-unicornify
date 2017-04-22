package core

import (
	"math"
)

type WorldView struct {
	CameraPosition Vector
	LookAtPoint    Vector
	FocalLength    float64
	ux, uy, zero   Vector
	ray            Vector
}

func (wv *WorldView) Init() {
	view := wv.LookAtPoint.Plus(wv.CameraPosition.Neg())
	n := view.Times(1.0 / view.Length())

	wv.ux, wv.uy = CrossAxes(n)
	wv.zero = wv.CameraPosition.Plus(wv.LookAtPoint.Plus(wv.CameraPosition.Neg()).Unit().Times(wv.FocalLength))
}

func (wv WorldView) UnProject(p Vector) Vector {
	pos := wv.zero.Plus(wv.ux.Times(p.X())).Plus(wv.uy.Times(p.Y()))
	return wv.CameraPosition.Plus(pos.Minus(wv.CameraPosition).Unit().Times(p.Z()))
}

func (wv WorldView) ProjectSphere(center Vector, radius float64) SphereProjection {
	cam2c := center.Minus(wv.CameraPosition)
	dist := cam2c.Length()

	ok, intf := IntersectionOfPlaneAndLine(wv.zero, wv.ux, wv.uy, wv.CameraPosition, cam2c)
	if !ok { //FIXME
		return SphereProjection{}
	} else {

		projection := SphereProjection{
			ProjectedCenterOS: wv.CameraPosition.Plus(cam2c.Times(intf[2])),
			ProjectedCenterCS: Vector{intf[0], intf[1], wv.FocalLength},
			WorldView:         wv,
		}
		dir := 1.0
		if intf[2] < 0 {
			dir = -1
		}
		projection.CenterCS = projection.ProjectedCenterCS.Times(dir * dist / projection.ProjectedCenterCS.Length())
		if intf[2] < 0 {
			projection.ProjectedCenterCS[0] *= -1
			projection.ProjectedCenterCS[1] *= -1
		}
		if radius == 0 {
			return projection
		}

		closestToCam := wv.CameraPosition.Plus(cam2c.Times(1 - radius/dist))

		u1, u2 := CrossAxes(cam2c.Times(1 / dist))
		r := 0.0
		for c1 := -1.0; c1 <= 1; c1 += 2 {
			for c2 := -1.0; c2 <= 1; c2 += 2 {
				p := closestToCam.Plus(u1.Times(c1 * radius)).Plus(u2.Times(c2 * radius))
				pr := wv.ProjectSphere(p, 0)
				r = math.Max(r, math.Abs(pr.X()-projection.X()))
				r = math.Max(r, math.Abs(pr.Y()-projection.Y()))
			}
		}

		projection.ProjectedRadius = r
		return projection

	}

}

func (wv *WorldView) Ray(x, y float64) Vector {
	return Vector{x, y, wv.FocalLength}.Unit()
}
