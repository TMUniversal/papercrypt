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
	"hash/crc32"
)

const (
	// CRC24Polynomial defines the polynomial used by the CRC-24 algorithm. See https://en.wikipedia.org/wiki/Cyclic_redundancy_check#:~:text=OpenPGP%2C%20RTCM104v3-,0x864cfb,-0xDF3261. This is the polynomial used in OpenPGP and RTCM104v3, represented as a uint32.
	CRC24Polynomial = uint32(0x864CFB) // CRC-24 polynomial
	// CRC24Initial is the initial value used for CRC-24 calculations. This is the initial value used in OpenPGP and RTCM104v3, represented as a uint32.
	CRC24Initial = uint32(0xB704CE) // Initial value
	// CRC24TableSize is the size of the CRC-24 table used for faster computation.
	CRC24TableSize = uint32(256) // Table size for faster computation
)

var crc24Table [CRC24TableSize]uint32

func generateCRCTable() {
	for i := uint32(0); i < CRC24TableSize; i++ {
		crc := i << 16
		for j := 0; j < 8; j++ {
			if (crc & 0x800000) != 0 {
				crc = (crc << 1) ^ CRC24Polynomial
			} else {
				crc <<= 1
			}
		}
		crc24Table[i] = crc & 0xFFFFFF
	}
}

// Crc24Checksum generates a CRC-24 checksum for the given data.
func Crc24Checksum(data []byte) uint32 {
	if crc24Table[0] == 0 {
		generateCRCTable()
	}

	crc := CRC24Initial

	for _, b := range data {
		index := byte(crc>>16) ^ b
		crc = (crc << 8) ^ crc24Table[index]
	}

	return crc & 0xFFFFFF
}

// ValidateCRC24 validates the CRC-24 checksum of the given data against the provided checksum.
// It returns true if the checksum matches, false otherwise.
func ValidateCRC24(data []byte, checksum uint32) bool {
	return Crc24Checksum(data) == checksum
}

// ValidateCRC32 validates the CRC-32 checksum of the given data against the provided checksum.
// It returns true if the checksum matches, false otherwise.
func ValidateCRC32(data []byte, checksum uint32) bool {
	return crc32.ChecksumIEEE(data) == checksum
}
