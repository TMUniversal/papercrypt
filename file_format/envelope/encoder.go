package envelope

import (
	"fmt"

	"github.com/dasio/base45"
)

type EncodingType uint8

const (
	EncodingTypeRaw EncodingType = iota
	EncodingTypeBase45
)

// ContentEncoder implementations must produce deterministic output for a
// given input.
type ContentEncoder interface {
	EncodeToString(data []byte) string
	DecodeString(data string) ([]byte, error)
	EncodedCRCSize() int
	EncodingType() EncodingType
}

type Base45Encoder struct{}

func (Base45Encoder) EncodeToString(data []byte) string {
	return base45.EncodeToString(data)
}

func (Base45Encoder) DecodeString(data string) ([]byte, error) {
	return base45.DecodeString(data)
}

// EncodedCRCSize is 6: base45 packs 4 bytes as 6 characters.
func (Base45Encoder) EncodedCRCSize() int {
	return 6
}

func (Base45Encoder) EncodingType() EncodingType {
	return EncodingTypeBase45
}

func NewEncoder(t EncodingType) (ContentEncoder, error) {
	switch t {
	case EncodingTypeBase45:
		return Base45Encoder{}, nil
	default:
		return nil, fmt.Errorf("unsupported envelope encoding type %d", t)
	}
}
