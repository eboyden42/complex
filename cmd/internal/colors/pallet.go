package colors

import (
	"image/color"
	"math"
)

type Pallet []color.RGBA

func inverseNormFunc(x float64) float64 { // this function is too concave for pallet generation but it's pretty cool so I'm leaving it in the codebase
	a := float64(-255.0+math.Sqrt(65029)) / 2.0
	return 1/(x+a) - a // formula for a f(x) = 1/x plot where f(0) = 255 and f(255) = 0
}

func variableNormFunc(x, p float64) float64 {
	return 255.0 * (1 - math.Pow(x/255.0, p))
}

func NewPalletLinear(depth int) Pallet {
	res := make([]color.RGBA, depth)
	m := -float32(255) / float32(depth)
	for i := 0; i < depth; i++ {
		val := uint8(float32(i)*m + 255)
		clr := color.RGBA{val, val, val, 255}
		res[i] = clr
	}
	return Pallet(res)
}

func NewPalletVar(depth int, p float64) Pallet {
	res := make([]color.RGBA, depth)
	for i := 0; i < depth; i++ {
		val := uint8(variableNormFunc(float64(i), p))
		clr := color.RGBA{val, val, val, 255}
		res[i] = clr
	}
	return Pallet(res)
}
