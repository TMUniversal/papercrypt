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

package cmd

import (
	"os"
	"testing"

	"github.com/caarlos0/log"
	"github.com/stretchr/testify/assert"
)

const input = `{
  "your_backup": {
    "backup_location": "https://your-bucket.s3.amazonaws.com/your_backup.tar.gz",
    "access_key_id": "YOUR_S3_ACCESS_KEY_ID",
    "secret_key": "YOUR_S3_SECRET_KEY",
    "encryption_key": "your-backup-encryption-key",
    "another_property": "another_value",
    "a_number": 123,
    "a_boolean": true,
    "an_array": ["a", "b", "c"],
    "an_object": {
      "another_property": "another_value"
    }
  }
}
`

const doc = `# PaperCrypt Version: 2.0.0
# Content Serial: EIPESR
# Purpose: Example Sheet
# Comment: Regular PDF Example
# Date: Thu, 01 Aug 2024 20:38:10.306596100 +0200
# Data Format: PGP
# Content Length: 390
# Content CRC-24: d6f1c0
# Content CRC-32: bc4b3672
# Content SHA-256: NT7wwW5Tq5fk1J82M1tzE82VGxIlad5vpF5cDMzg+yg=
# Header CRC-32: ecded03b


 1: 1F 8B 08 00 00 00 00 00 02 FF 00 6A 01 95 FE C3 2E 04 09 03 08 7A D49E51
 2: 7D 43 1C 18 E4 C9 19 E0 23 B0 2A D5 58 E1 72 93 E0 06 BB F2 7E C8 D183B9
 3: 2F C9 00 C8 90 6D 83 04 E9 22 FB 07 98 BB 4D 68 CE 04 96 D2 C0 77 730736
 4: 01 01 8D 41 D1 46 E3 82 11 09 E7 15 77 1C EB 92 26 FE 5A B2 84 C3 462812
 5: B4 98 DC D2 27 C1 B1 AF 22 B6 3B CB 95 DC D8 4D 0A 4E FF ED 8E A2 0B74A2
 6: DF C1 72 41 7F 08 AF 9C 43 EA 50 9C 43 30 84 4F F8 82 BC 62 4A 0E DFCF21
 7: 27 91 DF 15 9E 1C 3F 37 77 FB D2 E0 4A F1 73 3E 2D 7B 73 47 96 35 E94DF5
 8: 55 F9 A4 D2 7F 4C 24 4A 0B A1 04 1B 49 95 91 5C D0 6B E2 AF 2D AC 361154
 9: 98 E4 22 BB 62 61 BB 93 97 9A 04 4B 7B AC BF 86 7E 7B DE AB B3 83 9DD5D4
10: A9 66 F3 99 D7 94 2E 4E 72 E6 6D 09 35 11 68 A9 B7 6C EE 5E BC 3F E27CAE
11: E6 1B C7 5A 76 B0 B1 E5 DB 7A 56 13 23 DB 9C 23 8F 85 FF 72 60 56 252FD9
12: F4 26 17 EA 2E AE 05 D7 0F 02 78 A5 BE 3A 61 F0 39 EE 31 4F D6 E3 2ECC66
13: 7E 84 E6 99 D1 E7 71 CE 5D 34 6F A2 1D 66 74 1A 09 FC E2 81 91 AB 444B35
14: AC 88 A5 93 14 38 37 FD BA 49 5E B7 3F 33 55 D0 83 D8 2C 48 35 FF AE678F
15: 54 F5 38 85 EC C8 4E 37 03 B9 22 C7 50 58 7F BD 04 0C 8E EE 8B B9 B26C81
16: E4 6B C0 67 C6 18 54 0D F1 20 73 D8 FC 40 D3 D2 90 00 0F 84 7E BD C47477
17: 1D 67 D8 71 AD 2D D3 89 43 54 8A F5 33 CD 0E AF B0 80 08 29 68 59 E37012
18: 53 15 99 01 00 00 FF FF CC 08 E8 C8 6A 01 00 00 2436D9
19: D6F1C0`

const docRaw = `# PaperCrypt Version: 2.0.0
# Content Serial: BVC36O
# Purpose: Example Sheet
# Comment: Regular PDF Example
# Date: Sun, 04 Aug 2024 12:36:20.800974400 +0200
# Data Format: Raw
# Content Length: 261
# Content CRC-24: ce50b7
# Content CRC-32: 5f6b9b4b
# Content SHA-256: w8b3gibx3gdQsGXmlWVtET631gMT0TDLbe94IC5hXuw=
# Header CRC-32: 3967f39


 1: 1F 8B 08 00 00 00 00 00 02 FF 8C 90 3F 6B F3 30 10 C6 77 7F 0A A1 39 B6 C7FEC2
 2: 79 5F 6F DE 8A EB A1 74 28 C4 ED 10 4A 39 4E EA D1 A4 FE 23 73 92 5A 9C D4CD76
 3: 92 EF 5E CE C6 4E C6 0E 12 E8 B9 9F 1E 7E DC 4F A2 94 9E 5C 64 30 68 DB ADA862
 4: 38 EA 52 49 A4 94 5E DE D0 39 8B E1 E4 06 5D 2A 7D 0C 61 F4 65 9E 0B 9F 175C0A
 5: 9A 68 5B 0A 99 2F 32 EC F1 EC 06 FC F6 99 75 7D 7E 53 96 05 E4 EC E3 AC 260CAD
 6: 77 4B 23 5A 4B DE 43 4B 13 9C DE A5 EF F0 F4 B2 87 A6 80 BB AA AA 9B 06 F96B7B
 7: 1E EB 03 3C DC AF B4 27 CB 14 84 BE 45 9B BA DA D7 CF 82 AE 1C 0D 96 A7 4BBC58
 8: 51 14 57 76 D1 9B 0D D2 EB 34 95 E9 6A 32 B8 70 24 86 91 DD 48 1C E6 5F F4B766
 9: 6B F6 85 5D A4 0D 84 21 F6 86 58 97 EA DF FF 62 0B 8D 73 1D A1 EC 24 70 C33342
10: A4 AD 14 90 19 A5 EC 55 A3 DE 29 6D E4 B2 FA ED 0A 38 F3 49 36 6C 4B FE 285F92
11: 93 CA 4C 5E 12 39 97 E4 37 00 00 FF FF BD EA 20 F9 B0 01 00 00 BA6594
12: CE50B7`

func TestDecode(t *testing.T) {
	log.SetLevel(log.DebugLevel)

	tempDir := t.TempDir()
	inPath := tempDir + "/input.txt"
	outPath := tempDir + "/output.json"

	cmd := rootCmd
	cmd.SetArgs([]string{"decode", "-v", "-i", inPath, "-o", outPath, "-P", "example"})

	if err := os.WriteFile(inPath, []byte(doc), 0o600); err != nil {
		t.Fatal(err)
	}

	assert.NoError(t, cmd.Execute(), "command execution failed")

	out, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(
		t,
		input,
		string(out),
		"expected output to match input. Expected: %s, got: %s",
		input,
		string(out),
	)
}

func TestDecodeRaw(t *testing.T) {
	log.SetLevel(log.DebugLevel)

	tempDir := t.TempDir()
	inPath := tempDir + "/input.txt"
	outPath := tempDir + "/output.json"

	cmd := rootCmd
	cmd.SetArgs([]string{"decode", "-v", "-i", inPath, "-o", outPath, "-P", "example"})

	if err := os.WriteFile(inPath, []byte(docRaw), 0o600); err != nil {
		t.Fatal(err)
	}

	assert.NoError(t, cmd.Execute(), "command execution failed")

	out, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(
		t,
		input,
		string(out),
		"expected output to match input. Expected: %s, got: %s",
		input,
		string(out),
	)
}
