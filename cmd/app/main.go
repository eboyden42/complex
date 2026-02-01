package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
	"runtime"
	"sync"
	"time"
)

var size int
var numThreads int
var depth int
var palletConcavityPower float64
var pallet []color.RGBA

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

func makePalletVar(depth int, p float64) []color.RGBA {
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

func inverseNormFunc(x float64) float64 { // this function is too concave for pallet generation but it's pretty cool so I'm leaving it in the codebase
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
	processors := runtime.NumCPU()

	// arg parsing
	flag.IntVar(&size, "size", 2000, "Specify the size of the image. Example -size=3000 produces a 3000x3000 bitmap png.")
	flag.IntVar(&numThreads, "threads", processors, "Specify the number of threads used to calculate the status of each point.")
	flag.IntVar(&depth, "depth", 150, "Specify the number of iterations that will be performed on each complex number before it is determined to be inside the mandelbrot set.")
	flag.Float64Var(&palletConcavityPower, "concavity", 0.2, "Specify the power of the function used to generate a pallet.")
	filename := flag.String("out", "mandelbrot.png", "Specify the output file name.")

	flag.Parse()

	// ensure filename ends with .png
	if (*filename)[len(*filename)-4:] != ".png" {
		*filename += ".png"
	}

	// create pallet
	pallet = makePalletVar(depth, palletConcavityPower)

	// calculate dimensions
	startWidth := size/2 - size
	endWidth := size / 2
	startHeight := startWidth
	endHeight := endWidth

	// create rectangle and image
	rect := image.Rect(startWidth, startHeight, endWidth, endHeight)
	img := image.NewRGBA(rect)

	// create WaitGroups and channels
	var wgRead sync.WaitGroup
	var wgWrite sync.WaitGroup
	var wgTracker sync.WaitGroup
	chanSize := size * size
	points := make(chan point, chanSize)
	writes := make(chan write, chanSize)
	nums := make(chan int, chanSize)

	// create threads
	fmt.Printf("Creating %d threads\n", numThreads)
	wgRead.Add(numThreads)
	for i := 0; i < numThreads-2; i++ {
		go reader(endWidth, points, writes, &wgRead)
	}

	wgWrite.Add(1)
	go writer(writes, nums, img, &wgWrite)

	wgTracker.Add(1)
	go tracker(size*size, nums, &wgTracker)

	// starting timer and issuing points to pipeline
	start := time.Now()
	fmt.Println("Issuing point calculations to threads...")

	for x := startWidth; x < endWidth; x++ {
		for y := startHeight; y < endHeight; y++ {
			points <- point{x, y}
		}
	}

	// closing channels and waiting for routines
	close(points)
	wgRead.Wait()

	close(writes)
	wgWrite.Wait()

	close(nums)
	wgTracker.Wait()

	// calculating time
	end := time.Now()
	diff := end.Sub(start)
	fmt.Printf("Total time: %.5f \n", diff.Seconds())

	// writing output image
	file, err := os.Create(*filename)
	if err != nil {
		panic(err)
	}

	err = png.Encode(file, img)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Image written to %s\n", *filename)
}
