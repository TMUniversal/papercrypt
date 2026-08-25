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
	"crypto/sha256"
	"errors"
	"fmt"
	"time"
)

const (
	// BytesPerLine denominates the amount of bytes to be encoded per line of the serialized output
	BytesPerLine = 24
)

const (
	// HeaderFieldVersion holds the name of the header field Version. Constant to avoid parsing issues.
	HeaderFieldVersion = "PaperCrypt Version"
	// HeaderFieldSerial holds the name of the header field for the serial number. Constant to avoid parsing issues.
	HeaderFieldSerial = "Content Serial"
	// HeaderFieldPurpose holds the name of the header field Purpose. Constant to avoid parsing issues.
	HeaderFieldPurpose = "Purpose"
	// HeaderFieldComment holds the name of the header field Comment. Constant to avoid parsing issues.
	HeaderFieldComment = "Comment"
	// HeaderFieldDate holds the name of the header field Date. Constant to avoid parsing issues.
	HeaderFieldDate = "Date"
	// HeaderFieldDataFormat holds the name of the header field Data Format. Constant to avoid parsing issues.
	HeaderFieldDataFormat = "Data Format"
	// HeaderFieldContentLength holds the name of the header field Content Length. Constant to avoid parsing issues.
	HeaderFieldContentLength = "Content Length"
	// HeaderFieldSHA256 holds the name of the header field for the SHA-256 checksum. Constant to avoid parsing issues.
	HeaderFieldSHA256 = "Content SHA-256"
	// HeaderFieldHeaderCRC32 holds the name of the header field for the CRC-32 checksum of the header. Constant to avoid parsing issues.
	HeaderFieldHeaderCRC32 = "Header CRC-32"
)

var (
	errorParsingHeader     = errors.New("error parsing header")
	errorParsingBody       = errors.New("error parsing body")
	errorValidationFailure = errors.New("validation failure")
)

// PaperCrypt represents a PaperCrypt document.
// It contains metadata about the document, such as its version, serial number, purpose, comment, creation date, and the data itself.
type PaperCrypt struct {
	// Version is the version of papercrypt used to generate the document.
	Version string `json:"v"`

	// DataFormat determines whether the data is raw (uncompressed, unencrypted), or follows the PGP message format (encrypted and gzipped).
	DataFormat PaperCryptDataFormat `json:"f"`

	// SerialNumber is the serial number of document, used to identify it. It is generated randomly if not provided.
	SerialNumber string `json:"sn"`

	// Purpose is the purpose of document
	Purpose string `json:"p"`

	// Comment is the comment on document
	Comment string `json:"cm"`

	// CreatedAt is the creation timestamp
	CreatedAt time.Time `json:"ct"`

	// DataSHA256 is the SHA-256 checksum of the encrypted data
	DataSHA256 [32]byte `json:"-"`

	// Data is the contents of the document
	// it can be either of two formats:
	//   a) ASCII armored OpenPGP data, if DataFormat is PGP
	//      the contained message is gzipped before encryption
	//   b) Raw data of any kind, if DataFormat is Raw
	// either way, data is always gzipped after processing
	Data []byte `json:"d"`
}

// NewPaperCrypt creates a new paper crypt.
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

// GetBinarySerialized returns the binary serialized representation of the PaperCrypt document as a string.
func (p *PaperCrypt) GetBinarySerialized() (string, error) {
	if p.Data == nil {
		return "", errors.New("no data to serialize")
	}

	if len(p.Data) == 0 {
		return "", errors.New("no data to serialize")
	}

	return SerializeBinary(&p.Data, BytesPerLine), nil
}

// GetDataLength returns the length of the data in bytes as an integer.
func (p *PaperCrypt) GetDataLength() int {
	return len(p.Data)
}

func newFieldNotPresentError(field string) error {
	return fmt.Errorf("`%s` not present in header", field)
}
