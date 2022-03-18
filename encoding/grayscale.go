package encoding

import (
	"image"
	"image/color"
)

type GrayscaleEncoder struct{}

func NewGrayscaleEncoder() *GrayscaleEncoder {
	return &GrayscaleEncoder{}
}

// Encode uint8 values to 8-bit grayscale PNG
func (e *GrayscaleEncoder) Encode(buffer []uint8, width int, height int, bits uint8) ([]byte, error) {
	img := image.NewGray(image.Rect(0, 0, width, height))

	var value uint8
	for row := 0; row < height; row++ {
		for col := 0; col < width; col++ {
			value = buffer[row*width+col]
			img.Set(col, row, color.Gray{value})
		}
	}

	return encodePNG(img)
}
