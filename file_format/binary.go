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

var BinaryMagic = [2]byte{'P', 'C'}

// Bumped whenever the binary wire format changes; readers reject any other value.
const CurrentBinaryFormatVersion = 5

const BinaryHeaderSize = 3

var (
	ErrBinaryInvalidMagic       = errors.New("binary: invalid magic")
	ErrBinaryUnsupportedVersion = errors.New("binary: unsupported container format version")
	ErrBinaryTruncated          = errors.New("binary: truncated data")
)

// Components must fit the uint8 wire fields, so serialization can't silently
// rewrite version metadata.
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

func formatVersion(major, minor, patch uint8) string {
	return fmt.Sprintf("%d.%d.%d", major, minor, patch)
}
