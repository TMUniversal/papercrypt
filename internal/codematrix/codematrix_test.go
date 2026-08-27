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
	"image/color"
	"strings"
	"testing"

	"github.com/tmuniversal/papercrypt/v3/internal/file_format/envelope"
)

func TestRoundtrip(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping slow QR encode/decode test")
	}
	data := envelope.Wrap(
		[]byte(strings.Repeat("hello world ", 100)),
		envelope.Base45Encoder{},
	)

	img, err := Encode(data)
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}
	got, err := Decode(img)
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}
	if got != data {
		t.Errorf("roundtrip mismatch: got %q, want %q", got, data)
	}
}

func TestRoundtripAlphaNumeric(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping slow QR encode/decode test")
	}
	data := "ABC123$%*+-./:"
	img, err := Encode(data)
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}
	got, err := Decode(img)
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}
	if got != data {
		t.Errorf("roundtrip mismatch: got %q, want %q", got, data)
	}
}

func TestRoundtripEmpty(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping slow QR encode/decode test")
	}
	img, err := Encode("")
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}
	got, err := Decode(img)
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}
	if got != "" {
		t.Errorf("expected empty, got %q", got)
	}
}

func TestEncodePNG(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping slow QR encode/decode test")
	}
	data := envelope.Wrap([]byte("hello"), envelope.Base45Encoder{})
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
	if got != data {
		t.Errorf("roundtrip mismatch: got %q, want %q", got, data)
	}
}

func FuzzRoundtrip(f *testing.F) {
	f.Add("")
	f.Add(envelope.Wrap([]byte("HELLO WORLD $%*+-./:"), envelope.Base45Encoder{}))
	f.Add(strings.Repeat("A", 500))
	f.Add(strings.Repeat("Z", 500))

	f.Fuzz(func(t *testing.T, data string) {
		if testing.Short() {
			t.Skip("skipping slow QR encode/decode in short mode")
		}

		img, err := Encode(data)
		if err != nil {
			t.Skipf("Encode failed: %v", err)
		}

		got, err := Decode(img)
		if err != nil {
			t.Fatalf("Decode failed after successful Encode: %v", err)
		}

		if got != data {
			t.Errorf("roundtrip mismatch: got %d chars, want %d", len(got), len(data))
		}
	})
}

func FuzzDecodeRandomImage(f *testing.F) {
	f.Add(uint8(0), uint8(0))
	f.Add(uint8(1), uint8(255))

	f.Fuzz(func(_ *testing.T, seed1, seed2 uint8) {
		w, h := 100+int(seed1)*3, 100+int(seed2)*3
		img := image.NewGray(image.Rect(0, 0, w, h))
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				img.SetGray(x, y, color.Gray{
					Y: uint8((x*7 + y*13 + int(seed1)*31 + int(seed2)*37) % 256),
				})
			}
		}

		_, _ = Decode(img)
	})
}
