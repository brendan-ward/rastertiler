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

// Calculate Window based on transform and bounds
func WindowFromBounds(transform *affine.Affine, bounds [4]float64) *Window {
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
	x, y = invTransform.Multiply(bounds[0], bounds[1])
	xs[0] = x
	ys[0] = y

	x, y = invTransform.Multiply(bounds[0], bounds[3])
	xs[1] = x
	ys[1] = y

	x, y = invTransform.Multiply(bounds[2], bounds[1])
	xs[2] = x
	ys[2] = y

	x, y = invTransform.Multiply(bounds[2], bounds[3])
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

	fmt.Printf("xmin: %v, ymin: %v, xmax: %v, ymax: %v\n", xmin, ymin, xmax, ymax)

	return &Window{
		XOffset: xmin,
		YOffset: ymin,
		Width:   xmax - xmin,
		Height:  ymax - ymin,
	}
}

func (w *Window) String() string {
	return fmt.Sprintf("Window(xoff: %v, yoff: %v, width: %v, height: %v)", w.XOffset, w.YOffset, w.Width, w.Height)
}
