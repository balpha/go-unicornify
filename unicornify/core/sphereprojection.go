package core

type SphereProjection struct {
	CenterCS          Vector // the center in camera space (camera at (0,0,0), Z axis in view direction)
	ProjectedCenterCS Vector // the projection in camera space (note that by definition, Z will always be = focal length)
	ProjectedCenterOS Vector // the projection in original space
	ProjectedRadius   float64
	WorldView         WorldView
}

func (bp SphereProjection) X() float64 {
	return bp.ProjectedCenterCS.X()
}

func (bp SphereProjection) Y() float64 {
	return bp.ProjectedCenterCS.Y()
}

func (bp SphereProjection) Z() float64 {
	return bp.CenterCS.Length()
}
