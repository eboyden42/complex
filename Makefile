all: main

main: imagewriter.go
	go build -o main imagewriter.go

clean:
	rm -f main mandelbrot.png

.PHONY: all