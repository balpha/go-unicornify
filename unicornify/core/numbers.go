package core

import (
	"math"
)

const DEGREE = math.Pi / 180

func MixBytes(b1, b2 byte, f float64) byte {
	return b1 + byte(f*(float64(b2)-float64(b1))+.5)
}

func MixFloats(f1, f2, f float64) float64 {
	return f1 + f*(f2-f1)
}

func Round(v float64) int {
	if v >= 0 {
		return int(v + .5)
	}
	return int(v - .5)
}
func RoundUp(v float64) int {
	i, f := math.Modf(v)
	if f > 0 {
		return int(i) + 1
	}
	return int(i) // note that this is also correct for f < 0
}
func RoundDown(v float64) int {
	i, f := math.Modf(v)
	if f < 0 {
		return int(i) - 1
	}
	return int(i)
}

func Sqr(x float64) float64 {
	return x * x
}

func Min(a int, b int) int {
	if a <= b {
		return a
	}
	return b
}
func Max(a int, b int) int {
	if a >= b {
		return a
	}
	return b
}
func QuickRand(seed int) int {
	return (1103515245*seed + 12345) % 2147483648
}
func QuickRand2(seed1, seed2 int) int {
	return QuickRand(QuickRand(seed1*seed2) * QuickRand(seed1-seed2))
}
