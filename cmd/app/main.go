package main

import (
	"flag"
	"fmt"
	"image/color"
	"runtime"
	"sync"
	"time"

	"github.com/eboyden42/mandelbrot/cmd/internal/calculation"
	"github.com/eboyden42/mandelbrot/cmd/internal/colors"
	"github.com/eboyden42/mandelbrot/cmd/internal/images"
)

var size int
var numThreads int
var depth int
var palletConcavityPower float64
var pallet []color.RGBA

const percentUpdates = 100

type write struct {
	p     calculation.IntPoint
	value color.RGBA
}

func reader(points <-chan calculation.IntPoint, writes chan<- write, wg *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()

	for {
		ip, ok := <-points
		if !ok {
			return
		}

		fp := calculation.NewFloatPoint(ip)
		fp.Normalize(float64(size/2), 2.0)
		_, iterations := fp.IsInMandelbrot(depth, 4.0)

		wr := write{ip, pallet[iterations-1]}
		writes <- wr
	}
}

func writer(writes <-chan write, nums chan<- int, img images.Image, wg *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()

	for {
		wr, ok := <-writes
		if !ok {
			return
		}

		img.WritePoint(wr.p.X0, wr.p.Y0, wr.value)

		nums <- 1
	}
}

func tracker(total int, nums <-chan int, wg *sync.WaitGroup) {
	wg.Add(1)
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
			fmt.Printf("\rCalculating: %d %%", int(100.0*float64(partial)/float64(total)))
		}
	}
}

func main() {
	// arg parsing
	processors := runtime.NumCPU()
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

	// create pallet and image
	pallet = colors.NewPalletVar(depth, palletConcavityPower)
	newImage := images.NewImage(size)

	// create WaitGroups and channels
	var wgRead sync.WaitGroup
	var wgWrite sync.WaitGroup
	var wgTracker sync.WaitGroup
	chanSize := size * size
	points := make(chan calculation.IntPoint, chanSize)
	writes := make(chan write, chanSize)
	nums := make(chan int, chanSize)

	// create threads
	fmt.Printf("Creating %d threads\n", numThreads)

	for i := 0; i < numThreads-2; i++ {
		go reader(points, writes, &wgRead)
	}
	go writer(writes, nums, newImage, &wgWrite)
	go tracker(size*size, nums, &wgTracker)

	// starting timer and issuing points to pipeline
	start := time.Now()
	for x := newImage.StartWidth; x < newImage.EndWidth; x++ {
		for y := newImage.StartHeight; y < newImage.EndHeight; y++ {
			points <- calculation.IntPoint{X0: x, Y0: y}
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
	fmt.Printf("\nTotal time: %.5f \n", diff.Seconds())
	newImage.WriteImage(*filename)

}
