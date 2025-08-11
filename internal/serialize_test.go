/*
 * This file is part of PaperCrypt.
 *
 * PaperCrypt lets you prepare encrypted messages for printing on paper.
 * Copyright (C) 2023-2024 TMUniversal <me@tmuniversal.eu>.
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

package internal

import (
	"bytes"
	"encoding/base64"
	"testing"
)

func TestParseHexUint32(t *testing.T) {
	t.Run("parse hex to unit32", func(t *testing.T) {
		hex := "0x40"
		parsed, err := ParseHexUint32(hex)
		if err != nil {
			t.Errorf("ParseHexUint32 failed with error %s", err)
		}

		// 0x40 in decimal is 64
		if parsed != 64 {
			t.Errorf("Parsed value was incorrect, got: %d, want: %d.", parsed, 64)
		}
	})

	t.Run("parse invalid hex", func(t *testing.T) {
		hex := "0xg"
		_, err := ParseHexUint32(hex)

		if err == nil {
			t.Errorf("ParseHexUint32 should fail with invalid hex")
		}
	})

	t.Run("parse large hex number that exceeds uint32", func(t *testing.T) {
		hex := "0xFFFFFFFFF"
		_, err := ParseHexUint32(hex)

		if err == nil {
			t.Errorf("ParseHexUint32 should fail with hex number that exceeds uint32")
		}
	})

	t.Run("parse hex number without prefix", func(t *testing.T) {
		hex := "FF"
		_, err := ParseHexUint32(hex)
		if err != nil {
			t.Errorf("ParseHexUint32 should not fail with hex number without prefix")
		}
	})
}

func TestBytesFromBase64(t *testing.T) {
	example := "VR/3qgEcL9O8CDz95xLK0PznmkF9cncMcfnJOnlPSDk="

	t.Run("normal base64 decoding", func(t *testing.T) {
		encoding := base64.StdEncoding.EncodeToString([]byte(example))
		decoded, _ := BytesFromBase64(encoding)

		if string(decoded) != example {
			t.Errorf("Decoding was incorrect, got: %s, want: %s.", decoded, example)
		}
	})

	t.Run("decoding of an empty string", func(t *testing.T) {
		decoded, _ := BytesFromBase64("")
		if string(decoded) != "" {
			t.Errorf("Decoding was incorrect, got: %s, want: %s.", decoded, "")
		}
	})

	t.Run("decoding of a string that is not correctly base64 encoded", func(t *testing.T) {
		_, err := BytesFromBase64("@@@")

		if err == nil {
			t.Errorf("Error should be thrown for incorrect base64 string")
		}
	})
}

func TestDeserializeBinary(t *testing.T) {
	correctFile := ` 1: C3 2E 04 09 03 08 B6 92 73 1C A2 AF D9 5E E0 23 D8 A9 30 70 01 47 E3940B
	 2: 99 61 57 AE 0F C4 EB 77 3A 2D 1D 7B 41 3D 7B 2C 79 B2 49 9D 47 10 59F9F5
	 3: 19 80 D4 1E D2 C0 7B 01 3E BC 67 E6 A6 BB CC 3F 81 B6 00 0B 28 E2 782D20
	 4: D2 C7 47 C3 88 4F 08 BA 3D 35 4C C3 55 9C 0E 6D 33 92 F3 A2 C6 8C FF7639
	 5: 53 4B 57 7B 1B 7B 07 1F 10 D2 36 45 C6 68 AF E0 16 C6 DE B6 61 57 2F42DE
	 6: 5A DD 27 B9 B8 D2 6F A6 C8 0C 6C 03 87 E2 E7 B7 01 90 F0 27 95 46 A195FE
	 7: C2 DB 53 63 0E 9D 8F 5A B9 B3 52 C2 D4 5D 2D 05 BD FB 84 80 7D 56 2974C5
	 8: 4F 0E 7D 9E 22 F4 2A 82 53 36 B0 D1 C3 94 D7 1D 63 62 DF 6D 56 A1 A01C74
	 9: 78 8A B8 82 60 40 07 B1 20 A5 25 7F EC 56 ED 15 D6 41 E3 1E 93 16 071B96
	10: 49 BB 14 26 A2 B1 0B 17 7A 60 E7 58 41 55 D0 E3 02 A6 ED EC E0 4B 28C54C
	11: F2 D2 E8 86 9B 1D 31 58 69 FF 6C B9 6D DB 87 83 F1 18 5B DE E5 4C CC5CFC
	12: DE 30 46 EA B6 70 1D 83 39 31 77 E9 E7 EC 53 92 DD 36 9F 32 C8 6A 42A899
	13: FF 2F E4 E2 2F FA 17 8E F9 34 DA A1 EF B8 72 A7 32 18 65 24 56 E8 B31D34
	14: 39 71 E4 53 64 63 83 E2 07 33 3D 43 1D ED 07 9C 62 FE 24 44 53 22 38BBFD
	15: E7 68 FC 9B 3C E4 3E ED 26 44 3C BC 2E 66 27 FD EE 7B 0D 0B EB 78 B1295E
	16: 8D 88 F2 7B 93 A6 F1 50 AC 96 62 94 B7 16 8E 4A 33 F9 A5 93 D9 31 FE9BCC
	17: 74 98 9A 8D A9 54 DD 16 C5 34 FA 42 F2 4C 3278B5
	18: 22DF5F`

	t.Run("deserialize binary", func(t *testing.T) {
		data := []byte(correctFile)
		_, err := DeserializeBinary(&data)
		if err != nil {
			t.Errorf("DeserializeBinary failed with error %s", err)
		}
	})

	t.Run("deserialize binary with expected result", func(t *testing.T) {
		data := []byte(correctFile)
		res, err := DeserializeBinary(&data)
		if err != nil {
			t.Errorf("DeserializeBinary failed with error %s", err)
		}

		expected := []byte{
			0xc3,
			0x2e,
			0x04,
			0x09,
			0x03,
			0x08,
			0xb6,
			0x92,
			0x73,
			0x1c,
			0xa2,
			0xaf,
			0xd9,
			0x5e,
			0xe0,
			0x23,
			0xd8,
			0xa9,
			0x30,
			0x70,
			0x01,
			0x47,
			0x99,
			0x61,
			0x57,
			0xae,
			0x0f,
			0xc4,
			0xeb,
			0x77,
			0x3a,
			0x2d,
			0x1d,
			0x7b,
			0x41,
			0x3d,
			0x7b,
			0x2c,
			0x79,
			0xb2,
			0x49,
			0x9d,
			0x47,
			0x10,
			0x19,
			0x80,
			0xd4,
			0x1e,
			0xd2,
			0xc0,
			0x7b,
			0x01,
			0x3e,
			0xbc,
			0x67,
			0xe6,
			0xa6,
			0xbb,
			0xcc,
			0x3f,
			0x81,
			0xb6,
			0x00,
			0x0b,
			0x28,
			0xe2,
			0xd2,
			0xc7,
			0x47,
			0xc3,
			0x88,
			0x4f,
			0x08,
			0xba,
			0x3d,
			0x35,
			0x4c,
			0xc3,
			0x55,
			0x9c,
			0x0e,
			0x6d,
			0x33,
			0x92,
			0xf3,
			0xa2,
			0xc6,
			0x8c,
			0x53,
			0x4b,
			0x57,
			0x7b,
			0x1b,
			0x7b,
			0x07,
			0x1f,
			0x10,
			0xd2,
			0x36,
			0x45,
			0xc6,
			0x68,
			0xaf,
			0xe0,
			0x16,
			0xc6,
			0xde,
			0xb6,
			0x61,
			0x57,
			0x5a,
			0xdd,
			0x27,
			0xb9,
			0xb8,
			0xd2,
			0x6f,
			0xa6,
			0xc8,
			0x0c,
			0x6c,
			0x03,
			0x87,
			0xe2,
			0xe7,
			0xb7,
			0x01,
			0x90,
			0xf0,
			0x27,
			0x95,
			0x46,
			0xc2,
			0xdb,
			0x53,
			0x63,
			0x0e,
			0x9d,
			0x8f,
			0x5a,
			0xb9,
			0xb3,
			0x52,
			0xc2,
			0xd4,
			0x5d,
			0x2d,
			0x05,
			0xbd,
			0xfb,
			0x84,
			0x80,
			0x7d,
			0x56,
			0x4f,
			0x0e,
			0x7d,
			0x9e,
			0x22,
			0xf4,
			0x2a,
			0x82,
			0x53,
			0x36,
			0xb0,
			0xd1,
			0xc3,
			0x94,
			0xd7,
			0x1d,
			0x63,
			0x62,
			0xdf,
			0x6d,
			0x56,
			0xa1,
			0x78,
			0x8a,
			0xb8,
			0x82,
			0x60,
			0x40,
			0x07,
			0xb1,
			0x20,
			0xa5,
			0x25,
			0x7f,
			0xec,
			0x56,
			0xed,
			0x15,
			0xd6,
			0x41,
			0xe3,
			0x1e,
			0x93,
			0x16,
			0x49,
			0xbb,
			0x14,
			0x26,
			0xa2,
			0xb1,
			0x0b,
			0x17,
			0x7a,
			0x60,
			0xe7,
			0x58,
			0x41,
			0x55,
			0xd0,
			0xe3,
			0x02,
			0xa6,
			0xed,
			0xec,
			0xe0,
			0x4b,
			0xf2,
			0xd2,
			0xe8,
			0x86,
			0x9b,
			0x1d,
			0x31,
			0x58,
			0x69,
			0xff,
			0x6c,
			0xb9,
			0x6d,
			0xdb,
			0x87,
			0x83,
			0xf1,
			0x18,
			0x5b,
			0xde,
			0xe5,
			0x4c,
			0xde,
			0x30,
			0x46,
			0xea,
			0xb6,
			0x70,
			0x1d,
			0x83,
			0x39,
			0x31,
			0x77,
			0xe9,
			0xe7,
			0xec,
			0x53,
			0x92,
			0xdd,
			0x36,
			0x9f,
			0x32,
			0xc8,
			0x6a,
			0xff,
			0x2f,
			0xe4,
			0xe2,
			0x2f,
			0xfa,
			0x17,
			0x8e,
			0xf9,
			0x34,
			0xda,
			0xa1,
			0xef,
			0xb8,
			0x72,
			0xa7,
			0x32,
			0x18,
			0x65,
			0x24,
			0x56,
			0xe8,
			0x39,
			0x71,
			0xe4,
			0x53,
			0x64,
			0x63,
			0x83,
			0xe2,
			0x07,
			0x33,
			0x3d,
			0x43,
			0x1d,
			0xed,
			0x07,
			0x9c,
			0x62,
			0xfe,
			0x24,
			0x44,
			0x53,
			0x22,
			0xe7,
			0x68,
			0xfc,
			0x9b,
			0x3c,
			0xe4,
			0x3e,
			0xed,
			0x26,
			0x44,
			0x3c,
			0xbc,
			0x2e,
			0x66,
			0x27,
			0xfd,
			0xee,
			0x7b,
			0x0d,
			0x0b,
			0xeb,
			0x78,
			0x8d,
			0x88,
			0xf2,
			0x7b,
			0x93,
			0xa6,
			0xf1,
			0x50,
			0xac,
			0x96,
			0x62,
			0x94,
			0xb7,
			0x16,
			0x8e,
			0x4a,
			0x33,
			0xf9,
			0xa5,
			0x93,
			0xd9,
			0x31,
			0x74,
			0x98,
			0x9a,
			0x8d,
			0xa9,
			0x54,
			0xdd,
			0x16,
			0xc5,
			0x34,
			0xfa,
			0x42,
			0xf2,
			0x4c,
		}

		if !bytes.Equal(res, expected) {
			t.Errorf("Deserialized value was incorrect, got: %x, want: %x.", res, expected)
		}
	})

	t.Run("deserialize binary with wrong block hash", func(t *testing.T) {
		wrongBlockHash := correctFile[:len(correctFile)-1] + "0"
		data := []byte(wrongBlockHash)
		_, err := DeserializeBinary(&data)
		if err == nil {
			t.Errorf("DeserializeBinary should fail with wrong block hash")
		}
	})

	t.Run("deserialize binary with wrong line hash", func(t *testing.T) {
		wrongLineHash := []byte(correctFile)
		wrongLineHash[5] = '0'
		_, err := DeserializeBinary(&wrongLineHash)
		if err == nil {
			t.Errorf("DeserializeBinary should fail with wrong line hash")
		}
	})

	t.Run("deserialize binary with invalid base16", func(t *testing.T) {
		data := []byte(` 1: G3 2E 04 09 03 08 B6 92 73 1C A2 AF D9 5E E0 23 D8 A9 30 70 01 47 E3940B
	 2: 99 61 57 AE 0F C4 EB 77 3A 2D 1D 7B 41 3D 7B 2C 79 B2 49 9D 47 10 59F9F5
	 3: 19 80 D4 1E D2 C0 7B 01 3E BC 67 E6 A6 BB CC 3F 81 B6 00 0B 28 E2 782D20
	 4: D2 C7 47 C3 88 4F 08 BA 3D 35 4C C3 55 9C 0E 6D 33 92 F3 A2 C6 8C FF7639
	 5: 53 4B 57 7B 1B 7B 07 1F 10 D2 36 45 C6 68 AF E0 16 C6 DE B6 61 57 2F42DE
	 6: 5A DD 27 B9 B8 D2 6F A6 C8 0C 6C 03 87 E2 E7 B7 01 90 F0 27 95 46 A195FE
	 7: C2 DB 53 63 0E 9D 8F 5A B9 B3 52 C2 D4 5D 2D 05 BD FB 84 80 7D 56 2974C5
	 8: 4F 0E 7D 9E 22 F4 2A 82 53 36 B0 D1 C3 94 D7 1D 63 62 DF 6D 56 A1 A01C74
	 9: 78 8A B8 82 60 40 07 B1 20 A5 25 7F EC 56 ED 15 D6 41 E3 1E 93 16 071B96
	10: 49 BB 14 26 A2 B1 0B 17 7A 60 E7 58 41 55 D0 E3 02 A6 ED EC E0 4B 28C54C
	11: F2 D2 E8 86 9B 1D 31 58 69 FF 6C B9 6D DB 87 83 F1 18 5B DE E5 4C CC5CFC
	12: DE 30 46 EA B6 70 1D 83 39 31 77 E9 E7 EC 53 92 DD 36 9F 32 C8 6A 42A899
	13: FF 2F E4 E2 2F FA 17 8E F9 34 DA A1 EF B8 72 A7 32 18 65 24 56 E8 B31D34
	14: 39 71 E4 53 64 63 83 E2 07 33 3D 43 1D ED 07 9C 62 FE 24 44 53 22 38BBFD
	15: E7 68 FC 9B 3C E4 3E ED 26 44 3C BC 2E 66 27 FD EE 7B 0D 0B EB 78 B1295E
	16: 8D 88 F2 7B 93 A6 F1 50 AC 96 62 94 B7 16 8E 4A 33 F9 A5 93 D9 31 FE9BCC
	17: 74 98 9A 8D A9 54 DD 16 C5 34 FA 42 F2 4C 3278B5
	18: 22DF5F`)
		_, err := DeserializeBinary(&data)
		if err == nil {
			t.Errorf("DeserializeBinary should fail with invalid base16")
		}
	})

	t.Run("deserialize binary with invalid line numbers", func(t *testing.T) {
		data := []byte(` 1: C3 2E 04 09 03 08 B6 92 73 1C A2 AF D9 5E E0 23 D8 A9 30 70 01 47 E3940B
	 A: 99 61 57 AE 0F C4 EB 77 3A 2D 1D 7B 41 3D 7B 2C 79 B2 49 9D 47 10 59F9F5
	 3: 19 80 D4 1E D2 C0 7B 01 3E BC 67 E6 A6 BB CC 3F 81 B6 00 0B 28 E2 782D20
	 4: D2 C7 47 C3 88 4F 08 BA 3D 35 4C C3 55 9C 0E 6D 33 92 F3 A2 C6 8C FF7639
	 5: 53 4B 57 7B 1B 7B 07 1F 10 D2 36 45 C6 68 AF E0 16 C6 DE B6 61 57 2F42DE
	 6: 5A DD 27 B9 B8 D2 6F A6 C8 0C 6C 03 87 E2 E7 B7 01 90 F0 27 95 46 A195FE
	 7: C2 DB 53 63 0E 9D 8F 5A B9 B3 52 C2 D4 5D 2D 05 BD FB 84 80 7D 56 2974C5
	 8: 4F 0E 7D 9E 22 F4 2A 82 53 36 B0 D1 C3 94 D7 1D 63 62 DF 6D 56 A1 A01C74
	 9: 78 8A B8 82 60 40 07 B1 20 A5 25 7F EC 56 ED 15 D6 41 E3 1E 93 16 071B96
	10: 49 BB 14 26 A2 B1 0B 17 7A 60 E7 58 41 55 D0 E3 02 A6 ED EC E0 4B 28C54C
	11: F2 D2 E8 86 9B 1D 31 58 69 FF 6C B9 6D DB 87 83 F1 18 5B DE E5 4C CC5CFC
	12: DE 30 46 EA B6 70 1D 83 39 31 77 E9 E7 EC 53 92 DD 36 9F 32 C8 6A 42A899
	13: FF 2F E4 E2 2F FA 17 8E F9 34 DA A1 EF B8 72 A7 32 18 65 24 56 E8 B31D34
	14: 39 71 E4 53 64 63 83 E2 07 33 3D 43 1D ED 07 9C 62 FE 24 44 53 22 38BBFD
	15: E7 68 FC 9B 3C E4 3E ED 26 44 3C BC 2E 66 27 FD EE 7B 0D 0B EB 78 B1295E
	16: 8D 88 F2 7B 93 A6 F1 50 AC 96 62 94 B7 16 8E 4A 33 F9 A5 93 D9 31 FE9BCC
	17: 74 98 9A 8D A9 54 DD 16 C5 34 FA 42 F2 4C 3278B5
	18: 22DF5F`)
		_, err := DeserializeBinary(&data)
		if err == nil {
			t.Errorf("DeserializeBinary should fail with invalid base16")
		}
	})

	t.Run("deserialize binary with out-of-order lines", func(t *testing.T) {
		data := []byte(` 1: C3 2E 04 09 03 08 B6 92 73 1C A2 AF D9 5E E0 23 D8 A9 30 70 01 47 E3940B
	 2: 99 61 57 AE 0F C4 EB 77 3A 2D 1D 7B 41 3D 7B 2C 79 B2 49 9D 47 10 59F9F5
	 4: D2 C7 47 C3 88 4F 08 BA 3D 35 4C C3 55 9C 0E 6D 33 92 F3 A2 C6 8C FF7639
	 3: 19 80 D4 1E D2 C0 7B 01 3E BC 67 E6 A6 BB CC 3F 81 B6 00 0B 28 E2 782D20
	 5: 53 4B 57 7B 1B 7B 07 1F 10 D2 36 45 C6 68 AF E0 16 C6 DE B6 61 57 2F42DE
	 6: 5A DD 27 B9 B8 D2 6F A6 C8 0C 6C 03 87 E2 E7 B7 01 90 F0 27 95 46 A195FE
	 7: C2 DB 53 63 0E 9D 8F 5A B9 B3 52 C2 D4 5D 2D 05 BD FB 84 80 7D 56 2974C5
	10: 49 BB 14 26 A2 B1 0B 17 7A 60 E7 58 41 55 D0 E3 02 A6 ED EC E0 4B 28C54C
	 9: 78 8A B8 82 60 40 07 B1 20 A5 25 7F EC 56 ED 15 D6 41 E3 1E 93 16 071B96
	12: DE 30 46 EA B6 70 1D 83 39 31 77 E9 E7 EC 53 92 DD 36 9F 32 C8 6A 42A899
	11: F2 D2 E8 86 9B 1D 31 58 69 FF 6C B9 6D DB 87 83 F1 18 5B DE E5 4C CC5CFC
	13: FF 2F E4 E2 2F FA 17 8E F9 34 DA A1 EF B8 72 A7 32 18 65 24 56 E8 B31D34
	 8: 4F 0E 7D 9E 22 F4 2A 82 53 36 B0 D1 C3 94 D7 1D 63 62 DF 6D 56 A1 A01C74
	14: 39 71 E4 53 64 63 83 E2 07 33 3D 43 1D ED 07 9C 62 FE 24 44 53 22 38BBFD
	18: 22DF5F
	15: E7 68 FC 9B 3C E4 3E ED 26 44 3C BC 2E 66 27 FD EE 7B 0D 0B EB 78 B1295E
	16: 8D 88 F2 7B 93 A6 F1 50 AC 96 62 94 B7 16 8E 4A 33 F9 A5 93 D9 31 FE9BCC
	17: 74 98 9A 8D A9 54 DD 16 C5 34 FA 42 F2 4C 3278B5`)
		_, err := DeserializeBinary(&data)
		if err != nil {
			t.Errorf("DeserializeBinary should not fail with lines swapped")
		}
	})
}
