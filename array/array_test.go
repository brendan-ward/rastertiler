package array

import (
	"testing"
)

func TestAllEquals(t *testing.T) {
	size := 4
	var fill uint8 = 0
	array := make([]uint8, size)
	for i := 0; i < size; i++ {
		array[i] = fill
	}

	if !AllEquals(array, fill) {
		t.Errorf("AllEquals() returned false when should have returned true")
	}

	var value uint8 = 2
	if AllEquals(array, value) {
		t.Errorf("AllEquals() returned true when should have returned false")
	}
}

func TestEquals(t *testing.T) {
	var i uint8
	var size uint8 = 4
	var fill uint8 = 0
	left := make([]uint8, int(size))
	for i = 0; i < size; i++ {
		left[i] = fill
	}

	right := make([]uint8, int(size))
	for i = 0; i < size; i++ {
		right[i] = i
	}

	if !Equals(left, left) {
		t.Errorf("Equals() returned false when should have returned true")
	}

	if Equals(left, right) {
		t.Errorf("Equals() returned true when should have returned false")
	}
}

func TestFill(t *testing.T) {
	var fill uint8 = 2
	target := make([]uint8, 8)
	Fill(target, fill)

	if !AllEquals(target, fill) {
		t.Errorf("Fill() did not set all values as expected")
	}

	fill = 4
	Fill(target, fill)
	if !AllEquals(target, fill) {
		t.Errorf("Fill() did not set all values as expected")
	}
}

func TestPaste(t *testing.T) {
	var fill uint8 = 0
	var value uint8 = 3

	// 10x10 array
	targetWidth := 10
	targetHeight := 10
	target := make([]uint8, targetWidth*targetHeight)

	Fill(target, fill)

	// 3x2 array
	sourceWidth := 2
	sourceHeight := 3
	source := make([]uint8, sourceWidth*sourceHeight)

	var sourceFill uint8 = 2
	Fill(source, sourceFill)
	// set cell 1,1 in source
	source[sourceWidth+1] = value

	expected := make([]uint8, targetWidth*targetHeight)
	Fill(expected, fill)

	expected[targetWidth*1+1] = sourceFill
	expected[targetWidth*1+2] = sourceFill
	expected[targetWidth*2+1] = sourceFill
	expected[targetWidth*2+2] = value
	expected[targetWidth*3+1] = sourceFill
	expected[targetWidth*3+2] = sourceFill

	Paste(target, targetHeight, targetWidth, source, sourceHeight, sourceWidth, 1, 1)
	if !Equals(target, expected) {
		t.Errorf("data:\n%v\ndoes not match expected:\n%v", target, expected)
	}
}
