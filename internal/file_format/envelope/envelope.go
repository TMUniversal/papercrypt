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

// Package envelope provides a binary envelope format with a CRC-32 checksum
// for integrity verification of enclosed payloads.
//
// Wire format:
//
//	[4]byte magic  — "PCE1"
//	[4]byte crc32  — IEEE CRC-32 of payload, big-endian
//	[N]byte payload — the enclosed data
package envelope

import (
	"encoding/binary"
	"errors"
	"hash/crc32"
)

// Magic is the 4-byte identifier for the envelope format.
var Magic = [4]byte{'P', 'C', 'E', '1'}

// HeaderSize is the total size of the envelope header (magic + CRC).
const HeaderSize = 8

var (
	// ErrInvalidMagic indicates the envelope header does not match Magic.
	ErrInvalidMagic = errors.New("envelope: invalid magic")
	// ErrCRCMismatch indicates the CRC-32 checksum does not match the payload.
	ErrCRCMismatch = errors.New("envelope: CRC-32 mismatch")
	// ErrPayloadTooShort indicates the data is shorter than HeaderSize.
	ErrPayloadTooShort = errors.New("envelope: payload too short")
)

// Wrap wraps payload in an envelope with a CRC-32 checksum.
func Wrap(payload []byte) []byte {
	out := make([]byte, HeaderSize+len(payload))
	copy(out[0:4], Magic[:])
	binary.BigEndian.PutUint32(out[4:8], crc32.ChecksumIEEE(payload))
	copy(out[8:], payload)
	return out
}

// Unwrap validates the envelope and returns the payload.
func Unwrap(data []byte) ([]byte, error) {
	if len(data) < HeaderSize {
		return nil, ErrPayloadTooShort
	}

	if [4]byte(data[0:4]) != Magic {
		return nil, ErrInvalidMagic
	}

	stored := binary.BigEndian.Uint32(data[4:8])
	payload := data[HeaderSize:]

	if crc32.ChecksumIEEE(payload) != stored {
		return nil, ErrCRCMismatch
	}

	return payload, nil
}
