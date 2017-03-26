package unicornify

import (
	"math"
)

type WorldView interface {
	ProjectBall(*Ball)
}

type PerspectiveWorldView struct {
	CameraPosition Point3d
	LookAtPoint    Point3d
	FocalLength    float64
}

func (wv PerspectiveWorldView) ProjectBall(b *Ball) {
	cam2c := b.Center.Minus(wv.CameraPosition)
	view := wv.LookAtPoint.Shifted(wv.CameraPosition.Neg())
	n := view.Times(1.0 / view.Length())
	n1, n2, n3 := n.Decompose()

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

	ok, intf := IntersectionOfPlaneAndLine(wv.CameraPosition.Shifted(n.Times(wv.FocalLength)), ux, uy, wv.CameraPosition, cam2c)
	if !ok { //FIXME
		b.Projection = BallProjection{}
		b.Projection.Radius = b.Radius
	} else {
		
		b.Projection = BallProjection{
			ProjectedCenterOS: wv.CameraPosition.Shifted(cam2c.Times(intf[2])),
			ProjectedCenterCS: Point3d{intf[0], intf[1], wv.FocalLength},
		}
		b.Projection.CenterCS = b.Projection.ProjectedCenterCS.Times(cam2c.Length() / b.Projection.ProjectedCenterCS.Length())
		
		count := 0.0
		max := 0.0
		for dx := -1.0; dx <= 1; dx += 1 {
			for dy := -1.0; dy <= 1; dy += 1 {
				for dz := -1.0; dz <= 1; dz += 1 {
					shift := Point3d{dx, dy, dz}
					if shift.Length() == 0 {
						continue
					}
					shift = shift.Times(b.Radius / shift.Length())
					ok, intf := IntersectionOfPlaneAndLine(wv.CameraPosition.Shifted(n.Times(wv.FocalLength)), ux, uy, wv.CameraPosition, b.Center.Shifted(shift).Minus(wv.CameraPosition))
					if ok {
						count++
						rp := Point3d{intf[0], intf[1], 0}
						max = math.Max(max, math.Sqrt(sqr(rp[0]-b.Projection.X())+sqr(rp[1]-b.Projection.Y())))
					}
				}
			}

		}

		b.Projection.Radius = max
	}
	return

}
