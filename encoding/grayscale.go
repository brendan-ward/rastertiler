package encoding

import (
	"bytes"
	"image"
	"image/png"
)

type GrayscaleEncoder struct {
	img       *image.Gray
	pngBuffer bytes.Buffer
	width     int
	height    int
}

func NewGrayscaleEncoder(width int, height int) *GrayscaleEncoder {
	return &GrayscaleEncoder{
		img:    image.NewGray(image.Rect(0, 0, width, height)),
		width:  width,
		height: height,
	}
}

// Encode uint8 values to 8-bit grayscale PNG
func (e *GrayscaleEncoder) Encode(buffer interface{}) ([]byte, error) {
	switch typedBuffer := buffer.(type) {
	case []uint8:
		for row := 0; row < e.height; row++ {
			for col := 0; col < e.width; col++ {
				e.img.Pix[row*e.width+col] = typedBuffer[row*e.width+col]
			}
		}
	}

	e.pngBuffer.Reset()
	err := png.Encode(&e.pngBuffer, e.img)
	if err != nil {
		return nil, err
	}
	return e.pngBuffer.Bytes(), nil
}
