package unicornify

import (
	"math"
)

type WorldView struct {
	CameraPosition Point3d
	LookAtPoint    Point3d
	FocalLength    float64
	ux, uy, zero   Point3d
}

// Given a vector v, returns the two vectors that form a right-hand rule system
// (u1, u2, v) such that u2 points upward. If v is a unit vector, then so are u1 and u2.
func CrossAxes(v Point3d) (u1, u2 Point3d) {
	n1, n2, n3 := v.Decompose()

	var x1, x3 float64

	if n1 != 0 {
		x3 = math.Sqrt(1 / (n3*n3/(n1*n1) + 1))
		if n1 > 0 {
			x3 = -x3
		}
		x1 = -x3 * n3 / n1
	} else if n3 != 0 {
		x1 = math.Sqrt(1 / (n1*n1/(n3*n3) + 1))
		if n3 < 0 {
			x1 = -x1
		}
		x3 = -x1 * n1 / n3
	} else { // both 0 -- looking down
		x1 = 1
		x3 = 0
	}

	ux := Point3d{x1, 0, x3}

	// cross product of ux and normal (=uz) gives the y axis but in the wrong direction
	// (because x-z-y is not a right-hand rule system)
	y1 := -(-x3 * n2)
	y2 := -(x3*n1 - x1*n3)
	y3 := -(x1 * n2)

	uy := Point3d{y1, y2, y3}

	return ux, uy
}

func (wv *WorldView) Init() {
	view := wv.LookAtPoint.Shifted(wv.CameraPosition.Neg())
	n := view.Times(1.0 / view.Length())
	wv.ux, wv.uy = CrossAxes(n)
	wv.zero = wv.CameraPosition.Shifted(wv.LookAtPoint.Shifted(wv.CameraPosition.Neg()).Unit().Times(wv.FocalLength))
}

func (wv WorldView) UnProject(p Point3d) Point3d {
	pos := wv.zero.Shifted(wv.ux.Times(p.X())).Shifted(wv.uy.Times(p.Y()))
	return wv.CameraPosition.Shifted(pos.Minus(wv.CameraPosition).Unit().Times(p.Z()))
}

func (wv WorldView) ProjectBall(b *Ball) BallProjection {
	cam2c := b.Center.Minus(wv.CameraPosition)
	view := wv.LookAtPoint.Shifted(wv.CameraPosition.Neg())
	n := view.Times(1.0 / view.Length())

	ok, intf := IntersectionOfPlaneAndLine(wv.CameraPosition.Shifted(n.Times(wv.FocalLength)), wv.ux, wv.uy, wv.CameraPosition, cam2c)
	if !ok { //FIXME
		return BallProjection{BaseBall: *b}
	} else {

		projection := BallProjection{
			ProjectedCenterOS: wv.CameraPosition.Shifted(cam2c.Times(intf[2])),
			ProjectedCenterCS: Point3d{intf[0], intf[1], wv.FocalLength},
			WorldView:         wv,
			BaseBall:          *b,
		}
		projection.CenterCS = projection.ProjectedCenterCS.Times(cam2c.Length() / projection.ProjectedCenterCS.Length())
		if b.Radius == 0 {
			return projection
		}

		closestToCam := wv.CameraPosition.Shifted(cam2c.Times(1 - b.Radius/cam2c.Length()))
		secondPoint := closestToCam.Shifted(wv.uy.Times(b.Radius))

		p1 := wv.ProjectBall(NewBallP(closestToCam, 0, Color{}))
		p2 := wv.ProjectBall(NewBallP(secondPoint, 0, Color{}))

		projection.ProjectedRadius = p1.ProjectedCenterCS.Minus(p2.ProjectedCenterCS).Length()

		return projection

	}

}
