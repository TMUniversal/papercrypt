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
	"bytes"
	"compress/gzip"
	"errors"
	"strings"
	"testing"

	"github.com/tmuniversal/papercrypt/v3/internal/decompression"
)

func TestProcessPGPDataRejectsOversizedDecompression(t *testing.T) {
	data := gzipped(t, make([]byte, 2*1024))

	_, err := decodePGPData(1024, data, nil)
	if !errors.Is(err, decompression.ErrSizeExceeded) {
		t.Fatalf("expected ErrSizeExceeded, got %v", err)
	}
}

func TestProcessPGPDataUnlimited(t *testing.T) {
	data := gzipped(t, make([]byte, 2*1024))

	if _, err := decodePGPData(-1, data, nil); err == nil {
		t.Fatal("unexpected success")
	} else if errors.Is(err, decompression.ErrSizeExceeded) {
		t.Fatalf("size-limit error raised despite unlimited mode: %v", err)
	}
}

func TestProcessPGPDataAcceptsWithinLimit(t *testing.T) {
	data := gzipped(t, make([]byte, 512))

	if _, err := decodePGPData(1024, data, nil); err == nil {
		t.Fatal("unexpected success")
	} else if strings.Contains(err.Error(), "size limit") {
		t.Fatalf("size-limit error raised within the limit: %v", err)
	}
}

func gzipped(t *testing.T, payload []byte) []byte {
	t.Helper()
	var buf bytes.Buffer
	gz, err := gzip.NewWriterLevel(&buf, gzip.BestCompression)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := gz.Write(payload); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}
