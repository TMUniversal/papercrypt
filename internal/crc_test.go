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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestChecksumValidation(t *testing.T) {
	data := []byte{0x12, 0x34, 0x56, 0x78, 0x9A, 0xBC, 0xDE, 0xF0}
	generateCRCTable()
	checksum := Crc24Checksum(data)

	assert.True(
		t,
		ValidateCRC24(data, checksum),
		"Expected determined checksum to be valid, but was not.",
	)
}

func TestChecksumInvalidation(t *testing.T) {
	data := []byte{0x12, 0x34, 0x56, 0x78, 0x9A, 0xBC, 0xDE, 0xF0}
	generateCRCTable()
	checksum := Crc24Checksum(data)

	// Modify the data to invalidate the checksum
	data[0] = 0xAB

	assert.False(
		t,
		ValidateCRC24(data, checksum),
		"Expected checksum validation to fail for changed data, but got true.",
	)
}

func TestBoth(t *testing.T) {
	data := []byte{
		0x2d,
		0x2d,
		0x2d,
		0x2d,
		0x2d,
		0x42,
		0x45,
		0x47,
		0x49,
		0x4e,
		0x20,
		0x50,
		0x47,
		0x50,
		0x20,
		0x4d,
		0x45,
		0x53,
		0x53,
		0x41,
		0x47,
		0x45,
	}
	generateCRCTable()
	checksum := uint32(0xc55238)

	assert.True(
		t,
		ValidateCRC24(data, checksum),
		"Expected checksum validation to be true for pre-determined valid checksum, but got false.",
	)
	valid := ValidateCRC24(data, checksum)
	if !valid {
		t.Errorf("Expected checksum validation to be true, but got false.")
	}
}

func TestValidateCRC32(t *testing.T) {
	data := []byte{
		0x2d,
		0x2d,
		0x2d,
		0x2d,
		0x2d,
		0x42,
		0x45,
		0x47,
		0x49,
		0x4e,
		0x20,
		0x50,
		0x47,
		0x50,
		0x20,
		0x4d,
		0x45,
		0x53,
		0x53,
		0x41,
		0x47,
		0x45,
	}
	checksum := uint32(0x59f08912)

	assert.True(
		t,
		ValidateCRC32(data, checksum),
		"Expected checksum validation to pass for pre-determined valid checksum, but got false.",
	)
}
