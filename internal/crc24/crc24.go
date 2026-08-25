/*
 * This file is part of PaperCrypt.
 *
 * PaperCrypt lets you prepare encrypted messages for printing on paper.
 * Copyright (C) 2023-2026 TMUniversal <me@tmuniversal.eu>.
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

// Package crc24 implements CRC-24 checksums using the OpenPGP/RTCM104v3 polynomial.
package crc24

const (
	// polynomial defines the CRC-24 polynomial used in OpenPGP and RTCM104v3.
	polynomial = uint32(0x864CFB)
	// initial is the initial value for CRC-24 calculations.
	initial   = uint32(0xB704CE)
	tableSize = uint32(256)
)

var table [tableSize]uint32

func init() {
	for i := range tableSize {
		crc := i << 16
		for range 8 {
			if (crc & 0x800000) != 0 {
				crc = (crc << 1) ^ polynomial
			} else {
				crc <<= 1
			}
		}
		table[i] = crc & 0xFFFFFF
	}
}

// Checksum generates a CRC-24 checksum for the given data.
func Checksum(data []byte) uint32 {
	crc := initial
	for _, b := range data {
		index := byte(crc>>16) ^ b
		crc = (crc << 8) ^ table[index]
	}
	return crc & 0xFFFFFF
}

// Validate checks data against a provided CRC-24 checksum.
func Validate(data []byte, checksum uint32) bool {
	return Checksum(data) == checksum
}
