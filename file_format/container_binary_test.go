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
	"crypto/sha256"
	"errors"
	"testing"
	"time"

	"github.com/tmuniversal/papercrypt/v3/file_format/envelope"
)

func TestBinaryRoundtrip(t *testing.T) {
	pc := &PaperCrypt{
		Version:      "3.0.0",
		DataFormat:   PaperCryptDataFormatPGP,
		SerialNumber: "ABC123",
		Purpose:      "Backup",
		Comment:      "Test comment",
		CreatedAt:    time.Date(2026, 8, 26, 12, 0, 0, 123456789, time.UTC),
		Data:         []byte("hello, world"),
	}
	pc.DataSHA256 = sha256.Sum256(pc.Data)

	data, err := MarshalBinary(pc)
	if err != nil {
		t.Fatalf("MarshalBinary: %v", err)
	}

	got, err := UnmarshalBinary(data)
	if err != nil {
		t.Fatalf("UnmarshalBinary: %v", err)
	}

	if got.Version != pc.Version {
		t.Errorf("Version: got %q, want %q", got.Version, pc.Version)
	}
	if got.SerialNumber != pc.SerialNumber {
		t.Errorf("SerialNumber: got %q, want %q", got.SerialNumber, pc.SerialNumber)
	}
	if got.Purpose != pc.Purpose {
		t.Errorf("Purpose: got %q, want %q", got.Purpose, pc.Purpose)
	}
	if got.Comment != pc.Comment {
		t.Errorf("Comment: got %q, want %q", got.Comment, pc.Comment)
	}
	if got.DataFormat != pc.DataFormat {
		t.Errorf("DataFormat: got %v, want %v", got.DataFormat, pc.DataFormat)
	}
	if !got.CreatedAt.Equal(pc.CreatedAt) {
		t.Errorf("CreatedAt: got %v, want %v", got.CreatedAt, pc.CreatedAt)
	}
	if got.DataSHA256 != pc.DataSHA256 {
		t.Errorf("DataSHA256 mismatch")
	}
	if !bytes.Equal(got.Data, pc.Data) {
		t.Errorf("Data mismatch")
	}
}

func TestBinaryRoundtripEmptyFields(t *testing.T) {
	pc := &PaperCrypt{
		Version:    "3.0.0",
		DataFormat: PaperCryptDataFormatRaw,
		CreatedAt:  time.Now(),
		Data:       []byte{0x01, 0x02, 0x03},
	}
	pc.DataSHA256 = sha256.Sum256(pc.Data)

	data, err := MarshalBinary(pc)
	if err != nil {
		t.Fatalf("MarshalBinary: %v", err)
	}

	got, err := UnmarshalBinary(data)
	if err != nil {
		t.Fatalf("UnmarshalBinary: %v", err)
	}

	if got.Version != pc.Version {
		t.Errorf("Version: got %q, want %q", got.Version, pc.Version)
	}
	if got.SerialNumber != "" {
		t.Errorf("expected empty SerialNumber, got %q", got.SerialNumber)
	}
	if got.Purpose != "" {
		t.Errorf("expected empty Purpose, got %q", got.Purpose)
	}
	if got.Comment != "" {
		t.Errorf("expected empty Comment, got %q", got.Comment)
	}
	if !bytes.Equal(got.Data, pc.Data) {
		t.Errorf("Data mismatch")
	}
}

func TestBinaryWithEnvelope(t *testing.T) {
	pc := &PaperCrypt{
		Version:      "3.0.0",
		DataFormat:   PaperCryptDataFormatPGP,
		SerialNumber: "TEST01",
		Purpose:      "Testing",
		Comment:      "Envelope roundtrip",
		CreatedAt:    time.Date(2026, 1, 1, 0, 0, 0, 987654321, time.UTC),
		Data:         bytes.Repeat([]byte{0xFF}, 500),
	}
	pc.DataSHA256 = sha256.Sum256(pc.Data)

	bin, err := MarshalBinary(pc)
	if err != nil {
		t.Fatalf("MarshalBinary: %v", err)
	}

	wrapped := envelope.Wrap(bin, envelope.Base45Encoder{})
	got, err := envelope.Unwrap(wrapped, envelope.Base45Encoder{})
	if err != nil {
		t.Fatalf("Unwrap: %v", err)
	}

	gotPC, err := UnmarshalBinary(got)
	if err != nil {
		t.Fatalf("UnmarshalBinary: %v", err)
	}

	if !bytes.Equal(gotPC.Data, pc.Data) {
		t.Errorf("Data mismatch after envelope roundtrip")
	}
}

func TestBinaryInvalidMagic(t *testing.T) {
	pc := &PaperCrypt{
		Version:    "3.0.0",
		DataFormat: PaperCryptDataFormatRaw,
		CreatedAt:  time.Now(),
		Data:       []byte("test"),
	}

	data, err := MarshalBinary(pc)
	if err != nil {
		t.Fatalf("MarshalBinary: %v", err)
	}

	data[0] = 'X'
	_, err = UnmarshalBinary(data)
	if err != ErrBinaryInvalidMagic {
		t.Fatalf("expected ErrBinaryInvalidMagic, got %v", err)
	}
}

func TestBinaryUnsupportedFormatVersion(t *testing.T) {
	pc := &PaperCrypt{
		Version:    "3.0.0",
		DataFormat: PaperCryptDataFormatRaw,
		CreatedAt:  time.Now(),
		Data:       []byte("test"),
	}

	data, err := MarshalBinary(pc)
	if err != nil {
		t.Fatalf("MarshalBinary: %v", err)
	}

	data[2] = CurrentBinaryFormatVersion + 1
	_, err = UnmarshalBinary(data)
	if !errors.Is(err, ErrBinaryUnsupportedVersion) {
		t.Fatalf("expected ErrBinaryUnsupportedVersion, got %v", err)
	}
}

func TestBinaryTruncated(t *testing.T) {
	pc := &PaperCrypt{
		Version:      "3.0.0",
		DataFormat:   PaperCryptDataFormatRaw,
		SerialNumber: "SERIAL",
		Purpose:      "PURPOSE",
		Comment:      "COMMENT",
		CreatedAt:    time.Now(),
		Data:         []byte("data"),
	}
	pc.DataSHA256 = sha256.Sum256(pc.Data)

	full, err := MarshalBinary(pc)
	if err != nil {
		t.Fatalf("MarshalBinary: %v", err)
	}

	// Cut points truncate the payload right before each r[0] access
	// (DataFormat, serialLen, purposeLen, commentLen), which must return
	// ErrBinaryTruncated instead of panicking on an out-of-range index.
	cases := []struct {
		name string
		data []byte
	}{
		{"shorter than magic", []byte{0x01, 0x02}},
		{"before DataFormat", full[:BinaryHeaderSize+3]},
		{"before serial length", full[:BinaryHeaderSize+3+1]},
		{"before purpose length", full[:BinaryHeaderSize+3+1+1+len(pc.SerialNumber)]},
		{
			"before comment length",
			full[:BinaryHeaderSize+3+1+1+len(pc.SerialNumber)+1+len(pc.Purpose)],
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			_, err := UnmarshalBinary(tt.data)
			if err != ErrBinaryTruncated {
				t.Fatalf("expected ErrBinaryTruncated, got %v", err)
			}
		})
	}
}

func FuzzBinaryRoundtrip(f *testing.F) {
	f.Add([]byte(""), "ABC123", "Backup", "comment")
	f.Add(bytes.Repeat([]byte{0}, 100), "XYZ", "", "")
	f.Add(bytes.Repeat([]byte{0xFF}, 1000), "TM001", "Long purpose text", "A comment")

	f.Fuzz(func(t *testing.T, data []byte, serial, purpose, comment string) {
		if len(serial) > 255 || len(purpose) > 255 || len(comment) > 255 {
			t.Skip("field too long")
		}

		pc := &PaperCrypt{
			Version:      "3.0.0",
			DataFormat:   PaperCryptDataFormatRaw,
			SerialNumber: serial,
			Purpose:      purpose,
			Comment:      comment,
			CreatedAt:    time.Now(),
			Data:         data,
		}
		pc.DataSHA256 = sha256.Sum256(data)

		bin, err := MarshalBinary(pc)
		if err != nil {
			t.Skipf("MarshalBinary: %v", err)
		}

		got, err := UnmarshalBinary(bin)
		if err != nil {
			t.Fatalf("UnmarshalBinary failed after MarshalBinary: %v", err)
		}

		if got.SerialNumber != serial {
			t.Errorf("SerialNumber: got %q, want %q", got.SerialNumber, serial)
		}
		if got.Purpose != purpose {
			t.Errorf("Purpose: got %q, want %q", got.Purpose, purpose)
		}
		if got.Comment != comment {
			t.Errorf("Comment: got %q, want %q", got.Comment, comment)
		}
		if !bytes.Equal(got.Data, data) {
			t.Errorf("Data mismatch")
		}
	})
}

func TestParseVersion(t *testing.T) {
	valid := []struct {
		input                     string
		wantMaj, wantMin, wantPat uint8
	}{
		{"v3.1.2", 3, 1, 2},
		{"3.1.2", 3, 1, 2},
		{"v0.0.0", 0, 0, 0},
		{"v255.255.255", 255, 255, 255},
	}
	for _, tt := range valid {
		maj, mi, pat, err := ParseVersion(tt.input)
		if err != nil {
			t.Errorf("ParseVersion(%q): unexpected error %v", tt.input, err)
		}
		if maj != tt.wantMaj || mi != tt.wantMin || pat != tt.wantPat {
			t.Errorf("ParseVersion(%q) = %d.%d.%d, want %d.%d.%d",
				tt.input, maj, mi, pat, tt.wantMaj, tt.wantMin, tt.wantPat)
		}
	}

	invalid := []string{"devel", "", "v1.2", "1.2", "300.0.0", "1.300.0", "1.0.300", "-1.0.0"}
	for _, input := range invalid {
		if _, _, _, err := ParseVersion(input); err == nil {
			t.Errorf("ParseVersion(%q): expected error", input)
		}
	}
}

func TestFormatVersion(t *testing.T) {
	got := formatVersion(3, 1, 2)
	if got != "3.1.2" {
		t.Errorf("formatVersion(3,1,2) = %q, want %q", got, "3.1.2")
	}
}

func TestParseFormatRoundtrip(t *testing.T) {
	for _, v := range []string{"1.0.0", "3.1.2", "0.0.0", "255.255.255"} {
		maj, mi, pat, err := ParseVersion(v)
		if err != nil {
			t.Fatalf("ParseVersion(%q): %v", v, err)
		}
		got := formatVersion(maj, mi, pat)
		if got != v {
			t.Errorf("roundtrip %q: parse -> format = %q", v, got)
		}
	}
}

func TestMarshalBinaryVersionValidation(t *testing.T) {
	base := &PaperCrypt{
		DataFormat: PaperCryptDataFormatRaw,
		CreatedAt:  time.Now(),
		Data:       []byte("x"),
	}

	for _, v := range []string{"3.1.2", "v3.0.0", "0.0.0", "255.255.255"} {
		pc := *base
		pc.Version = v
		if _, err := MarshalBinary(&pc); err != nil {
			t.Errorf("MarshalBinary with version %q: unexpected error %v", v, err)
		}
	}

	for _, v := range []string{"devel", "", "1.2", "abc", "300.0.0", "1.300.0", "1.0.300"} {
		pc := *base
		pc.Version = v
		if _, err := MarshalBinary(&pc); err == nil {
			t.Errorf("MarshalBinary with version %q: expected error", v)
		}
	}
}
