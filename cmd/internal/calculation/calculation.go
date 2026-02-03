package calculation

type IntPoint struct {
	X0, Y0 int
}

type FloatPoint struct {
	x0, y0 float64
}

func (p FloatPoint) IsInMandelbrot(iterations int, threshold float64) (bool, int) {
	// calculate if a point p is in the mandelbrot set
	// return boolean on if it's in the set and an int representing the number of iterations used
	// if returns true, then used iterations = iterations

	x2 := 0.0
	y2 := 0.0
	x := 0.0
	y := 0.0

	for i := 0; i < iterations; i++ {
		if x2+y2 > threshold {
			return false, i + 1
		}
		x2 = x * x
		y2 = y * y
		y = (x+x)*y + p.y0
		x = x2 - y2 + p.x0
	}
	return true, iterations
}

func IntNormToFloat(num, inputRange int, outputRange float64) float64 {
	// takes in an int num from [-inputRange, inputRange] returns a float from [-outputRange, outputRange]
	// if input is out of bounds clamps values
	if num < -inputRange {
		return -float64(outputRange)
	}
	if num > inputRange {
		return float64(outputRange)
	}
	return (1 / float64(inputRange)) * float64(num) * outputRange
}

func FloatNormToFloat(num, inputRange, outputRange float64) float64 {
	if num < -inputRange {
		return -outputRange
	}
	if num > inputRange {
		return outputRange
	}
	return (1 / inputRange) * num * outputRange
}

func NewFloatPoint(intPoint IntPoint) FloatPoint {
	return FloatPoint{float64(intPoint.X0), float64(intPoint.Y0)}
}

func (p *FloatPoint) Normalize(inputRange, outputRange float64) {
	p.x0 = FloatNormToFloat(p.x0, inputRange, outputRange)
	p.y0 = FloatNormToFloat(p.y0, inputRange, outputRange)
}
