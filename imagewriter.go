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

type point struct {
	x, y int
}

type write struct {
	p     point
	value bool
}

func bitMapToNormFloat(num, end int) float64 {
	return (1 / float64(end)) * float64(num) * 2
}

func isInMandlebrot(c complex128, n int, threshold float64) bool {
	// iterate the function f(z) = z^2 + c n times tracking the previous value
	// if the absolute difference between the previous and current value ever exceeds threshold return false, else true
	prev := c
	curr := c
	for i := 0; i < n; i++ {
		if math.Abs(modulus(curr)-modulus(prev)) > threshold {
			return false
		}
		tmp := curr
		curr = prev*prev + c
		prev = tmp
	}
	return true
}

func modulus(c complex128) float64 {
	r := real(c)
	i := imag(c)
	return math.Sqrt(r*r + i*i)
}

func reader(id, width int, points <-chan point, writes chan<- write, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		p, ok := <-points
		// fmt.Printf("worker %d recieved %d, %d", id, p.x, p.y)
		if !ok {
			fmt.Printf("Read channel closed %d\n", id)
			return
		}
		xval := bitMapToNormFloat(p.x, width)
		yval := bitMapToNormFloat(p.y, width)
		z := complex(xval, yval)

		wr := write{p, isInMandlebrot(z, 100, 30.0)}
		writes <- wr
	}
}

func writer(width int, writes <-chan write, nums chan<- int, img *image.RGBA, wg *sync.WaitGroup) {
	defer wg.Done()
	black := color.RGBA{0, 0, 0, 0}
	white := color.RGBA{255, 255, 255, 255}

	for {
		wr, ok := <-writes
		if !ok {
			fmt.Println("Write channel closed")
			return
		}
		if wr.value {
			img.Set(wr.p.x, wr.p.y, white)
		} else {
			img.Set(wr.p.x, wr.p.y, black)
		}
		nums <- 1
	}
}

func tracker(total int, nums <-chan int, wg *sync.WaitGroup) {
	defer wg.Done()
	const percentUpdates = 4

	interval := total / percentUpdates
	partial := 0

	for {
		_, ok := <-nums
		if !ok {
			fmt.Println("Tracker channel closed")
			return
		}
		partial++
		if partial%interval == 0 {
			fmt.Printf("%d %% \n", int(100.0*float64(partial)/float64(total)))
		}
	}
}

func main() {
	const width = 2000
	const height = 2000
	const numThreads = 4

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
		go reader(i, endWidth, points, writes, &wgRead)
	}

	wgWrite.Add(1)
	go writer(endWidth, writes, nums, img, &wgWrite)

	wgTracker.Add(1)
	go tracker(width*height, nums, &wgTracker)

	start := time.Now()
	fmt.Println("Started calculating...")

	for x := startWidth; x < endWidth; x++ {
		for y := startHeight; y < endHeight; y++ {
			points <- point{x, y}
		}
	}

	close(points)
	wgRead.Wait()
	end1 := time.Now()
	diff1 := end1.Sub(start)
	fmt.Printf("Total time for reader execution: %.5f \n", diff1.Seconds())

	close(writes)
	wgWrite.Wait()
	end := time.Now()
	diff := end.Sub(start)
	fmt.Printf("Total time for reader and writer execution: %.5f \n", diff.Seconds())

	close(nums)
	wgTracker.Wait()

	file, err := os.Create("mandlebrot.png")
	if err != nil {
		panic(err)
	}

	err = png.Encode(file, img)
	if err != nil {
		panic(err)
	}

	fmt.Println("Image written to mandlebrot.png")
}
