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

package file_format

import (
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
)

// MarshalBinary serializes the PaperCrypt struct to the compact binary format.
//
// Wire format:
//
//	[2]byte  magic     — "PC"
//	[1]byte  format    — container format version
//	[3]byte  version   — major, minor, patch (uint8 each)
//	[1]byte  format    — 0=PGP, 1=Raw
//	var      serial    — 1-byte length prefix + UTF-8
//	var      purpose   — 1-byte length prefix + UTF-8
//	var      comment   — 1-byte length prefix + UTF-8
//	[8]byte  createdAt — Unix timestamp in nanoseconds, big-endian
//	[32]byte dataSHA256
//	var      data      — remaining bytes
func MarshalBinary(p *PaperCrypt) ([]byte, error) {
	if p == nil {
		return nil, errors.New("binary: nil PaperCrypt")
	}

	serialBytes := []byte(p.SerialNumber)
	purposeBytes := []byte(p.Purpose)
	commentBytes := []byte(p.Comment)

	if len(serialBytes) > 255 {
		return nil, fmt.Errorf("binary: serial number too long (%d > 255)", len(serialBytes))
	}
	if len(purposeBytes) > 255 {
		return nil, fmt.Errorf("binary: purpose too long (%d > 255)", len(purposeBytes))
	}
	if len(commentBytes) > 255 {
		return nil, fmt.Errorf("binary: comment too long (%d > 255)", len(commentBytes))
	}

	major, minor, patch := parseVersion(p.Version)

	size := BinaryHeaderSize +
		3 + // version
		1 + // format
		1 + len(serialBytes) +
		1 + len(purposeBytes) +
		1 + len(commentBytes) +
		8 + // createdAt
		32 + // dataSHA256
		len(p.Data)

	out := make([]byte, 0, size)

	out = append(out, BinaryMagic[:]...)
	out = append(out, CurrentBinaryFormatVersion)
	out = append(out, major, minor, patch)
	out = append(out, byte(p.DataFormat))

	out = append(out, byte(len(serialBytes))) //nolint:gosec // length is validated <= 255 above
	out = append(out, serialBytes...)
	out = append(out, byte(len(purposeBytes))) //nolint:gosec // length is validated <= 255 above
	out = append(out, purposeBytes...)
	out = append(out, byte(len(commentBytes))) //nolint:gosec // length is validated <= 255 above
	out = append(out, commentBytes...)

	var ts [8]byte
	tsVal := uint64(p.CreatedAt.UnixNano()) //nolint:gosec // Unix timestamps fit in uint64
	binary.BigEndian.PutUint64(ts[:], tsVal)
	out = append(out, ts[:]...)

	if p.DataSHA256 == ([32]byte{}) {
		p.DataSHA256 = sha256.Sum256(p.Data)
	}
	out = append(out, p.DataSHA256[:]...)

	out = append(out, p.Data...)

	return out, nil
}
