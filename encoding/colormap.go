package encoding

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"strconv"
	"strings"
)

type Colormap struct {
	values  map[uint8]uint8 // map of value to index in palette
	palette color.Palette
}

// Returns palette index of value
// any values not in original colormap are set to transparent
func (c *Colormap) GetIndex(value uint8) uint8 {
	if index, ok := c.values[value]; ok {
		return index
	}
	return uint8(len(c.palette) - 1)
}

func (c *Colormap) Palette() color.Palette {
	return c.palette
}

// Create new colormap by parsing colormap string, which is a comma-delimited
// set of <value>:<hex> entries, e.g., "1:#AABBCC,2:#DDEEFF"
func NewColormap(colormap string) (*Colormap, error) {
	entries := strings.Split(strings.ReplaceAll(colormap, " ", ""), ",")

	palette := make([]color.Color, len(entries)+1)
	values := make(map[uint8]uint8, len(entries))
	for i, entry := range entries {
		parts := strings.Split(entry, ":")
		value, err := strconv.ParseUint(parts[0], 10, 8)
		if err != nil {
			return nil, err
		}
		values[uint8(value)] = uint8(i)

		color, err := parseHex(parts[1])
		if err != nil {
			return nil, err
		}

		palette[i] = color
	}
	palette[len(entries)] = color.Transparent

	return &Colormap{
		values:  values,
		palette: palette,
	}, nil
}

// from: https://stackoverflow.com/a/54200713/2740575
func parseHex(hex string) (c color.NRGBA, err error) {
	c.A = 0xff

	if hex[0] != '#' {
		return c, fmt.Errorf("Invalid hex color format")
	}

	hexToByte := func(b byte) byte {
		switch {
		case b >= '0' && b <= '9':
			return b - '0'
		case b >= 'a' && b <= 'f':
			return b - 'a' + 10
		case b >= 'A' && b <= 'F':
			return b - 'A' + 10
		}
		err = fmt.Errorf("Invalid hex color format")
		return 0
	}

	switch len(hex) {
	case 7:
		c.R = hexToByte(hex[1])<<4 + hexToByte(hex[2])
		c.G = hexToByte(hex[3])<<4 + hexToByte(hex[4])
		c.B = hexToByte(hex[5])<<4 + hexToByte(hex[6])
	case 4:
		c.R = hexToByte(hex[1]) * 17
		c.G = hexToByte(hex[2]) * 17
		c.B = hexToByte(hex[3]) * 17
	default:
		err = fmt.Errorf("Invalid hex color format")
	}
	return c, err
}

type ColormapEncoder struct {
	colormap  *Colormap
	img       *image.Paletted
	pngBuffer bytes.Buffer
	width     int
	height    int
}

func NewColormapEncoder(width int, height int, colormap *Colormap) *ColormapEncoder {
	return &ColormapEncoder{
		colormap: colormap,
		img:      image.NewPaletted(image.Rect(0, 0, width, height), colormap.Palette()),
		width:    width,
		height:   height,
	}
}

func (e *ColormapEncoder) Encode(buffer interface{}) ([]byte, error) {
	switch typedBuffer := buffer.(type) {
	case []uint8:
		var value uint8
		for row := 0; row < e.height; row++ {
			for col := 0; col < e.width; col++ {
				value = typedBuffer[row*e.width+col]
				e.img.SetColorIndex(col, row, e.colormap.GetIndex(value))
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
