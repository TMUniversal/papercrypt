/*
 * This file is part of PaperCrypt.
 *
 * PaperCrypt lets you prepare encrypted messages for printing on paper.
 * Copyright (C) 2026 TMUniversal <me@tmuniversal.eu>.
 *
 * PaperCrypt is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as published
 * by the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

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
	if infoIdx > 0x0f {
		return hdr, "", fmt.Errorf("%w: reserved header bits set %q", ErrInvalidVersion, rest[0])
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
