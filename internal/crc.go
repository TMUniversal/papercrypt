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

	"github.com/tmuniversal/papercrypt/v3/internal/crc24"
)

const (
	// CRC24Polynomial is the CRC-24 polynomial used by PaperCrypt.
	CRC24Polynomial = crc24.Polynomial
	// CRC24Initial is the initial value for CRC-24 computation.
	CRC24Initial = crc24.Initial
)

// Crc24Checksum generates a CRC-24 checksum for the given data.
func Crc24Checksum(data []byte) uint32 {
	return crc24.Checksum(data)
}

// ValidateCRC24 validates the CRC-24 checksum of the given data against the provided checksum.
func ValidateCRC24(data []byte, checksum uint32) bool {
	return crc24.Validate(data, checksum)
}

// ValidateCRC32 reports whether the IEEE CRC-32 checksum of data matches checksum.
func ValidateCRC32(data []byte, checksum uint32) bool {
	return crc32.ChecksumIEEE(data) == checksum
}
