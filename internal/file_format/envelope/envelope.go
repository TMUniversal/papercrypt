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
//	"PC" + base36(info) + base36(version) + encoder(CRC32) + encoder(content)
//
// The header is a 4-character prefix: the magic "PC", followed by the
// envelope info and the envelope version, each encoded as a single
// base36 character (0-9A-Z, alphabet "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ").
// The info character encodes the envelope type in its least significant
// bit (1 = envelope), the content encoding type in the next two bits,
// and the content compression type in the fourth bit (1 = gzip).
// The CRC-32 is the IEEE checksum of the content, encoded using the same
// ContentEncoder as the payload. The content encoder (e.g. base45) is
// injected via the ContentEncoder interface, making it replaceable.
package envelope

import (
	"encoding/binary"
	"errors"
	"fmt"
	"hash/crc32"
)

var (
	ErrCRCMismatch     = errors.New("envelope: CRC-32 mismatch")
	ErrPayloadTooShort = errors.New("envelope: payload too short")
	ErrDecode          = errors.New("envelope: decode error")
)

// Wrap compresses with gzip only when it makes the payload strictly smaller.
func Wrap(content []byte, enc ContentEncoder) string {
	stored, comp := selectCompression(content)
	crc := crc32.ChecksumIEEE(stored)
	crcBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(crcBytes, crc)

	header := headerString(TypeEnvelope, enc, comp)
	return header + enc.EncodeToString(crcBytes) + enc.EncodeToString(stored)
}

// Unwrap decompresses using the compressor named in the header; the
// ContentEncoder used to decode must match the header's encoding type.
func Unwrap(data string, enc ContentEncoder, opts ...CompressorOption) ([]byte, error) {
	hdr, encoded, err := ParseHeader(data)
	if err != nil {
		return nil, err
	}
	if hdr.Type != TypeEnvelope {
		return nil, fmt.Errorf("%w: not an envelope", ErrInvalidVersion)
	}
	if hdr.Encoding != enc.EncodingType() {
		return nil, ErrEncodingType
	}
	if hdr.Version != EnvelopeVersion {
		return nil, fmt.Errorf("%w: %d", ErrInvalidVersion, hdr.Version)
	}

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

	comp, err := NewCompressor(hdr.Compression, opts...)
	if err != nil {
		return nil, err
	}
	return comp.Decompress(content)
}
