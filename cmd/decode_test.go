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

const doc = `# PaperCrypt Version: devel
# Content Serial: KKJW6T
# Purpose: Example Sheet
# Comment: Regular PDF Example
# Date: Tue, 26 Sep 2023 21:54:48.118735000 CEST
# Content Length: 362
# Content CRC-24: bf7977
# Content CRC-32: 50981f7f
# Content SHA-256: AdsQVgMTzx0omoZyr+VpnC0Gh7qCqUELXS3mqlcgF7M=
# Header CRC-32: e9cc370d


 1: C3 2E 04 09 03 08 EB 13 B0 9B E9 94 C6 9A E0 93 4E DE 30 1F E8 F1 6AE308
 2: 86 B7 EA C1 28 07 A7 54 35 FF DA 9E 35 5B E3 C6 76 BE 7F A1 1D CE A81873
 3: EF B6 67 7B D2 C0 77 01 54 2C 0E 92 C7 55 8B 77 3E F1 E0 74 39 11 C7C98A
 4: 87 39 33 54 68 19 66 DF 1D 2C E7 C6 42 B0 4A F9 87 40 01 6D 45 0F A2E1A6
 5: 8D B2 F8 34 75 D5 D1 BA F5 69 41 88 F9 A1 33 F2 FC 7E 5B CA 3B 72 A8474B
 6: 52 43 B3 02 3A BC D2 26 75 4B 85 05 2E A5 27 B8 8A CF 9E 68 A3 13 C69503
 7: 0F 56 3F 64 20 53 3A 93 F0 2B 6D 0E 5D 0E 0C 6F 36 D5 97 96 39 BB 4BB28B
 8: 20 60 7F 7A 53 EC 67 1B EE DF 4B 1E C9 60 28 AC 52 04 4C 7A 40 C6 44F349
 9: DB E1 90 7A 1A 4C 6A C0 52 A9 07 B9 09 1B 79 BF 04 EF 2E C3 D3 D1 60E7CE
10: 85 3A 4A DB 17 AB 45 29 44 07 A4 E2 85 B1 72 F2 20 8C 4F 59 37 BA 8714D7
11: A5 B6 09 B0 43 B1 DC AA 92 43 FF 8F 72 D1 AB F5 F8 E1 04 16 11 36 A8A02D
12: 0F 05 2C AC D3 2D 42 AA 79 BC E5 AA DA A1 09 33 CA 2A 8C A0 AB 67 D11164
13: 68 2E 85 1B A1 36 35 AF 56 BB D3 88 43 EE 0F BD 95 65 58 1F E4 17 BFEAB0
14: 2A 49 9B FD 5D D1 77 84 DF F5 3F 79 3E 42 71 D6 3E 68 1D 39 05 00 9DBA89
15: 21 9C 63 5B 05 40 44 60 8A 7B 1E D3 B6 37 D2 F8 3A 1D 19 17 00 59 D4E3AE
16: DE 41 5E D5 53 82 82 F3 29 00 18 CF 35 76 60 8C 7A 3C 9A 78 2D 55 8A3EDE
17: 3C 04 B8 22 D1 74 9F 3C 4D CB 1B640E
18: BF7977`

var docBytes = []byte(doc)

func TestDecode(t *testing.T) {
	log.SetLevel(log.DebugLevel)

	tempDir := t.TempDir()
	inPath := tempDir + "/input.txt"
	outPath := tempDir + "/output.json"

	cmd := rootCmd
	cmd.SetArgs([]string{"decode", "-v", "-i", inPath, "-o", outPath, "-P", "example"})

	// cmd.PersistentFlags().Set("input", inPath)
	// cmd.PersistentFlags().Set("output", outPath)

	if err := os.WriteFile(inPath, docBytes, 0o600); err != nil {
		t.Fatal(err)
	}

	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	out, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatal(err)
	}

	if string(out) != input {
		t.Fatalf("Expected %s, got %s", input, string(out))
	}
}
