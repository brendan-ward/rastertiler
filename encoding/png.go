package encoding

// PNGEncoder provides an Encode() function for encoding uint8 data to PNG
type PNGEncoder interface {
	// Encode uint8 data to PNG based on width and height
	// bits can be 8 (grayscale / paletted), 16 (grayscale), 24 (RGB), 32 (RGBA)
	// There should be (bits / 8) number of uint8 values per pixel (interleaved)
	Encode(data []uint8, bits uint8) ([]byte, error)
}
