package core

type RenderingParameters struct {
	PixelSize              float64
	XMin, XMax, YMin, YMax float64
}

func (rp RenderingParameters) Scaled(scale float64) RenderingParameters {
	return RenderingParameters{
		PixelSize: rp.PixelSize / scale,
		XMin:      rp.XMin / scale,
		XMax:      rp.XMax / scale,
		YMin:      rp.YMin / scale,
		YMax:      rp.YMax / scale,
	}
}

func (rp RenderingParameters) Translated(dx, dy float64) RenderingParameters {
	return RenderingParameters{
		PixelSize: rp.PixelSize,
		XMin:      rp.XMin - dx,
		XMax:      rp.XMax - dx,
		YMin:      rp.YMin - dy,
		YMax:      rp.YMax - dy,
	}
}

func (rp RenderingParameters) Contains(b Bounds) bool {
	return b.XMax >= rp.XMin && b.XMin <= rp.XMax && b.YMax >= rp.YMin && b.YMin <= rp.YMax && b.ZMax > 0
}
