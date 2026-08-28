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

package envelope

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

var testEncoder = Base45Encoder{}

func TestRoundtrip(t *testing.T) {
	content := []byte("hello, world")
	wrapped := Wrap(content, testEncoder)
	got, err := Unwrap(wrapped, testEncoder)
	if err != nil {
		t.Fatalf("Unwrap: %v", err)
	}
	if !bytes.Equal(got, content) {
		t.Errorf("roundtrip mismatch")
	}
}

func TestRoundtripEmpty(t *testing.T) {
	wrapped := Wrap([]byte{}, testEncoder)
	got, err := Unwrap(wrapped, testEncoder)
	if err != nil {
		t.Fatalf("Unwrap: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected empty content, got %d bytes", len(got))
	}
}

func TestRoundtripLarge(t *testing.T) {
	content := bytes.Repeat([]byte{0xAB}, 10_000)
	wrapped := Wrap(content, testEncoder)
	got, err := Unwrap(wrapped, testEncoder)
	if err != nil {
		t.Fatalf("Unwrap: %v", err)
	}
	if !bytes.Equal(got, content) {
		t.Errorf("roundtrip mismatch")
	}
}

func TestEnvelopeFormat(t *testing.T) {
	content := []byte("test")
	wrapped := Wrap(content, testEncoder)

	header := Magic + versionString()
	if !strings.HasPrefix(wrapped, header) {
		t.Errorf("expected prefix %q, got %q", header, wrapped)
	}

	crcSize := testEncoder.EncodedCRCSize()
	encodedLen := len(wrapped) - len(header)
	if encodedLen < crcSize {
		t.Errorf("encoded part too short: %d < %d", encodedLen, crcSize)
	}
}

func TestInvalidMagic(t *testing.T) {
	wrapped := Wrap([]byte("test"), testEncoder)
	corrupted := "X" + wrapped[1:]
	_, err := Unwrap(corrupted, testEncoder)
	if err != ErrInvalidMagic {
		t.Fatalf("expected ErrInvalidMagic, got %v", err)
	}
}

func TestInvalidVersion(t *testing.T) {
	wrapped := Wrap([]byte("test"), testEncoder)
	headerLen := len(Magic) + len(versionString())

	tests := []string{
		// wrong numeric version
		wrapped[:headerLen-1] + "2" + wrapped[headerLen:],
		// non-numeric version
		wrapped[:headerLen-1] + "x" + wrapped[headerLen:],
	}

	for _, tc := range tests {
		_, err := Unwrap(tc, testEncoder)
		if !errors.Is(err, ErrInvalidVersion) {
			t.Fatalf("expected ErrInvalidVersion, got %v", err)
		}
	}
}

func TestCRCMismatch(t *testing.T) {
	wrapped := Wrap([]byte("test"), testEncoder)
	// Corrupt the CRC by replacing its first encoded character, keeping the length intact.
	headerLen := len(Magic) + len(versionString())
	corrupted := wrapped[:headerLen] + "!" + wrapped[headerLen+1:]
	_, err := Unwrap(corrupted, testEncoder)
	if err == nil {
		t.Fatal("expected error for corrupted CRC")
	}
}

func TestPayloadTooShort(t *testing.T) {
	_, err := Unwrap(Magic, testEncoder)
	if err != ErrPayloadTooShort {
		t.Fatalf("expected ErrPayloadTooShort, got %v", err)
	}
}

func FuzzWrapUnwrap(f *testing.F) {
	f.Add([]byte(""))
	f.Add([]byte("hello"))
	f.Add(bytes.Repeat([]byte{0}, 1000))

	f.Fuzz(func(t *testing.T, data []byte) {
		wrapped := Wrap(data, testEncoder)
		got, err := Unwrap(wrapped, testEncoder)
		if err != nil {
			t.Fatalf("Unwrap failed: %v", err)
		}
		if !bytes.Equal(got, data) {
			t.Errorf("roundtrip mismatch")
		}
	})
}
