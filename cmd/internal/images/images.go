package images

import (
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"os"
)

type Image struct {
	StartWidth, EndWidth, StartHeight, EndHeight int
	Img                                          *image.RGBA
}

func NewImage(size int) Image {
	// calculate dimensions
	startWidth := size/2 - size
	endWidth := size / 2
	startHeight := startWidth
	endHeight := endWidth

	// create rectangle and image
	rect := image.Rect(startWidth, startHeight, endWidth, endHeight)
	return Image{startWidth, endWidth, startHeight, endHeight, image.NewRGBA(rect)}
}

func (i *Image) WritePoint(x, y int, color color.RGBA) {
	i.Img.Set(x, y, color)
}

func (i *Image) WriteImageAsPNG(filename string) {
	// writing output image
	file, err := os.Create(filename)
	if err != nil {
		panic(err)
	}

	err = png.Encode(file, i.Img)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Image written to %s\n", filename)
}

func (i *Image) WriteImageAsJPEG(filename string) {
	file, err := os.Create(filename)
	if err != nil {
		panic(err)
	}

	err = jpeg.Encode(file, i.Img, nil)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Image written to %s\n", filename)
}
