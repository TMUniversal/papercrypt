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
	"strings"
	"testing"
)

func TestGenerateSerialLength(t *testing.T) {
	const alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZ234567"
	for _, length := range []uint8{1, 3, 6, 12} {
		serial, err := GenerateSerial(length)
		if err != nil {
			t.Fatalf("GenerateSerial(%d) failed with error %s", length, err)
		}
		if len(serial) != int(length) {
			t.Errorf(
				"GenerateSerial(%d) returned %d characters, want %d",
				length,
				len(serial),
				length,
			)
		}
		for _, r := range serial {
			if !strings.ContainsRune(alphabet, r) {
				t.Errorf("GenerateSerial(%d) returned non-base32 character %q", length, r)
				break
			}
		}
	}
}
