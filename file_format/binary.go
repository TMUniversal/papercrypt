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

// formatVersion returns "M.m.p" from three uint8 components.
func formatVersion(major, minor, patch uint8) string {
	return fmt.Sprintf("%d.%d.%d", major, minor, patch)
}
