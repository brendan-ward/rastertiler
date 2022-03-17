package gdal

import (
	"testing"

	"github.com/brendan-ward/rastertiler/affine"
)

func TestWindowFromBounds(t *testing.T) {
	transform := &affine.Affine{
		A: 30,
		B: 0,
		C: 1000,
		D: 0,
		E: -30,
		F: 2000,
	}

	bounds := [4]float64{
		0, 10, 100, 200,
	}

	expected := Window{
		XOffset: -33.333333333333336,
		YOffset: 60.00000000000001,
		Width:   3.333333333333332,
		Height:  6.333333333333336,
	}

	window := WindowFromBounds(transform, bounds)

	if *window != expected {
		t.Errorf("%v not expected value: %v", window, expected)
	}
}
