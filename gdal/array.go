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

//  Create a new array and fill with fill value, which must be of same type
// as dtype
func NewArray(width int, height int, dtype string, fillValue interface{}) *Array {
	size := width * height
	var buffer interface{}

	switch dtype {
	case "uint8":
		typedbuffer := make([]uint8, size)
		fill := fillValue.(uint8)
		for i := 0; i < size; i++ {
			typedbuffer[i] = fill
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
	switch buffer := a.buffer.(type) {
	case []uint8:
		return buffer[row*a.Width+col]
	default:
		panic("Set() not implemented yet for other dtypes")
	}
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

func (a *Array) Uint8Buffer() (buffer []uint8, bits uint8, err error) {
	switch typedBuffer := a.buffer.(type) {
	case []uint8:
		bits = 8
		buffer = typedBuffer
	default:
		panic("Uint8Buffer() not yet supported for other data types")
	}
	return
}

func (target *Array) Paste(source *Array, rowOffset int, colOffset int) error {
	if source.DType != target.DType {
		return fmt.Errorf("data types do not match")
	}

	if rowOffset < 0 || colOffset < 0 {
		return fmt.Errorf("offsets must be >= 0")
	}

	if rowOffset+source.Height > target.Height || colOffset+source.Width > target.Width {
		return fmt.Errorf("size of array to paste is too big for target array, given offsets")
	}

	var i int
	var srcIndex int

	switch targetBuffer := target.buffer.(type) {
	case []uint8:
		sourceBuffer := source.buffer.([]uint8)
		for row := rowOffset; row < rowOffset+source.Height; row++ {
			for col := colOffset; col < colOffset+source.Width; col++ {
				i = row*target.Width + col
				srcIndex = (row-rowOffset)*source.Width + (col - colOffset)
				targetBuffer[i] = sourceBuffer[srcIndex]
			}
		}
	default:
		panic("other dtypes not yet supported for Paste()")
	}

	return nil
}

// Count number of pixels per value and return as a map[count]value
func (a *Array) Histogram() map[int]int {
	counts := make(map[int]int)

	switch buffer := a.buffer.(type) {
	case []uint8:
		var prior int
		var ok bool
		var value int

		for i := 0; i < a.Width*a.Height; i++ {
			value = int(buffer[i])
			if prior, ok = counts[value]; !ok {
				prior = 0
			}
			counts[value] = prior + 1
		}

	default:
		panic("UniqueCounts() not implemented yet for other dtypes")
	}

	return counts
}
