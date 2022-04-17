package encoding

// PNGEncoder provides an Encode() function for encoding buffer to PNG
type PNGEncoder interface {
	Encode(buffer interface{}) ([]byte, error)
}
