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

package file_format

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"time"

	"github.com/tmuniversal/papercrypt/v3/internal"
)

// JSONPaperCrypt is the JSON representation of PaperCrypt with base64 encoded hashes.
type JSONPaperCrypt struct {
	Version      string `json:"v"`
	DataFormat   string `json:"f"`
	SerialNumber string `json:"sn"`
	Purpose      string `json:"p,omitempty"`
	Comment      string `json:"cm,omitempty"`
	CreatedAt    string `json:"t"`
	DataSHA256   string `json:"s256"`
	Data         []byte `json:"d"`
}

// MarshalJSON implements the json.Marshaler interface for PaperCrypt.
func (p *PaperCrypt) MarshalJSON() ([]byte, error) {
	jpc := JSONPaperCrypt{
		Version:      p.Version,
		DataFormat:   p.DataFormat.String(),
		SerialNumber: p.SerialNumber,
		Purpose:      p.Purpose,
		Comment:      p.Comment,
		CreatedAt:    p.CreatedAt.Format(internal.TimeStampFormatJSON),
		DataSHA256:   base64.StdEncoding.EncodeToString(p.DataSHA256[:]),
		Data:         p.Data,
	}
	return json.Marshal(jpc)
}

// UnmarshalJSON implements the json.Unmarshaler interface for PaperCrypt.
func (p *PaperCrypt) UnmarshalJSON(data []byte) error {
	var jpc JSONPaperCrypt
	if err := json.Unmarshal(data, &jpc); err != nil {
		return err
	}

	createdAt, err := time.Parse(internal.TimeStampFormatJSON, jpc.CreatedAt)
	if err != nil {
		return err
	}
	p.CreatedAt = createdAt

	dataSHA256Bytes, err := base64.StdEncoding.DecodeString(jpc.DataSHA256)
	if err != nil {
		return err
	}
	if len(dataSHA256Bytes) != 32 {
		return errors.New("invalid DataSHA256 length")
	}
	copy(p.DataSHA256[:], dataSHA256Bytes)

	p.Version = jpc.Version
	p.DataFormat = PaperCryptDataFormatFromString(jpc.DataFormat)
	p.SerialNumber = jpc.SerialNumber
	p.Purpose = jpc.Purpose
	p.Comment = jpc.Comment
	p.Data = jpc.Data

	return nil
}
