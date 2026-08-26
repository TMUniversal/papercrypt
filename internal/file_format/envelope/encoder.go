package envelope

import "github.com/dasio/base45"

// ContentEncoder encodes and decodes content for the envelope.
// Implementations must produce deterministic output for a given input.
type ContentEncoder interface {
	// EncodeToString encodes the given bytes into a string.
	EncodeToString(data []byte) string
	// DecodeString decodes the given string back to bytes.
	DecodeString(data string) ([]byte, error)
	// EncodedCRCSize returns the number of characters produced
	// by encoding a 4-byte CRC-32 value. For base45 this is 6.
	EncodedCRCSize() int
}

// Base45Encoder implements ContentEncoder using base45 encoding.
type Base45Encoder struct{}

// EncodeToString encodes bytes using base45.
func (Base45Encoder) EncodeToString(data []byte) string {
	return base45.EncodeToString(data)
}

// DecodeString decodes a base45-encoded string.
func (Base45Encoder) DecodeString(data string) ([]byte, error) {
	return base45.DecodeString(data)
}

// EncodedCRCSize returns 6, the number of base45 characters for 4 bytes.
func (Base45Encoder) EncodedCRCSize() int {
	return 6
}
