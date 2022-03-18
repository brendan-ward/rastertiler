package encoding

import (
	"image/color"
	"testing"
)

func TestNewColormap(t *testing.T) {
	colormap, err := NewColormap("1:#000000,3:#FFFFFF,4:#FF0000")
	if err != nil {
		t.Error(err)
	}
	expectedPalette := make([]color.Color, 4)
	expectedPalette[0] = color.NRGBA{0, 0, 0, 255}
	expectedPalette[1] = color.NRGBA{255, 255, 255, 255}
	expectedPalette[2] = color.NRGBA{255, 0, 0, 255}
	expectedPalette[3] = color.Transparent

	expectedIndexes := map[uint8]uint8{
		0: 3,
		1: 0,
		2: 3,
		3: 1,
		4: 2,
		5: 3,
	}

	palette := colormap.Palette()
	if len(palette) != len(expectedPalette) {
		t.Errorf("Palette is not expected size, got: %v", palette)
	}
	for i, expected := range expectedPalette {
		if palette[i] != expected {
			t.Errorf("palette %v: %v does not match expected value %v", i, palette[i], expected)
		}
	}

	for key, expected := range expectedIndexes {
		value := colormap.GetIndex(key)
		if value != expected {
			t.Errorf("value %v: index: %v does not match expected index %v", key, value, expected)
		}
	}
}
