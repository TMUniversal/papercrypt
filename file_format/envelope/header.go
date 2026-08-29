package envelope

import (
	"errors"
	"fmt"
	"strings"
)

const Magic = "PC"

const EnvelopeVersion = 1

type HeaderType uint8

const (
	TypeContainer HeaderType = 0
	TypeEnvelope  HeaderType = 1
)

var (
	ErrInvalidMagic   = errors.New("envelope: invalid magic")
	ErrInvalidVersion = errors.New("envelope: unsupported envelope version")
	ErrInvalidType    = errors.New("envelope: unsupported envelope type")
	ErrEncodingType   = errors.New("envelope: encoding type mismatch")
)

const headerAlphabet = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"

const headerChars = 2

type Header struct {
	Type        HeaderType
	Encoding    EncodingType
	Compression CompressionType
	Version     uint8
}

func headerString(typ HeaderType, enc ContentEncoder, comp CompressionType) string {
	info := uint8(typ) | uint8(enc.EncodingType())<<1 | uint8(comp)<<3
	return Magic + string(headerAlphabet[info]) + string(headerAlphabet[EnvelopeVersion])
}

// ParseHeader does not pick a ContentEncoder; the caller inspects
// Header.Encoding to choose the encoder to pass to Unwrap.
func ParseHeader(data string) (Header, string, error) {
	var hdr Header

	if !strings.HasPrefix(data, Magic) {
		return hdr, "", ErrInvalidMagic
	}

	rest := data[len(Magic):]

	if len(rest) < headerChars {
		return hdr, "", ErrPayloadTooShort
	}

	infoIdx := strings.IndexByte(headerAlphabet, rest[0])
	if infoIdx == -1 {
		return hdr, "", fmt.Errorf("%w: invalid header character %q", ErrInvalidVersion, rest[0])
	}
	info := uint8(infoIdx) //nolint:gosec // index is valid alphabet position
	hdr.Type = HeaderType(info & 1)
	hdr.Encoding = EncodingType((info >> 1) & 0b11)
	hdr.Compression = CompressionType((info >> 3) & 1)

	versionIdx := strings.IndexByte(headerAlphabet, rest[1])
	if versionIdx == -1 {
		return hdr, "", fmt.Errorf("%w: invalid header character %q", ErrInvalidVersion, rest[1])
	}
	hdr.Version = uint8(versionIdx) //nolint:gosec // index is valid alphabet position

	return hdr, rest[headerChars:], nil
}
