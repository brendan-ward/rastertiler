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
	case []uint16:
		typedValue := value.(uint16)
		for i := 0; i < len(typedBuffer); i++ {
			if typedBuffer[i] != typedValue {
				return false
			}
		}
		return true
	case []uint32:
		typedValue := value.(uint32)
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
	case []uint16:
		rightBuffer := right.([]uint16)
		if len(leftBuffer) != len(rightBuffer) {
			return false
		}
		for i := 0; i < len(leftBuffer); i++ {
			if leftBuffer[i] != rightBuffer[i] {
				return false
			}
		}
		return true
	case []uint32:
		rightBuffer := right.([]uint32)
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
	case []uint16:
		typedValue := value.(uint16)
		for i := 0; i < len(typedBuffer); i++ {
			typedBuffer[i] = typedValue
		}
	case []uint32:
		typedValue := value.(uint32)
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
	case []uint16:
		sourceBuffer := source.([]uint16)
		for row := rowOffset; row < rowOffset+sourceHeight; row++ {
			for col := colOffset; col < colOffset+sourceWidth; col++ {
				i = row*targetWidth + col
				srcIndex = (row-rowOffset)*sourceWidth + (col - colOffset)
				targetBuffer[i] = sourceBuffer[srcIndex]
			}
		}
	case []uint32:
		sourceBuffer := source.([]uint32)
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
