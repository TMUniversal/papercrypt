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
	"testing"
)

func TestRoundtrip(t *testing.T) {
	payload := []byte("hello, world")
	wrapped := Wrap(payload)
	got, err := Unwrap(wrapped)
	if err != nil {
		t.Fatalf("Unwrap: %v", err)
	}
	if !bytes.Equal(got, payload) {
		t.Errorf("roundtrip mismatch")
	}
}

func TestRoundtripEmpty(t *testing.T) {
	wrapped := Wrap([]byte{})
	got, err := Unwrap(wrapped)
	if err != nil {
		t.Fatalf("Unwrap: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected empty payload, got %d bytes", len(got))
	}
}

func TestRoundtripLarge(t *testing.T) {
	payload := bytes.Repeat([]byte{0xAB}, 10_000)
	wrapped := Wrap(payload)
	got, err := Unwrap(wrapped)
	if err != nil {
		t.Fatalf("Unwrap: %v", err)
	}
	if !bytes.Equal(got, payload) {
		t.Errorf("roundtrip mismatch")
	}
}

func TestInvalidMagic(t *testing.T) {
	wrapped := Wrap([]byte("test"))
	wrapped[0] = 'X'
	_, err := Unwrap(wrapped)
	if err != ErrInvalidMagic {
		t.Fatalf("expected ErrInvalidMagic, got %v", err)
	}
}

func TestCRCMismatch(t *testing.T) {
	wrapped := Wrap([]byte("test"))
	wrapped[4] ^= 0xFF
	_, err := Unwrap(wrapped)
	if err != ErrCRCMismatch {
		t.Fatalf("expected ErrCRCMismatch, got %v", err)
	}
}

func TestPayloadTooShort(t *testing.T) {
	_, err := Unwrap([]byte{0x01, 0x02})
	if err != ErrPayloadTooShort {
		t.Fatalf("expected ErrPayloadTooShort, got %v", err)
	}
}

func FuzzWrapUnwrap(f *testing.F) {
	f.Add([]byte(""))
	f.Add([]byte("hello"))
	f.Add(bytes.Repeat([]byte{0}, 1000))

	f.Fuzz(func(t *testing.T, data []byte) {
		wrapped := Wrap(data)
		got, err := Unwrap(wrapped)
		if err != nil {
			t.Fatalf("Unwrap failed: %v", err)
		}
		if !bytes.Equal(got, data) {
			t.Errorf("roundtrip mismatch")
		}
	})
}
