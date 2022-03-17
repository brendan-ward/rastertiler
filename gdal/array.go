package gdal

import (
	"fmt"
	"math"
	"strings"
)

type Array struct {
	DType  string
	Width  int
	Height int
	buffer interface{}
}

func (a *Array) String() string {
	padding := len(fmt.Sprint(a.Max())) + 1
	var arrayStr strings.Builder

	switch buffer := a.buffer.(type) {
	case []uint8:
		for row := 0; row < a.Height; row++ {
			if row > 0 {
				arrayStr.WriteString("\n")
			}
			for col := 0; col < a.Width; col++ {
				arrayStr.WriteString(fmt.Sprintf("%*v", padding, buffer[row*a.Width+col]))
			}
		}
	default:
		panic("String() not implemented for other dtypes")

	}

	return fmt.Sprintf("Array(%vx%v, dtype: %v)\n%v\n", a.Width, a.Height, a.DType, arrayStr.String())
}

// Calculate the maximum value, in the underlying data type
func (a *Array) Max() interface{} {
	maxValue := math.Inf(-1)
	size := a.Width * a.Height

	switch buffer := a.buffer.(type) {
	case []uint8:
		for i := 0; i < size; i++ {
			value := float64(buffer[i])
			if value > maxValue {
				maxValue = value
			}
		}
		return uint8(maxValue)
	default:
		panic("Max() not implemented yet for other dtypes")
	}
}

// Return true if all values equal the passed in value
func (a *Array) EqualsValue(value interface{}) bool {
	size := a.Height * a.Width
	switch array := a.buffer.(type) {
	case []uint8:
		typedValue := value.(uint8)
		for i := 0; i < size; i++ {
			if array[i] != typedValue {
				return false
			}
		}
		return true
	default:
		panic("other data types not yet supported for Equals()")
	}
}

// Return true if the two arrays have equal dimensions, dtypes, and values
func (left *Array) Equals(right *Array) bool {
	if left.Width != right.Width || left.Height != right.Height || left.DType != right.DType {
		return false
	}

	size := left.Height * left.Width
	switch leftbuffer := left.buffer.(type) {
	case []uint8:
		rightbuffer := right.buffer.([]uint8)
		for i := 0; i < size; i++ {
			if leftbuffer[i] != rightbuffer[i] {
				return false
			}
		}
		return true
	default:
		panic("other data types not yet supported for Equals()")
	}
}

// Get value at row, col position
func (a *Array) Get(row int, col int) interface{} {
	return a.buffer.([]interface{})[row*a.Width+col]
}

// Set value into row, col position
func (a *Array) Set(row int, col int, value interface{}) {
	switch buffer := a.buffer.(type) {
	case []uint8:
		buffer[row*a.Width+col] = value.(uint8)
	default:
		panic("Set() not implemented yet for other dtypes")
	}
}

//  Create a new array and fill with fill value, which must be of same type
// as dtype
func NewArray(width int, height int, dtype string, fillValue interface{}) *Array {
	size := width * height
	var buffer interface{}

	switch dtype {
	case "uint8":
		typedbuffer := make([]uint8, size)
		for i := 0; i < size; i++ {
			typedbuffer[i] = fillValue.(uint8)
		}
		buffer = typedbuffer
	default:
		panic("Other dtypes not yet supported for NewArray")
	}

	return &Array{
		DType:  dtype,
		Width:  width,
		Height: height,
		buffer: buffer,
	}
}

func (target *Array) Paste(source *Array, rowOffset int, colOffset int) error {
	if source.DType != target.DType {
		return fmt.Errorf("data types do not match")
	}

	// TODO:  make sure that source will fit into target
	if rowOffset < 0 || colOffset < 0 {
		return fmt.Errorf("offsets must be >= 0")
	}

	if rowOffset+source.Height > target.Height || rowOffset+source.Width > target.Width {
		return fmt.Errorf("size of array to paste is too big for target array, given offsets")
	}

	var i int
	var srcIndex int
	for row := rowOffset; row < rowOffset+source.Height; row++ {
		for col := colOffset; col < colOffset+source.Width; col++ {
			// TODO: verify indexing
			i = row*target.Width + col
			srcIndex = (row-rowOffset)*source.Width + (col - colOffset)
			switch target.DType {
			case "uint8":
				target.buffer.([]uint8)[i] = source.buffer.([]uint8)[srcIndex]
			default:
				panic("other dtypes not yet supported for Paste()")
			}
		}
	}

	return nil
}
