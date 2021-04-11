package unicornify

import (
	. "github.com/balpha/go-unicornify/unicornify/core"
	. "github.com/balpha/go-unicornify/unicornify/elements"
	"github.com/balpha/gopyrand"
	"math"
)

type GrassData struct {
	Horizon        float64
	Wind           float64
	Color1, Color2 Color
}

func (d *GrassData) Randomize(rand *pyrand.Random) {
	_ = rand.RandBits(64)
	d.Wind = 1.6*rand.Random() - 0.8 // not yet used
}

func GrassSandwich(groundY float64, bgdata BackgroundData, grassdata GrassData, shift Point2d, scale float64, imageSize int) Thing {
	var grassSize float64 = 20000
	var bladeDistance float64 = 4
	var bladeDiameter float64 = 2
	fb1 := NewBall(-grassSize/2, groundY, -grassSize/2, 1, Color{255, 0, 0})
	fb2 := NewBall(grassSize/2, groundY, -grassSize/2, 1, Color{0, 255, 0})
	fb3 := NewBall(-grassSize/2, groundY, grassSize/2, 1, Color{0, 0, 255})

	swf := func(x, y float64, bOk bool, bV, bW, bZ float64, tOk bool, tV, tW, tZ float64) (bool, TraceIntervals) {

		if !bOk || !tOk {
			return false, TraceIntervals{}
		}

		I := Vector{tV * grassSize, 0, tW * grassSize}
		O := Vector{bV * grassSize, 15, bW * grassSize}
		C := O.Minus(I)

		sqrCX := Sqr(C.X())
		sqrCY := Sqr(C.Y())
		sqrCZ := Sqr(C.Z())

		sqrIY := Sqr(I.Y())

		CXCY := C.X() * C.Y()
		CZCY := C.Z() * C.Y()

		CXIY := C.X() * I.Y()
		CZIY := C.Z() * I.Y()

		twoIYCY := 2 * I.Y() * C.Y()

		r0 := bladeDiameter
		sqrr0 := Sqr(r0)

		crossingCells := 5 * float64(RoundUp(math.Abs(C.X())/bladeDistance)+RoundUp(math.Abs(C.Z())/bladeDistance))
		d := C.Times(1.0 / crossingCells)

		prevX := -999999999
		prevY := -999999999
		landColor := MixColors(bgdata.Color("Land", bgdata.LandLight), bgdata.Color("Land", bgdata.LandLight/2), (x*scale+shift[0])/float64(imageSize))

		for n := float64(0); n <= crossingCells; n++ {

			p := I.Plus(d.Times(n))
			cXb := float64(RoundDown(p.X()/bladeDistance)) * bladeDistance
			cYb := float64(RoundDown(p.Z()/bladeDistance)) * bladeDistance
			cxb := int(cXb)
			cyb := int(cYb)
			if cxb == prevX && cyb == prevY {
				continue
			}
			prevX = cxb
			prevY = cyb

			closest := EmptyInterval

			for ix := 0; ix <= 0; ix++ {
				for iy := 0; iy <= 0; iy++ {

					cx := cxb - ix
					cy := cyb - iy
					cX := float64(cx)
					cY := float64(cy)

					randomish1 := float64(QuickRand2(cx, cy)) / 2147483648.0
					randomish2 := float64(QuickRand2(cy, cx)) / 2147483648.0
					randomish3 := float64(QuickRand2(cx+cy, cy)) / 2147483648.0
					randomish4 := float64(QuickRand2(cy, cx+cy)) / 2147483648.0

					B := Vector{cX + bladeDiameter + randomish4*(2*bladeDistance-2*bladeDiameter), 15, cY + bladeDiameter + randomish2*(2*bladeDistance-2*bladeDiameter)}
					T := Vector{cX + 2*randomish3*bladeDistance, 0, cY + 2*randomish1*bladeDistance}
					D := B.Minus(T)

					sqrDX := Sqr(D.X())
					sqrDY := Sqr(D.Y())
					sqrDZ := Sqr(D.Z())

					twoDYDX := 2 * D.Y() * D.X()
					twoDYDZ := 2 * D.Y() * D.Z()

					Q := I.Minus(T)

					m1 := sqrDY*sqrCX - 2*D.Y()*D.X()*CXCY + sqrDX*sqrCY +
						sqrDY*sqrCZ - 2*D.Y()*D.Z()*CZCY + sqrDZ*sqrCY -
						sqrr0*sqrCY

					m2 := 2*Q.X()*C.X()*sqrDY - twoDYDX*(Q.X()*C.Y()+CXIY) + sqrDX*twoIYCY +
						2*Q.Z()*C.Z()*sqrDY - twoDYDZ*(Q.Z()*C.Y()+CZIY) + sqrDZ*twoIYCY -
						sqrr0*twoIYCY

					m3 := sqrDY*Sqr(Q.X()) - twoDYDX*Q.X()*I.Y() + sqrDX*sqrIY +
						sqrDY*Sqr(Q.Z()) - twoDYDZ*Q.Z()*I.Y() + sqrDZ*sqrIY -
						sqrr0*sqrIY

					ep := m2 / m1 //FIXME: zero
					eq := m3 / m1

					disc := Sqr(ep)/4 - eq

					if disc >= 0 {
						t := -ep/2 - math.Sqrt(disc)
						k := (I.Y() + t*C.Y()) / D.Y()

						if k >= 0 && k <= 1 {
							z := tZ + t*(bZ-tZ)

							dir := I.Plus(C.Times(t)).Minus(T.Plus(D.Times(k)))
							dir[1] = -0.1

							if closest.IsEmpty() || closest.Start.Z > z {
								closest = TraceInterval{
									TraceResult{z, dir, MixColors(grassdata.Color1, grassdata.Color2, k)},
									TraceResult{z + k*r0, dir.Neg() /*fixme*/, MixColors(grassdata.Color1, grassdata.Color2, k)},
								}
							}
							if false {
								return true, TraceIntervals{
									TraceInterval{
										TraceResult{z, dir, MixColors(grassdata.Color1, grassdata.Color2, k)},
										TraceResult{z + k*r0, dir.Neg() /*fixme*/, MixColors(grassdata.Color1, grassdata.Color2, k)},
									},
									TraceInterval{
										TraceResult{bZ - 0.1 /*fixme*/, Vector{0, -1, 0}, landColor},
										TraceResult{bZ /*fixme*/, Vector{0, 1, 0}, landColor},
									},
								}
							}
						}
					}
				}
			}
			if !closest.IsEmpty() {
				return true, TraceIntervals{
					closest,
					TraceInterval{
						TraceResult{bZ - 0.1 /*fixme*/, Vector{0, -1, 0}, landColor},
						TraceResult{bZ /*fixme*/, Vector{0, 1, 0}, landColor},
					},
				}
			}

		}

		return true, TraceIntervals{
			TraceInterval{
				TraceResult{bZ - 0.1 /*fixme*/, Vector{0, -1, 0}, landColor},
				TraceResult{bZ /*fixme*/, Vector{0, 1, 0}, landColor},
			},
		}
	}

	return NewSandwich(fb1, fb2, fb3, Vector{0, -15, 0}, swf)
}
