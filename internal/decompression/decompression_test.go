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

package decompression

import (
	"bytes"
	"errors"
	"io"
	"testing"
)

func TestReadAllRejectsOversizedOutput(t *testing.T) {
	in := bytes.NewReader(bytes.Repeat([]byte("x"), 16))
	if _, err := ReadAll(in, 8); !errors.Is(err, ErrSizeExceeded) {
		t.Fatalf("expected ErrSizeExceeded, got %v", err)
	}
}

func TestReadAllWithinLimit(t *testing.T) {
	in := bytes.NewReader(bytes.Repeat([]byte("x"), 8))
	out, err := ReadAll(in, 8)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(out) != 8 {
		t.Fatalf("got %d bytes, want 8", len(out))
	}
}

func TestReadAllErrorFromUnderlyingReader(t *testing.T) {
	if _, err := ReadAll(errReader{}, 8); err == nil || errors.Is(err, ErrSizeExceeded) {
		t.Fatalf("expected underlying error, got %v", err)
	}
}

type errReader struct{}

func (errReader) Read([]byte) (int, error) {
	return 0, io.ErrUnexpectedEOF
}
