package encoding

import (
	"bytes"
	"image"
	"image/png"
)

type GrayscaleEncoder struct {
	img    *image.Gray
	buffer bytes.Buffer
}

func NewGrayscaleEncoder(width int, height int) *GrayscaleEncoder {
	return &GrayscaleEncoder{
		img: image.NewGray(image.Rect(0, 0, width, height)),
	}
}

// Encode uint8 values to 8-bit grayscale PNG
func (e *GrayscaleEncoder) Encode(buffer []uint8, width int, height int, bits uint8) ([]byte, error) {
	var value uint8
	for row := 0; row < height; row++ {
		for col := 0; col < width; col++ {
			value = buffer[row*width+col]
			e.img.Pix[row*width+col] = value
		}
	}

	e.buffer.Reset()
	err := png.Encode(&e.buffer, e.img)
	if err != nil {
		return nil, err
	}
	return e.buffer.Bytes(), nil
}
