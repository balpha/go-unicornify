package core

import ()

type WorldView struct {
	CameraPosition Vector
	LookAtPoint    Vector
	FocalLength    float64
	ux, uy, zero   Vector
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
	view := wv.LookAtPoint.Plus(wv.CameraPosition.Neg())
	n := view.Times(1.0 / view.Length())

	ok, intf := IntersectionOfPlaneAndLine(wv.CameraPosition.Plus(n.Times(wv.FocalLength)), wv.ux, wv.uy, wv.CameraPosition, cam2c)
	if !ok { //FIXME
		return SphereProjection{}
	} else {

		projection := SphereProjection{
			ProjectedCenterOS: wv.CameraPosition.Plus(cam2c.Times(intf[2])),
			ProjectedCenterCS: Vector{intf[0], intf[1], wv.FocalLength},
			WorldView:         wv,
		}
		projection.CenterCS = projection.ProjectedCenterCS.Times(cam2c.Length() / projection.ProjectedCenterCS.Length())
		if radius == 0 {
			return projection
		}

		closestToCam := wv.CameraPosition.Plus(cam2c.Times(1 - radius/cam2c.Length()))
		secondPoint := closestToCam.Plus(wv.uy.Times(radius))

		p1 := wv.ProjectSphere(closestToCam, 0)
		p2 := wv.ProjectSphere(secondPoint, 0)

		projection.ProjectedRadius = p1.ProjectedCenterCS.Minus(p2.ProjectedCenterCS).Length()

		return projection

	}

}
