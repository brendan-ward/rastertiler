package encoding

import (
	"bytes"
	"image"
	"image/png"
)

// PNGEncoder provides an Encode() function for encoding uint8 data to PNG
type PNGEncoder interface {
	// TODO: reusable buffer for encoding

	// Encode uint8 data to PNG based on width and height
	// bits can be 8 (grayscale / paletted), 16 (grayscale), 24 (RGB), 32 (RGBA)
	// There should be (bits / 8) number of uint8 values per pixel (interleaved)
	Encode(buffer []uint8, width int, height int, bits uint8) ([]byte, error)
}

// Encode the Image to PNG bytes
func encodePNG(img image.Image) ([]byte, error) {
	var buffer bytes.Buffer
	err := png.Encode(&buffer, img)
	if err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}
