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

func TestTransform(t *testing.T) {
	transform := &affine.Affine{
		A: 30,
		B: 0,
		C: 1000,
		D: 0,
		E: -30,
		F: 2000,
	}

	tests := []struct {
		window   *Window
		expected *affine.Affine
	}{
		{
			window:   &Window{XOffset: 0, YOffset: 0, Width: 10, Height: 20},
			expected: &affine.Affine{30, 0, 1000, 0, -30, 2000},
		},
		{
			window:   &Window{XOffset: 10, YOffset: 20, Width: 10, Height: 20},
			expected: &affine.Affine{30, 0, 1300, 0, -30, 1400},
		},
		{
			window:   &Window{XOffset: -10, YOffset: -20, Width: 10, Height: 20},
			expected: &affine.Affine{30, 0, 700, 0, -30, 2600},
		},
	}

	for _, tc := range tests {
		windowTransform := WindowTransform(tc.window, transform)
		if *windowTransform != *tc.expected {
			t.Errorf("%v\nnot expected value: %v", windowTransform, tc.expected)
		}
	}

}
