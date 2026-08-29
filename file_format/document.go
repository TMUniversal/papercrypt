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
	"crypto/sha256"
	"time"
)

type PaperCrypt struct {
	Version      string               `json:"v"`
	DataFormat   PaperCryptDataFormat `json:"f"`
	SerialNumber string               `json:"sn"`
	Purpose      string               `json:"p"`
	Comment      string               `json:"cm"`
	CreatedAt    time.Time            `json:"ct"`
	DataSHA256   [32]byte             `json:"-"`

	// Data is either ASCII armored OpenPGP data (DataFormat PGP, gzipped
	// before encryption) or raw bytes (DataFormat Raw). Either way, the
	// payload is gzipped after processing.
	Data []byte `json:"d"`
}

func NewPaperCrypt(
	version string,
	data []byte,
	serialNumber string,
	purpose string,
	comment string,
	createdAt time.Time,
	format PaperCryptDataFormat,
) *PaperCrypt {
	dataSHA256 := sha256.Sum256(data)

	return &PaperCrypt{
		Version:      version,
		Data:         data,
		SerialNumber: serialNumber,
		Purpose:      purpose,
		Comment:      comment,
		CreatedAt:    createdAt,
		DataSHA256:   dataSHA256,
		DataFormat:   format,
	}
}
