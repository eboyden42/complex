package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
	"sync"
	"time"
)

const width = 20000
const height = 20000
const numThreads = 4
const depth = 150
const palletConcavityPower = 0.2

var pallet []color.RGBA = makePalletVar(depth, palletConcavityPower)

const percentUpdates = 100

type point struct {
	x, y int
}

type write struct {
	p     point
	value color.RGBA
}

func makePalletLinear(depth int) []color.RGBA {
	res := make([]color.RGBA, depth)
	m := -float32(255) / float32(depth)
	for i := 0; i < depth; i++ {
		val := uint8(float32(i)*m + 255)
		clr := color.RGBA{val, val, val, 255}
		res[i] = clr
	}
	return res
}

func makePalletVar(int, p float64) []color.RGBA {
	res := make([]color.RGBA, depth)
	for i := 0; i < depth; i++ {
		val := uint8(variableNormFunc(float64(i), p))
		clr := color.RGBA{val, val, val, 255}
		res[i] = clr
	}
	return res
}

func variableNormFunc(x, p float64) float64 {
	return 255.0 * (1 - math.Pow(x/255.0, p))
}

func inverseNormFunc(x float64) float64 { // this function is too concave for pallet generation
	a := float64(-255.0+math.Sqrt(65029)) / 2.0
	return 1/(x+a) - a // formula for a f(x) = 1/x plot where f(0) = 255 and f(255) = 0
}

func bitMapToNormFloat(num, end int) float64 {
	return (1 / float64(end)) * float64(num) * 2
}

func isInMandelbrot(x0, y0 float64, n int, threshold float64) color.RGBA {
	// iterate the function f(z) = z^2 + c n times
	x2 := 0.0
	y2 := 0.0
	x := 0.0
	y := 0.0

	for i := 0; i < n; i++ {
		if x2+y2 > threshold {
			return pallet[i]
		}
		x2 = x * x
		y2 = y * y
		y = (x+x)*y + y0
		x = x2 - y2 + x0
	}
	return pallet[n-1]
}

func reader(width int, points <-chan point, writes chan<- write, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		p, ok := <-points
		if !ok {
			return
		}
		xval := bitMapToNormFloat(p.x, width)
		yval := bitMapToNormFloat(p.y, width)

		wr := write{p, isInMandelbrot(xval, yval, depth, 4.0)}
		writes <- wr
	}
}

func writer(writes <-chan write, nums chan<- int, img *image.RGBA, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		wr, ok := <-writes
		if !ok {
			return
		}

		img.Set(wr.p.x, wr.p.y, wr.value)

		nums <- 1
	}
}

func tracker(total int, nums <-chan int, wg *sync.WaitGroup) {
	defer wg.Done()

	interval := total / percentUpdates
	partial := 0

	for {
		_, ok := <-nums
		if !ok {
			return
		}
		partial++
		if partial%interval == 0 {
			fmt.Printf("%d %% \n", int(100.0*float64(partial)/float64(total)))
		}
	}
}

func main() {
	startWidth := width/2 - width
	endWidth := width / 2
	startHeight := height/2 - height
	endHeight := height / 2

	rect := image.Rect(startWidth, startHeight, endWidth, endHeight)
	img := image.NewRGBA(rect)

	var wgRead sync.WaitGroup
	var wgWrite sync.WaitGroup
	var wgTracker sync.WaitGroup
	points := make(chan point)
	writes := make(chan write)
	nums := make(chan int)

	wgRead.Add(numThreads)
	for i := 0; i < numThreads; i++ {
		go reader(endWidth, points, writes, &wgRead)
	}

	wgWrite.Add(1)
	go writer(writes, nums, img, &wgWrite)

	wgTracker.Add(1)
	go tracker(width*height, nums, &wgTracker)

	start := time.Now()
	fmt.Println("Issuing point calculations to threads...")

	for x := startWidth; x < endWidth; x++ {
		for y := startHeight; y < endHeight; y++ {
			points <- point{x, y}
		}
	}

	close(points)
	wgRead.Wait()

	close(writes)
	wgWrite.Wait()
	end := time.Now()
	diff := end.Sub(start)
	fmt.Printf("Total time: %.5f \n", diff.Seconds())

	close(nums)
	wgTracker.Wait()

	filename := fmt.Sprintf(fmt.Sprintf("mandelbrot%d.png", width))
	file, err := os.Create(filename)
	if err != nil {
		panic(err)
	}

	err = png.Encode(file, img)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Image written to %s\n", filename)
}
