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
	"errors"
	"fmt"
	"strings"
)

// BinaryMagic is the 2-byte identifier for the binary container format.
var BinaryMagic = [2]byte{'P', 'C'}

// CurrentBinaryFormatVersion is the container format version of the binary
// container recorded in the byte following BinaryMagic. Bumped whenever the
// binary wire format changes; readers reject any other value.
const CurrentBinaryFormatVersion = 5

// BinaryHeaderSize is the fixed magic prefix of the binary container,
// comprising the 2-byte BinaryMagic and the single container format version byte.
const BinaryHeaderSize = 3

var (
	// ErrBinaryInvalidMagic indicates the binary container header does not match BinaryMagic.
	ErrBinaryInvalidMagic = errors.New("binary: invalid magic")
	// ErrBinaryUnsupportedVersion indicates the binary container uses an unsupported container format version.
	ErrBinaryUnsupportedVersion = errors.New("binary: unsupported container format version")
	// ErrBinaryTruncated indicates the binary data is shorter than the declared format.
	ErrBinaryTruncated = errors.New("binary: truncated data")
)

// ParseVersion extracts major, minor, patch from a version string like "v3.1.2".
// It errors when the string does not parse or any component falls outside the
// 0–255 range representable in the binary wire format, so serialization cannot
// silently rewrite version metadata.
func ParseVersion(v string) (major, minor, patch uint8, err error) {
	trimmed := strings.TrimPrefix(v, "v")
	var maj, mi, pat int
	if _, err := fmt.Sscanf(trimmed, "%d.%d.%d", &maj, &mi, &pat); err != nil {
		return 0, 0, 0, fmt.Errorf("unparseable version %q", v)
	}
	if maj < 0 || maj > 255 {
		return 0, 0, 0, fmt.Errorf("major %d out of range", maj)
	}
	if mi < 0 || mi > 255 {
		return 0, 0, 0, fmt.Errorf("minor %d out of range", mi)
	}
	if pat < 0 || pat > 255 {
		return 0, 0, 0, fmt.Errorf("patch %d out of range", pat)
	}
	return uint8(maj), uint8(mi), uint8(pat), nil
}

// formatVersion returns "M.m.p" from three uint8 components.
func formatVersion(major, minor, patch uint8) string {
	return fmt.Sprintf("%d.%d.%d", major, minor, patch)
}
