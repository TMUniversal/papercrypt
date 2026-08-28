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

// headerString returns the full header prefix for typ and enc:
// Magic + base32(info) + base32(version). The info character encodes the
// header type in its least significant bit and the content encoding type
// in the next two bits.
func headerString(typ HeaderType, enc ContentEncoder) string {
	info := uint8(typ) | uint8(enc.EncodingType())<<1
	return Magic + string(headerAlphabet[info]) + string(headerAlphabet[EnvelopeVersion])
}

// parseHeader validates the header prefix of data against typ and enc,
// and returns the remaining payload section after the header.
func parseHeader(data string, enc ContentEncoder, typ HeaderType) (string, error) {
	if !strings.HasPrefix(data, Magic) {
		return "", ErrInvalidMagic
	}

	rest := data[len(Magic):]

	if len(rest) < headerChars {
		return "", ErrPayloadTooShort
	}

	infoIdx := strings.IndexByte(headerAlphabet, rest[0])
	if infoIdx == -1 {
		return "", fmt.Errorf("%w: invalid header character %q", ErrInvalidVersion, rest[0])
	}
	info := uint8(infoIdx)
	if HeaderType(info&1) != typ {
		return "", fmt.Errorf("%w: not an envelope", ErrInvalidVersion)
	}
	if EncodingType(info>>1) != enc.EncodingType() {
		return "", ErrEncodingType
	}

	versionIdx := strings.IndexByte(headerAlphabet, rest[1])
	if versionIdx == -1 {
		return "", fmt.Errorf("%w: invalid header character %q", ErrInvalidVersion, rest[1])
	}
	if uint8(versionIdx) != EnvelopeVersion {
		return "", fmt.Errorf("%w: %q", ErrInvalidVersion, rest[1])
	}

	return rest[headerChars:], nil
}
