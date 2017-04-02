package core

import (
	"image/color"
	"math"
	"reflect"
)

type Color struct {
	R, G, B byte
}

var Black = Color{0, 0, 0}
var BlackRGBA = color.RGBA{0, 0, 0, 255}

func (c Color) RGBA() (r, g, b, a uint32) {
	return color.RGBA{c.R, c.G, c.B, 255}.RGBA()
}
func (c Color) ToRGBA() color.RGBA {
	return color.RGBA{
		c.R,
		c.G,
		c.B,
		255,
	}
}
func MixColors(c1 Color, c2 Color, f float64) Color {
	return Color{
		R: MixBytes(c1.R, c2.R, f),
		G: MixBytes(c1.G, c2.G, f),
		B: MixBytes(c1.B, c2.B, f),
	}
}
func MixColorsRGBA(c1 color.RGBA, c2 color.RGBA, f float64) color.RGBA {
	return color.RGBA{
		R: MixBytes(c1.R, c2.R, f),
		G: MixBytes(c1.G, c2.G, f),
		B: MixBytes(c1.B, c2.B, f),
		A: 255,
	}
}
func min(a, b uint8) uint8 {
	if a > b {
		return b
	}
	return a
}
func Darken(c Color, d uint8) Color {
	return Color{
		R: c.R - min(d, c.R),
		G: c.G - min(d, c.G),
		B: c.B - min(d, c.B),
	}
}
func DarkenRGBA(c color.RGBA, d uint8) color.RGBA {
	return color.RGBA{
		R: c.R - min(d, c.R),
		G: c.G - min(d, c.G),
		B: c.B - min(d, c.B),
		A: c.A,
	}
}
func Lighten(c Color, d uint8) Color {
	return Color{
		R: c.R + min(d, 255-c.R),
		G: c.G + min(d, 255-c.G),
		B: c.B + min(d, 255-c.B),
	}
}
func LightenRGBA(c color.RGBA, d uint8) color.RGBA {
	return color.RGBA{
		R: c.R + min(d, 255-c.R),
		G: c.G + min(d, 255-c.G),
		B: c.B + min(d, 255-c.B),
		A: c.A,
	}
}

func v(m1, m2, hue float64) float64 {
	_, hue = math.Modf(hue)
	_, hue = math.Modf(hue + 1)
	if hue < 1./6 {
		return m1 + (m2-m1)*hue*6
	}
	if hue < .5 {
		return m2
	}
	if hue < 2./3 {
		return m1 + (m2-m1)*(2./3-hue)*6
	}
	return m1
}

func Hsl2col(hue, sat, lig int) Color {
	h := float64(hue) / 360
	s := float64(sat) / 100
	l := float64(lig) / 100

	var rf, gf, bf, m1, m2 float64
	if s == 0 {
		rf, gf, bf = 1, 1, 1
	} else {
		if l <= .5 {
			m2 = l * (1.0 + s)
		} else {
			m2 = l + s - (l * s)
		}
		m1 = 2*l - m2
		rf = v(m1, m2, h+1./3)
		gf = v(m1, m2, h)
		bf = v(m1, m2, h-1./3)
	}
	return Color{
		R: byte(255 * rf),
		G: byte(255 * gf),
		B: byte(255 * bf),
	}
}

func ColorFromData(d interface{}, name string, lightness int) Color {
	dv := reflect.ValueOf(d)
	hue := int(dv.FieldByName(name + "Hue").Int())
	sat := int(dv.FieldByName(name + "Sat").Int())
	return Hsl2col(hue, sat, lightness)
}
