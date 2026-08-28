package envelope

import "github.com/dasio/base45"

// EncodingType identifies the encoding applied to an envelope's content.
type EncodingType uint8

const (
	// EncodingTypeRaw marks content encoded without a transform.
	EncodingTypeRaw EncodingType = iota
	// EncodingTypeBase45 marks content encoded using base45.
	EncodingTypeBase45
)

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
	// EncodingType returns the encoding used by this encoder,
	// stored in the envelope header.
	EncodingType() EncodingType
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

// EncodingType returns EncodingTypeBase45.
func (Base45Encoder) EncodingType() EncodingType {
	return EncodingTypeBase45
}
