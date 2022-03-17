package affine

import (
	"math"
	"testing"
)

func closeEnough(a, b, tolerance float64) bool {
	return math.Abs(a-b) < tolerance
}

func TestFromGDAL(t *testing.T) {
	transform := [6]float64{
		1000, 30, 0, 2000, 0, -30,
	}
	expected := Affine{
		A: 30,
		B: 0,
		C: 1000,
		D: 0,
		E: -30,
		F: 2000,
	}

	a := FromGDAL(transform)
	if *a != expected {
		t.Errorf("%v did not match expected: %v", a, expected)
	}
}

func TestToGDAL(t *testing.T) {
	transform := [6]float64{
		1000, 30, 0, 2000, 0, -30,
	}
	out := FromGDAL(transform).ToGDAL()
	for i := 0; i < 6; i++ {
		if !closeEnough(transform[i], out[i], 1e-6) {
			t.Errorf("%v: %v did not match expected value: %v", i, out[i], transform[i])
		}
	}
}

func TestInvert(t *testing.T) {
	transform := [6]float64{
		1000, 30, 0, 2000, 0, -30,
	}
	expected := Affine{
		A: 0.03333333333333333,
		B: 0.0,
		C: -33.333333333333336,
		D: 0.0,
		E: -0.03333333333333333,
		F: 66.66666666666667,
	}

	a := FromGDAL(transform).Invert()

	if *a != expected {
		t.Errorf("%v did not match expected: %v", a, expected)
	}
}

func TestMultiply(t *testing.T) {
	transform := [6]float64{
		1000, 30, 0, 2000, 0, -30,
	}
	expectedX, expectedY := 1060.0, 1910.0

	x, y := FromGDAL(transform).Multiply(2, 3)

	if x != expectedX || y != expectedY {
		t.Errorf("(%v, %v) is not expected value: (%v, %v)", x, y, expectedX, expectedY)
	}
}

func TestScale(t *testing.T) {
	transform := [6]float64{
		1000, 30, 0, 2000, 0, -30,
	}
	expected := Affine{
		A: 60,
		B: 0,
		C: 1000,
		D: 0,
		E: -90,
		F: 2000,
	}

	a := FromGDAL(transform).Scale(2, 3)
	if *a != expected {
		t.Errorf("%v did not match expected: %v", a, expected)
	}
}
