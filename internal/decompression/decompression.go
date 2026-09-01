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

// Package decompression is the single owner of the gzip-expansion size cap
// shared by every layer that inflates untrusted input.
package decompression

import (
	"errors"
	"fmt"
	"io"
)

// MaxSize is the module-wide default cap on gzip-expanded output, guarding
// against decompression bombs.
const MaxSize = 1 << 30 // 1 GiB

// ErrSizeExceeded is returned when gzip-expanded output exceeds the cap.
var ErrSizeExceeded = errors.New("decompressed data exceeds the size limit")

// ReadAll reads all of r, rejecting output above limit bytes with
// ErrSizeExceeded. A negative limit disables the cap; zero uses MaxSize.
func ReadAll(r io.Reader, limit int) ([]byte, error) {
	limitBytes := limit
	if limitBytes == 0 {
		limitBytes = MaxSize
	}

	in := r
	if limitBytes > 0 {
		in = io.LimitReader(in, int64(limitBytes)+1)
	}

	out, err := io.ReadAll(in)
	if err != nil {
		return nil, err
	}
	if limitBytes > 0 && len(out) > limitBytes {
		return nil, fmt.Errorf("%w: exceeds %d bytes", ErrSizeExceeded, limitBytes)
	}
	return out, nil
}
