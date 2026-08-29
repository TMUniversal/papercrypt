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
	"encoding/binary"
	"errors"
	"hash/crc32"
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

	header := headerString(TypeEnvelope, testEncoder, CompressionRaw)
	if !strings.HasPrefix(wrapped, header) {
		t.Errorf("expected prefix %q, got %q", header, wrapped)
	}
	if header != "PC31" {
		t.Errorf("expected header %q, got %q", "PC31", header)
	}

	crcSize := testEncoder.EncodedCRCSize()
	encodedLen := len(wrapped) - len(header)
	if encodedLen < crcSize {
		t.Errorf("encoded part too short: %d < %d", encodedLen, crcSize)
	}
}

func TestEnvelopeGzipCompression(t *testing.T) {
	// Highly compressible content: gzip makes it smaller, so the envelope
	// must store it compressed and set the gzip header bit.
	content := bytes.Repeat([]byte{0xAB}, 10_000)
	wrapped := Wrap(content, testEncoder)

	header := headerString(TypeEnvelope, testEncoder, CompressionGzip)
	if !strings.HasPrefix(wrapped, header) {
		t.Errorf("expected gzip header %q, got %q", header, wrapped)
	}

	hdr, _, err := ParseHeader(wrapped)
	if err != nil {
		t.Fatalf("ParseHeader: %v", err)
	}
	if hdr.Compression != CompressionGzip {
		t.Errorf("expected CompressionGzip, got %v", hdr.Compression)
	}

	got, err := Unwrap(wrapped, testEncoder)
	if err != nil {
		t.Fatalf("Unwrap: %v", err)
	}
	if !bytes.Equal(got, content) {
		t.Errorf("roundtrip mismatch after decompression")
	}
}

func TestEnvelopeRawKeptWhenGzipLarger(t *testing.T) {
	// Small content: gzip makes it larger, so it must stay raw.
	content := []byte("test")
	wrapped := Wrap(content, testEncoder)

	hdr, _, err := ParseHeader(wrapped)
	if err != nil {
		t.Fatalf("ParseHeader: %v", err)
	}
	if hdr.Compression != CompressionRaw {
		t.Errorf("expected CompressionRaw, got %v", hdr.Compression)
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
	headerLen := len(Magic) + 2

	tests := []string{
		// wrong numeric version
		wrapped[:headerLen-1] + "2" + wrapped[headerLen:],
		// non-header version character
		wrapped[:headerLen-1] + "x" + wrapped[headerLen:],
	}

	for _, tc := range tests {
		_, err := Unwrap(tc, testEncoder)
		if !errors.Is(err, ErrInvalidVersion) {
			t.Fatalf("expected ErrInvalidVersion, got %v", err)
		}
	}
}

func TestInvalidType(t *testing.T) {
	content := []byte("test")
	comp := CompressionRaw
	stored, _ := selectCompression(content)
	crc := crc32.ChecksumIEEE(stored)
	crcBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(crcBytes, crc)
	containerEnvelope := headerString(TypeContainer, testEncoder, comp) +
		testEncoder.EncodeToString(crcBytes) + testEncoder.EncodeToString(stored)

	_, err := Unwrap(containerEnvelope, testEncoder)
	if !errors.Is(err, ErrInvalidType) {
		t.Fatalf("expected ErrInvalidType, got %v", err)
	}
	if !strings.Contains(err.Error(), "not an envelope") {
		t.Fatalf("expected 'not an envelope' diagnostic, got %q", err.Error())
	}
}

func TestEncodingTypeMismatch(t *testing.T) {
	wrapped := Wrap([]byte("test"), testEncoder)
	// The info char encodes (type<<1)|envelope; corrupt the encoding type bits.
	corrupted := wrapped[:2] + "1" + wrapped[3:]
	_, err := Unwrap(corrupted, testEncoder)
	if !errors.Is(err, ErrEncodingType) {
		t.Fatalf("expected ErrEncodingType, got %v", err)
	}
}

func TestCRCMismatch(t *testing.T) {
	wrapped := Wrap([]byte("test"), testEncoder)
	// Corrupt the CRC by replacing its first encoded character, keeping the length intact.
	headerLen := len(Magic) + 2
	corrupted := wrapped[:headerLen] + "!" + wrapped[headerLen+1:]
	_, err := Unwrap(corrupted, testEncoder)
	if err == nil {
		t.Fatal("expected error for corrupted CRC")
	}
}

func TestParseHeader(t *testing.T) {
	wrapped := Wrap([]byte("test"), testEncoder)

	hdr, rest, err := ParseHeader(wrapped)
	if err != nil {
		t.Fatalf("ParseHeader: %v", err)
	}
	if hdr.Type != TypeEnvelope {
		t.Errorf("expected TypeEnvelope, got %v", hdr.Type)
	}
	if hdr.Encoding != EncodingTypeBase45 {
		t.Errorf("expected EncodingTypeBase45, got %v", hdr.Encoding)
	}
	if hdr.Version != EnvelopeVersion {
		t.Errorf("expected version %d, got %d", EnvelopeVersion, hdr.Version)
	}
	if hdr.Compression != CompressionRaw {
		t.Errorf("expected CompressionRaw, got %v", hdr.Compression)
	}

	if len(rest) == 0 {
		t.Errorf("expected payload section after header, got empty rest")
	}

	// The remaining section must decode with the encoder chosen from the header.
	enc := Base45Encoder{}
	if _, err := enc.DecodeString(rest); err != nil {
		t.Errorf("payload section does not decode as base45: %v", err)
	}
}

func TestPayloadTooShort(t *testing.T) {
	_, err := Unwrap(Magic, testEncoder)
	if err != ErrPayloadTooShort {
		t.Fatalf("expected ErrPayloadTooShort, got %v", err)
	}
}

func TestRawCompressorIdentity(t *testing.T) {
	input := []byte("abc123")
	comp := RawCompressor{}
	if comp.CompressionType() != CompressionRaw {
		t.Errorf("expected CompressionRaw, got %v", comp.CompressionType())
	}

	out, err := comp.Compress(input)
	if err != nil {
		t.Fatalf("Compress: %v", err)
	}
	if !bytes.Equal(out, input) {
		t.Errorf("Compress mutated data")
	}

	out, err = comp.Decompress(input)
	if err != nil {
		t.Fatalf("Decompress: %v", err)
	}
	if !bytes.Equal(out, input) {
		t.Errorf("Decompress mutated data")
	}
}

func TestGzipCompressorRoundtrip(t *testing.T) {
	input := bytes.Repeat([]byte{0xAB}, 10_000)
	comp := GzipCompressor{}
	if comp.CompressionType() != CompressionGzip {
		t.Errorf("expected CompressionGzip, got %v", comp.CompressionType())
	}

	compressed, err := comp.Compress(input)
	if err != nil {
		t.Fatalf("Compress: %v", err)
	}
	if len(compressed) >= len(input) {
		t.Errorf("expected compressed output to be smaller")
	}
	if len(compressed) < 2 || compressed[0] != 0x1f || compressed[1] != 0x8b {
		t.Errorf("output does not carry the gzip magic header")
	}

	out, err := comp.Decompress(compressed)
	if err != nil {
		t.Fatalf("Decompress: %v", err)
	}
	if !bytes.Equal(out, input) {
		t.Errorf("roundtrip mismatch after gzip decompression")
	}
}

func TestGzipCompressorRejectsInvalidData(t *testing.T) {
	_, err := GzipCompressor{}.Decompress([]byte("not gzip"))
	if err == nil {
		t.Fatal("expected error decompressing invalid gzip data")
	}
}

func TestGzipCompressorRejectsOversizedOutput(t *testing.T) {
	const customLimit = 1024
	bomb := bytes.Repeat([]byte{0}, customLimit*2)
	compressed, err := (GzipCompressor{}).Compress(bomb)
	if err != nil {
		t.Fatalf("Compress: %v", err)
	}

	comp, err := NewCompressor(CompressionGzip, WithMaxDecompressedSize(customLimit))
	if err != nil {
		t.Fatalf("NewCompressor: %v", err)
	}

	if _, err := comp.Decompress(compressed); !errors.Is(err, ErrDecompressedSizeExceeded) {
		t.Fatalf("expected ErrDecompressedSizeExceeded, got %v", err)
	}
}

func TestGzipCompressorNoDecompressionLimit(t *testing.T) {
	const customLimit = 1024
	bomb := bytes.Repeat([]byte{0}, customLimit*2)
	compressed, err := (GzipCompressor{}).Compress(bomb)
	if err != nil {
		t.Fatalf("Compress: %v", err)
	}

	comp, err := NewCompressor(CompressionGzip, WithNoDecompressionLimit())
	if err != nil {
		t.Fatalf("NewCompressor: %v", err)
	}

	out, err := comp.Decompress(compressed)
	if err != nil {
		t.Fatalf("Decompress with disabled limit: %v", err)
	}
	if !bytes.Equal(out, bomb) {
		t.Errorf("roundtrip mismatch when the size limit is disabled")
	}
}

func TestNewCompressor(t *testing.T) {
	if c, err := NewCompressor(CompressionRaw); err != nil {
		t.Fatalf("NewCompressor(CompressionRaw): %v", err)
	} else if _, ok := c.(RawCompressor); !ok {
		t.Errorf("expected RawCompressor, got %T", c)
	}

	if c, err := NewCompressor(CompressionGzip); err != nil {
		t.Fatalf("NewCompressor(CompressionGzip): %v", err)
	} else if _, ok := c.(GzipCompressor); !ok {
		t.Errorf("expected GzipCompressor, got %T", c)
	}

	if _, err := NewCompressor(CompressionType(9)); err == nil {
		t.Fatal("expected error for unsupported compression type")
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
