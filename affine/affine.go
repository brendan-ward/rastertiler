package affine

import (
	"fmt"
	"math"
)

// Affine data structure
// This is the same as Affine Python package used in rasterio
type Affine struct {
	A float64
	B float64
	C float64
	D float64
	E float64
	F float64
}

// Create an Affine transform from GDAL's representation
func FromGDAL(transform [6]float64) *Affine {
	return &Affine{
		A: transform[1],
		B: transform[2],
		C: transform[0],
		D: transform[4],
		E: transform[5],
		F: transform[3],
	}
}

// Convert Affine transform to GDAL's representation
func (a *Affine) ToGDAL() (transform [6]float64) {
	transform[0] = a.C
	transform[1] = a.A
	transform[2] = a.B
	transform[3] = a.F
	transform[4] = a.D
	transform[5] = a.E

	return transform
}

// Invert the Affine transform
func (a *Affine) Invert() *Affine {
	invDeterminant := 1 / (a.A*a.E - a.B*a.D)

	A := a.E * invDeterminant
	B := -a.B * invDeterminant
	D := -a.D * invDeterminant
	E := a.A * invDeterminant

	return &Affine{
		A: A,
		B: B,
		C: -a.C*A - a.F*B,
		D: D,
		E: E,
		F: -a.C*D - a.F*E,
	}
}

// Apply the transform to x and y using matrix multiplication
func (a *Affine) Multiply(x float64, y float64) (float64, float64) {
	return x*a.A + y*a.B + a.C, x*a.D + y*a.E + a.F
}

// Scale the Affine transform
func (a *Affine) Scale(x float64, y float64) *Affine {
	return &Affine{
		A: a.A * x,
		B: a.B,
		C: a.C,
		D: a.D,
		E: a.E * y,
		F: a.F,
	}
}

func (a *Affine) String() string {
	return fmt.Sprintf("Affine(%v, %v, %v,\n       %v, %v, %v)", a.A, a.B, a.C, a.D, a.E, a.F)
}

// Return the x, y resolution of the Affine transform
func (a *Affine) Resolution() (float64, float64) {
	return math.Abs(a.A), math.Abs(a.E)
}
