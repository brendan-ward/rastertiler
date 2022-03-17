package gdal

import (
	"testing"
)

func TestNewArray(t *testing.T) {
	width := 4
	height := 8
	dtype := "uint8"
	var fill uint8 = 2

	array := NewArray(width, height, dtype, fill)

	if array.Width != width || array.Height != height {
		t.Errorf("Array dimensions (%v, %v) do not match expected: %v, %v", array.Width, array.Height, width, height)
	}
	if array.DType != dtype {
		t.Errorf("Array dtype %s does not match expected: %v", array.DType, dtype)
	}

	for i := 0; i < width*height; i++ {
		if array.buffer.([]uint8)[i] != fill {
			t.Errorf("array does not match expected fill value: %v", fill)
		}
	}
}

func TestPaste(t *testing.T) {
	var fill uint8 = 2
	target := NewArray(10, 10, "uint8", uint8(0))
	source := NewArray(2, 3, "uint8", fill)
	source.Set(1, 1, uint8(3))

	expected := NewArray(10, 10, "uint8", uint8(0))

	expected.Set(1, 1, fill)
	expected.Set(1, 2, fill)
	expected.Set(2, 1, fill)
	expected.Set(2, 2, uint8(3))
	expected.Set(3, 1, fill)
	expected.Set(3, 2, fill)

	target.Paste(source, 1, 1)
	if !target.Equals(expected) {
		t.Errorf("data:\n%v\ndoes not match expected:\n%v", target, expected)
	}
}
