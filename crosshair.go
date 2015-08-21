package main

import (
	"fmt"
	"github.com/veandco/go-sdl2/sdl"
	"os"
	"reflect"
    "runtime"
	"unsafe"
)

type Pixel uint32

// Creates a slice of Pixel from a raw pointer
func sliceFromPixelPtr(data unsafe.Pointer, length int) []Pixel {
	var pixels []Pixel
	sliceHeader := (*reflect.SliceHeader)(unsafe.Pointer(&pixels))
	sliceHeader.Cap = int(length)
	sliceHeader.Len = int(length)
	sliceHeader.Data = uintptr(data)
	return pixels
}

func lockTexture(tex *sdl.Texture, width int, height int) (pixels []Pixel, pitch int) {
	var data unsafe.Pointer
	err := tex.Lock(nil, &data, &pitch)
	if err != nil {
		fmt.Fprintf(os.Stderr, "tex.Lock: %v", err)
		panic(err)
	}
	// Convert pitch units from byte to pixels
	pitch /= 4
	pixels = sliceFromPixelPtr(data, width*height)
	return
}

// Clip x to half-open interval [a,b)
func clip(x int, a int, b int) int {
	if x < a {
		return a
	}
	if x >= b {
		return b
	}
	return x
}

var escape *string
var counter int

// Generate some heap-allocated stuff
func generateGarbage() {
	escape = new(string)
	*escape = fmt.Sprintf("%v", counter)
}

// Draw cross-hairs through (x0,y0) onto the given pixel buffer.
func drawCross(pixels []Pixel, pitch int, width int, height int, x0 int, y0 int) {
	x0 = clip(x0, 0, width)
	y0 = clip(y0, 0, height)
	for y := 0; y < height; y++ {
		generateGarbage()
		for x := 0; x < width; x++ {
			var color Pixel
			if y == y0 || x == x0 {
				color = 0xFFFFFFFF
			} else {
				color = 0xFF000000
			}
			pixels[y*pitch+x] = color
		}
	}
}

var winTitle string = "Crosshair"
var winWidth, winHeight int = 800, 600

func run() int {
    // See discussion at https://github.com/golang/go/wiki/LockOSThread
    runtime.LockOSThread()
    defer runtime.UnlockOSThread()

	// Create window
	window, err := sdl.CreateWindow(winTitle, sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED,
		winWidth, winHeight, sdl.WINDOW_SHOWN)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create window: %v\n", err)
		panic(err)
	}
	defer window.Destroy()

	// Create renderer
	width, height := window.GetSize()
	renderer, err := sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED|sdl.RENDERER_PRESENTVSYNC)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create renderer: %v\n", err)
		panic(err)
	}
	defer renderer.Destroy()

	// Create texture
	tex, err := renderer.CreateTexture(sdl.PIXELFORMAT_ARGB8888, sdl.TEXTUREACCESS_STREAMING, width, height)
	if err != nil {
		fmt.Fprintf(os.Stderr, "renderer.CreateTexture: %v\n", err)
		panic(err)
	}
	defer tex.Destroy()

	// Show crosshair animation.  Quit when a key comes up.
	var mouseX, mouseY int
	for {
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch t := event.(type) {
			case *sdl.QuitEvent:
				return 0
			case *sdl.MouseMotionEvent:
				mouseX = int(t.X)
				mouseY = int(t.Y)
			case *sdl.KeyUpEvent:
				return 0
			}
		}

		pixels, pitch := lockTexture(tex, width, height)
		drawCross(pixels, pitch, width, height, mouseX, mouseY)
		tex.Unlock()

		err := renderer.Clear()
		if err != nil {
			fmt.Fprintf(os.Stderr, "renderer.Clear: %v", err)
			panic(err)
		}
		renderer.Copy(tex, nil, nil)
		if err != nil {
			fmt.Fprintf(os.Stderr, "renderer.Copy: %v", err)
			panic(err)
		}
		renderer.Present()
	}

	return 0
}

func main() {
	run()
}
