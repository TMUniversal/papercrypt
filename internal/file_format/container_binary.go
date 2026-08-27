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
	"io"
	"strings"
	"time"
)

// BinaryMagic is the 4-byte identifier for the binary container format.
var BinaryMagic = [4]byte{'P', 'C', 0x03, 0x00}

// BinaryHeaderSize is the fixed magic prefix of the binary container.
const BinaryHeaderSize = 4

var (
	// ErrBinaryInvalidMagic indicates the binary container header does not match BinaryMagic.
	ErrBinaryInvalidMagic = errors.New("binary: invalid magic")
	// ErrBinaryTruncated indicates the binary data is shorter than the declared format.
	ErrBinaryTruncated = errors.New("binary: truncated data")
)

// parseVersion extracts major, minor, patch from a version string like "v3.1.2".
// Returns 0,0,0 for unparseable strings.
func parseVersion(v string) (major, minor, patch uint8) {
	v = strings.TrimPrefix(v, "v")
	var maj, mi, pat int
	if _, err := fmt.Sscanf(v, "%d.%d.%d", &maj, &mi, &pat); err != nil {
		return 0, 0, 0
	}
	return uint8(maj), uint8(mi), uint8(pat) //nolint:gosec // version components fit in uint8
}

// formatVersion returns "vM.m.p" from three uint8 components.
func formatVersion(major, minor, patch uint8) string {
	return fmt.Sprintf("v%d.%d.%d", major, minor, patch)
}

// MarshalBinary serializes the PaperCrypt struct to the compact binary format.
//
// Wire format:
//
//	[4]byte  magic     — "PC\x03\x00"
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
		1 + // format
		3 + // version
		1 + len(serialBytes) +
		1 + len(purposeBytes) +
		1 + len(commentBytes) +
		8 + // createdAt
		32 + // dataSHA256
		len(p.Data)

	out := make([]byte, 0, size)

	out = append(out, BinaryMagic[:]...)
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

// UnmarshalBinary parses a binary container into a PaperCrypt struct.
func UnmarshalBinary(data []byte) (*PaperCrypt, error) {
	if len(data) < BinaryHeaderSize {
		return nil, ErrBinaryTruncated
	}

	if [4]byte(data[0:4]) != BinaryMagic {
		return nil, ErrBinaryInvalidMagic
	}

	r := data[BinaryHeaderSize:]
	p := &PaperCrypt{}

	if len(r) < 3 {
		return nil, ErrBinaryTruncated
	}
	p.Version = formatVersion(r[0], r[1], r[2])
	r = r[3:]

	if len(r) < 1 {
		return nil, ErrBinaryTruncated
	}
	p.DataFormat = PaperCryptDataFormat(r[0])
	r = r[1:]

	if len(r) < 1 {
		return nil, ErrBinaryTruncated
	}
	serialLen := int(r[0])
	r = r[1:]
	if len(r) < serialLen {
		return nil, ErrBinaryTruncated
	}
	p.SerialNumber = string(r[:serialLen])
	r = r[serialLen:]

	if len(r) < 1 {
		return nil, ErrBinaryTruncated
	}
	purposeLen := int(r[0])
	r = r[1:]
	if len(r) < purposeLen {
		return nil, ErrBinaryTruncated
	}
	p.Purpose = string(r[:purposeLen])
	r = r[purposeLen:]

	if len(r) < 1 {
		return nil, ErrBinaryTruncated
	}
	commentLen := int(r[0])
	r = r[1:]
	if len(r) < commentLen {
		return nil, ErrBinaryTruncated
	}
	p.Comment = string(r[:commentLen])
	r = r[commentLen:]

	if len(r) < 8 {
		return nil, ErrBinaryTruncated
	}
	tsVal := binary.BigEndian.Uint64(r[:8])
	p.CreatedAt = time.Unix(0, int64(tsVal)) //nolint:gosec // Unix timestamps are non-negative
	r = r[8:]

	if len(r) < 32 {
		return nil, ErrBinaryTruncated
	}
	copy(p.DataSHA256[:], r[:32])
	r = r[32:]

	p.Data = r

	return p, nil
}

// UnmarshalBinaryFromReader reads a binary container from r and returns it.
func UnmarshalBinaryFromReader(r io.Reader) (*PaperCrypt, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("binary: read: %w", err)
	}
	return UnmarshalBinary(data)
}
