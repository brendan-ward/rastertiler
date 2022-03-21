package array

import "fmt"

func AllEquals(buffer interface{}, value interface{}) bool {
	switch typedBuffer := buffer.(type) {
	case []uint8:
		typedValue := value.(uint8)
		for i := 0; i < len(typedBuffer); i++ {
			if typedBuffer[i] != typedValue {
				return false
			}
		}
		return true
	default:
		panic("other data types not yet supported for AllEquals()")
	}
}

// Return true if the two arrays have equal values
// will fail if arrays do not have same type
func Equals(left interface{}, right interface{}) bool {
	// if left.Width != right.Width || left.Height != right.Height || left.DType != right.DType {
	// 	return false
	// }

	// size := left.Height * left.Width
	switch leftBuffer := left.(type) {
	case []uint8:
		rightBuffer := right.([]uint8)
		if len(leftBuffer) != len(rightBuffer) {
			return false
		}
		for i := 0; i < len(leftBuffer); i++ {
			if leftBuffer[i] != rightBuffer[i] {
				return false
			}
		}
		return true
	default:
		panic("other data types not yet supported for Equals()")
	}
}

func Fill(buffer interface{}, value interface{}) {
	switch typedBuffer := buffer.(type) {
	case []uint8:
		typedValue := value.(uint8)
		for i := 0; i < len(typedBuffer); i++ {
			typedBuffer[i] = typedValue
		}
	default:
		panic("other data types not yet supported for AllEquals()")
	}
}

// Paste values from source 2D array into target 2D array
// Source must be no greater than target and must be of same dtype
func Paste(target interface{}, targetHeight int, targetWidth int, source interface{}, sourceHeight int, sourceWidth int, rowOffset int, colOffset int) error {
	if rowOffset < 0 || colOffset < 0 {
		return fmt.Errorf("offsets must be >= 0")
	}

	if rowOffset+sourceHeight > targetHeight || colOffset+sourceWidth > targetWidth {
		return fmt.Errorf("size of array to paste is too big for target array, given offsets")
	}

	var i int
	var srcIndex int

	switch targetBuffer := target.(type) {
	case []uint8:
		sourceBuffer := source.([]uint8)
		for row := rowOffset; row < rowOffset+sourceHeight; row++ {
			for col := colOffset; col < colOffset+sourceWidth; col++ {
				i = row*targetWidth + col
				srcIndex = (row-rowOffset)*sourceWidth + (col - colOffset)
				targetBuffer[i] = sourceBuffer[srcIndex]
			}
		}
	default:
		panic("other dtypes not yet supported for Paste()")
	}

	return nil
}

func AsUint8(buffer interface{}) []uint8 {
	switch typedBuffer := buffer.(type) {
	case []uint8:
		return typedBuffer
	default:
		panic("other dtypes not yet supported for AsUint8()")
	}
}
