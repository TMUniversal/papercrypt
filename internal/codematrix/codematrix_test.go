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

package codematrix

import (
	"bytes"
	"image"
	"strings"
	"testing"
)

func TestRoundtrip(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping encode/decode test")
	}
	data := append(
		[]byte(`{"v":"3.0.0-dev","f":"PGP","sn":"test","d":"`),
		bytes.Repeat([]byte("hello world "), 100)...,
	)
	data = append(data, []byte(`"}`)...)
	img, err := Encode(data)
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}
	got, err := Decode(img)
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}
	if !bytes.Equal(got, data) {
		t.Errorf("roundtrip mismatch")
	}
}

func TestRoundtripRawBytes(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping encode/decode test")
	}
	data := make([]byte, 500)
	for i := range data {
		data[i] = byte(i % 251)
	}
	img, err := Encode(data)
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}
	got, err := Decode(img)
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}
	if !bytes.Equal(got, data) {
		t.Errorf("roundtrip mismatch")
	}
}

func TestRoundtripEmpty(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping encode/decode test")
	}
	img, err := Encode([]byte{})
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}
	got, err := Decode(img)
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected empty, got %d bytes", len(got))
	}
}

func TestEncodePNG(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping encode/decode test")
	}
	data := []byte("hello")
	pngBytes, err := EncodePNG(data)
	if err != nil {
		t.Fatalf("EncodePNG: %v", err)
	}
	if len(pngBytes) == 0 {
		t.Fatal("empty PNG output")
	}
	img, _, err := image.Decode(bytes.NewReader(pngBytes))
	if err != nil {
		t.Fatalf("decode PNG: %v", err)
	}
	got, err := Decode(img)
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}
	if !bytes.Equal(got, data) {
		t.Errorf("roundtrip mismatch")
	}
}

func TestDecodeDecompressionBomb(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping encode/decode test")
	}
	// Lower the limit so Encode can produce a payload that triggers it.
	// The barcode format caps out at ~3 KiB of binary data, so the
	// decompressed size through Encode→Decode is bounded well below 10 MiB.
	saved := MaxDecodedPayloadSize
	MaxDecodedPayloadSize = 200
	defer func() { MaxDecodedPayloadSize = saved }()

	data := bytes.Repeat([]byte{0}, 300)
	img, err := Encode(data)
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}

	_, err = Decode(img)
	if err == nil {
		t.Fatal("expected error for decompression bomb, got nil")
	}
	if !strings.Contains(err.Error(), "exceeds maximum size") {
		t.Fatalf("expected 'exceeds maximum size' error, got: %v", err)
	}
}

func TestDecodeLimitDisabled(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping encode/decode test")
	}
	saved := MaxDecodedPayloadSize
	MaxDecodedPayloadSize = 200
	defer func() { MaxDecodedPayloadSize = saved }()

	SetLimitDecodedPayload(false)
	defer SetLimitDecodedPayload(true)

	data := bytes.Repeat([]byte{0}, 300)
	img, err := Encode(data)
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}

	got, err := Decode(img)
	if err != nil {
		t.Fatalf("Decode with limit disabled: unexpected error: %v", err)
	}
	if !bytes.Equal(got, data) {
		t.Errorf("disabled limit roundtrip mismatch: got %d bytes, want %d", len(got), len(data))
	}
}
