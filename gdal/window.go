package gdal

import (
	"fmt"
	"math"

	"github.com/brendan-ward/rastertiler/affine"
)

type Window struct {
	XOffset float64
	YOffset float64
	Width   float64
	Height  float64
}

func (w *Window) String() string {
	return fmt.Sprintf("Window(xoff: %v, yoff: %v, width: %v, height: %v)", w.XOffset, w.YOffset, w.Width, w.Height)
}

// Calculate Window based on Affine transform and bounds
func WindowFromBounds(transform *affine.Affine, bounds *affine.Bounds) *Window {
	invTransform := transform.Invert()

	// calculate outer bounds
	xmin := math.Inf(1)
	ymin := math.Inf(1)
	xmax := math.Inf(-1)
	ymax := math.Inf(-1)

	xs := [4]float64{}
	ys := [4]float64{}
	var x float64
	var y float64
	x, y = invTransform.Multiply(bounds.Xmin, bounds.Ymin)
	xs[0] = x
	ys[0] = y

	x, y = invTransform.Multiply(bounds.Xmin, bounds.Ymax)
	xs[1] = x
	ys[1] = y

	x, y = invTransform.Multiply(bounds.Xmax, bounds.Ymin)
	xs[2] = x
	ys[2] = y

	x, y = invTransform.Multiply(bounds.Xmax, bounds.Ymax)
	xs[3] = x
	ys[3] = y

	for i := 0; i < 4; i++ {
		if xs[i] < xmin {
			xmin = xs[i]
		}
		if ys[i] < ymin {
			ymin = ys[i]
		}
		if xs[i] > xmax {
			xmax = xs[i]
		}
		if ys[i] > ymax {
			ymax = ys[i]
		}
	}

	return &Window{
		XOffset: xmin,
		YOffset: ymin,
		Width:   xmax - xmin,
		Height:  ymax - ymin,
	}
}

// Calculate the Affine transform representing the relative window
// within the passed-in Affine transform
func WindowTransform(window *Window, transform *affine.Affine) *affine.Affine {
	x, y := transform.Multiply(window.XOffset, window.YOffset)

	return &affine.Affine{
		A: transform.A,
		B: transform.B,
		C: x,
		D: transform.D,
		E: transform.E,
		F: y,
	}
}
