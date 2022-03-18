package encoding

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
)

// Encode uint8 values to 8-bit grayscale PNG
func EncodeGrayPNG(buffer []uint8, width int, height int) ([]byte, error) {
	img := image.NewGray(image.Rect(0, 0, width, height))

	var value uint8
	for row := 0; row < height; row++ {
		for col := 0; col < width; col++ {
			value = buffer[row*width+col]
			// value = data.Get(row, col).(uint8)
			img.Set(col, row, color.Gray{value})
		}
	}

	return encodePNG(img)
}

func EncodePalettedPNG(buffer []uint8, width int, height int, colormap *Colormap) ([]byte, error) {
	img := image.NewPaletted(image.Rect(0, 0, width, height), colormap.Palette())

	var value uint8
	for row := 0; row < height; row++ {
		for col := 0; col < width; col++ {
			// value = data.Get(row, col).(uint8)
			value = buffer[row*width+col]
			img.SetColorIndex(col, row, colormap.GetIndex(value))
		}
	}

	return encodePNG(img)
}

func encodePNG(img image.Image) ([]byte, error) {
	var buffer bytes.Buffer
	err := png.Encode(&buffer, img)
	if err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}
