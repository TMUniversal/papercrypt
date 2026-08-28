package envelope

import (
	"errors"
	"fmt"
	"strings"
)

// Magic is the string identifier for the envelope format.
const Magic = "PC"

// EnvelopeVersion is the current version of the envelope format,
// encoded into the header as a single base32 character.
const EnvelopeVersion = 1

// HeaderType identifies the kind of payload described by the header.
type HeaderType uint8

const (
	// TypeContainer marks a raw binary container payload.
	TypeContainer HeaderType = 0
	// TypeEnvelope marks an envelope payload.
	TypeEnvelope HeaderType = 1
)

var (
	// ErrInvalidMagic indicates the envelope header does not start with Magic.
	ErrInvalidMagic = errors.New("envelope: invalid magic")
	// ErrInvalidVersion indicates the envelope version does not match EnvelopeVersion.
	ErrInvalidVersion = errors.New("envelope: unsupported envelope version")
	// ErrEncodingType indicates the header encoding type does not match the provided encoder.
	ErrEncodingType = errors.New("envelope: encoding type mismatch")
)

// headerAlphabet is the base32 alphabet used for the envelope header
// fields. Characters are ordered "0-9A-Z".
const headerAlphabet = "0123456789ABCDEFGHIJKLMNOPQRSTUV"

// headerChars is the fixed number of header characters following Magic.
const headerChars = 2

// Header contains the fields parsed from an envelope header.
type Header struct {
	// Type is the payload type stored in the least significant
	// bit of the info header field.
	Type HeaderType
	// Encoding is the content encoding stored in the next two
	// bits of the info header field.
	Encoding EncodingType
	// Version is the envelope format version.
	Version uint8
}

// headerString returns the full header prefix for typ and enc:
// Magic + base32(info) + base32(version). The info character encodes the
// header type in its least significant bit and the content encoding type
// in the next two bits.
func headerString(typ HeaderType, enc ContentEncoder) string {
	info := uint8(typ) | uint8(enc.EncodingType())<<1
	return Magic + string(headerAlphabet[info]) + string(headerAlphabet[EnvelopeVersion])
}

// ParseHeader validates that data starts with a well-formed envelope
// header and returns the parsed fields together with the payload section
// following the header. It does not validate against a ContentEncoder;
// callers inspect Header.Encoding to choose the encoder to pass to Unwrap.
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
	info := uint8(infoIdx)
	hdr.Type = HeaderType(info & 1)
	hdr.Encoding = EncodingType(info >> 1)

	versionIdx := strings.IndexByte(headerAlphabet, rest[1])
	if versionIdx == -1 {
		return hdr, "", fmt.Errorf("%w: invalid header character %q", ErrInvalidVersion, rest[1])
	}
	hdr.Version = uint8(versionIdx)

	return hdr, rest[headerChars:], nil
}
