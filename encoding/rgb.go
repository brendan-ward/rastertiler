package encoding

import (
	"bytes"
	"image"
	"image/png"
)

type RGBEncoder struct {
	img       *image.NRGBA // TODO:
	pngBuffer bytes.Buffer
	width     int
	height    int
}

func NewRGBEncoder(width int, height int) *RGBEncoder {
	return &RGBEncoder{
		img:    image.NewNRGBA(image.Rect(0, 0, width, height)),
		width:  width,
		height: height,
	}
}

// Encode uint8...uint32 values to 24-bit RGB PNG
func (e *RGBEncoder) Encode(buffer interface{}) ([]byte, error) {
	switch typedBuffer := buffer.(type) {
	case []uint32:
		var value uint32
		for row := 0; row < e.height; row++ {
			for col := 0; col < e.width; col++ {
				value = typedBuffer[row*e.width+col]
				i := e.img.PixOffset(col, row)
				e.img.Pix[i] = uint8(value>>16) & 255  // R
				e.img.Pix[i+1] = uint8(value>>8) & 255 // G
				e.img.Pix[i+2] = uint8(value) & 255    // B
				e.img.Pix[i+3] = 255                   // A, hardcoded for RGB
			}
		}
	default:
		panic("Other dtypes not yet supported for RGBEncoder::Encode()")
	}

	e.pngBuffer.Reset()
	err := png.Encode(&e.pngBuffer, e.img)
	if err != nil {
		return nil, err
	}
	return e.pngBuffer.Bytes(), nil
}
