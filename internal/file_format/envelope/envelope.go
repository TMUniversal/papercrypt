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

// Package envelope provides a text envelope format with a CRC-32 checksum
// for integrity verification of enclosed payloads.
//
// Wire format:
//
//	"PCE" + version + encoder(CRC32) + encoder(content)
//
// The envelope version is the decimal representation of EnvelopeVersion.
// The CRC-32 is IEEE checksum of the content, encoded using the same
// ContentEncoder as the payload. The content encoder (e.g. base45) is
// injected via the ContentEncoder interface, making it replaceable.
package envelope

import (
	"encoding/binary"
	"errors"
	"fmt"
	"hash/crc32"
	"strconv"
	"strings"
)

// Magic is the string identifier for the envelope format, without the envelope version suffix.
const Magic = "PCE"

// EnvelopeVersion is the current version of the envelope format, appended to Magic as a decimal string.
const EnvelopeVersion = 1

var (
	// ErrInvalidMagic indicates the envelope header does not start with Magic.
	ErrInvalidMagic = errors.New("envelope: invalid magic")
	// ErrInvalidVersion indicates the envelope version does not match EnvelopeVersion.
	ErrInvalidVersion = errors.New("envelope: unsupported envelope version")
	// ErrCRCMismatch indicates the CRC-32 checksum does not match the content.
	ErrCRCMismatch = errors.New("envelope: CRC-32 mismatch")
	// ErrPayloadTooShort indicates the data is shorter than the envelope header.
	ErrPayloadTooShort = errors.New("envelope: payload too short")
	// ErrDecode indicates the content could not be decoded.
	ErrDecode = errors.New("envelope: decode error")
)

// versionString is the fixed-width decimal representation of EnvelopeVersion,
// which follows the Magic in the wire format.
func versionString() string {
	return strconv.Itoa(EnvelopeVersion)
}

// Wrap encodes content using the provided encoder, computes a CRC-32
// checksum, encodes the checksum with the same encoder, and returns
// the envelope string: "PCE" + version + encoder(CRC32) + encoder(content).
func Wrap(content []byte, enc ContentEncoder) string {
	crc := crc32.ChecksumIEEE(content)
	crcBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(crcBytes, crc)

	return Magic + versionString() + enc.EncodeToString(crcBytes) + enc.EncodeToString(content)
}

// Unwrap validates the envelope and returns the content.
// It parses "PCE" + version + encodedCRC + encodedContent, decodes both parts,
// and verifies the CRC-32 checksum.
func Unwrap(data string, enc ContentEncoder) ([]byte, error) {
	if !strings.HasPrefix(data, Magic) {
		return nil, ErrInvalidMagic
	}

	rest := data[len(Magic):]

	versionLen := len(versionString())
	if len(rest) < versionLen {
		return nil, ErrPayloadTooShort
	}
	if rest[:versionLen] != versionString() {
		return nil, fmt.Errorf("%w: %q", ErrInvalidVersion, rest[:versionLen])
	}

	encoded := rest[versionLen:]

	crcSize := enc.EncodedCRCSize()

	if len(encoded) < crcSize {
		return nil, ErrPayloadTooShort
	}

	encodedCRC := encoded[:crcSize]
	encodedContent := encoded[crcSize:]

	crcBytes, err := enc.DecodeString(encodedCRC)
	if err != nil {
		return nil, errors.Join(ErrDecode, err)
	}
	if len(crcBytes) != 4 {
		return nil, ErrPayloadTooShort
	}
	storedCRC := binary.BigEndian.Uint32(crcBytes)

	content, err := enc.DecodeString(encodedContent)
	if err != nil {
		return nil, errors.Join(ErrDecode, err)
	}

	if crc32.ChecksumIEEE(content) != storedCRC {
		return nil, fmt.Errorf(
			"%w: expected %08X, found %08X (content length %d)",
			ErrCRCMismatch,
			storedCRC,
			crc32.ChecksumIEEE(content),
			len(content),
		)
	}

	return content, nil
}
