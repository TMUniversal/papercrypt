/*
 * This file is part of PaperCrypt.
 *
 * PaperCrypt lets you prepare encrypted messages for printing on paper.
 * Copyright (C) 2024 TMUniversal <me@tmuniversal.eu>.
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
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"hash/crc32"
	"strings"
	"time"

	"github.com/caarlos0/log"

	"github.com/ProtonMail/gopenpgp/v2/crypto"
)

type PaperCryptV1 struct {
	// Version is the version of papercrypt used to generate the document.
	Version string `json:"Version"`

	// Data is the encrypted data as a PGP message
	Data *crypto.PGPMessage `json:"Data"`

	// SerialNumber is the serial number of document, used to identify it. It is generated randomly if not provided.
	SerialNumber string `json:"SerialNumber"`

	// Purpose is the purpose of document
	Purpose string `json:"Purpose"`

	// Comment is the comment on document
	Comment string `json:"Comment"`

	// CreatedAt is the creation timestamp
	CreatedAt time.Time `json:"CreatedAt"`

	// DataCRC24 is the CRC-24 checksum of the encrypted data
	DataCRC24 uint32 `json:"DataCRC24"`

	// DataCRC32 is the CRC-32 checksum of the encrypted data
	DataCRC32 uint32 `json:"DataCRC32"`

	// DataSHA256 is the SHA-256 checksum of the encrypted data
	DataSHA256 [32]byte `json:"DataSHA256"`
}

// NewPaperCryptV1 creates a new paper crypt
func NewPaperCryptV1(version string, data *crypto.PGPMessage, serialNumber string, purpose string, comment string, createdAt time.Time) *PaperCryptV1 {
	binData := data.GetBinary()

	dataCRC24 := Crc24Checksum(binData)
	dataCRC32 := crc32.ChecksumIEEE(binData)
	dataSHA256 := sha256.Sum256(binData)

	return &PaperCryptV1{
		Version:      version,
		Data:         data,
		SerialNumber: serialNumber,
		Purpose:      purpose,
		Comment:      comment,
		CreatedAt:    createdAt,
		DataCRC24:    dataCRC24,
		DataCRC32:    dataCRC32,
		DataSHA256:   dataSHA256,
	}
}

// ToNextVersion converts the PaperCryptV1 to the next version,
// PaperCrypt,
// by encoding format information and compressing the data.
func (p *PaperCryptV1) ToNextVersion() (*PaperCrypt, error) {
	data := new(bytes.Buffer)
	gzipWriter := gzip.NewWriter(data)
	if _, err := gzipWriter.Write(p.Data.GetBinary()); err != nil {
		return nil, errors.Join(errors.New("error compressing data"), err)
	}
	if err := gzipWriter.Close(); err != nil {
		return nil, errors.Join(errors.New("error closing gzip writer"), err)
	}

	return &PaperCrypt{
		Version:      p.Version,
		Data:         data.Bytes(),
		SerialNumber: p.SerialNumber,
		Purpose:      p.Purpose,
		Comment:      p.Comment,
		CreatedAt:    p.CreatedAt,
		DataCRC24:    p.DataCRC24,
		DataCRC32:    p.DataCRC32,
		DataSHA256:   p.DataSHA256,
		DataFormat:   PaperCryptDataFormatPGP,
	}, nil
}

func (p *PaperCryptV1) GetBinary() []byte {
	return p.Data.GetBinary()
}

func (p *PaperCryptV1) GetBinarySerialized() string {
	data := p.GetBinary()
	return SerializeBinary(&data)
}

func (p *PaperCryptV1) GetLength() int {
	return len(p.GetBinary())
}

func (p *PaperCryptV1) GetText(lowerCaseEncoding bool) ([]byte, error) {
	header := fmt.Sprintf(
		`%s: %s
%s: %s
%s: %s
%s: %s
%s: %s
%s: %d
%s: %06x
%s: %08x
%s: %s`,
		HeaderFieldVersion,
		p.Version,
		HeaderFieldSerial,
		p.SerialNumber,
		HeaderFieldPurpose,
		p.Purpose,
		HeaderFieldComment,
		p.Comment,
		HeaderFieldDate,
		// format time with nanosecond precision
		// Sat, 12 Aug 2023 17:33:20.123456789
		p.CreatedAt.Format("Mon, 02 Jan 2006 15:04:05.000000000 MST"),
		HeaderFieldContentLength,
		p.GetLength(),
		HeaderFieldCRC24,
		p.DataCRC24,
		HeaderFieldCRC32,
		p.DataCRC32,
		HeaderFieldSHA256,
		base64.StdEncoding.EncodeToString(p.DataSHA256[:]))

	headerCRC32 := crc32.ChecksumIEEE([]byte(header))

	serializedData := p.GetBinarySerialized()
	if lowerCaseEncoding {
		serializedData = strings.ToLower(serializedData)
	}

	return []byte(
		fmt.Sprintf(`%s
%s: %08x


%s
`,
			header,
			HeaderFieldHeaderCRC32,
			headerCRC32,
			serializedData)), nil
}

func DeserializeV1Text(data []byte, ignoreVersionMismatch bool, ignoreChecksumMismatch bool) (*PaperCrypt, error) {
	paperCryptFileContents := NormalizeLineEndings(data)

	paperCryptFileContentsSplit := bytes.SplitN(paperCryptFileContents, []byte("\n\n\n"), 2)

	// 3. Read Headers if present
	if len(paperCryptFileContentsSplit) != 2 {
		return nil, errors.Join(errorParsingHeader, errors.New("header not discernible, header and content should be separated by two empty lines"))
	}

	headers, err := TextToHeaderMap(paperCryptFileContentsSplit[0])
	if err != nil {
		return nil, errors.Join(errorParsingHeader, err)
	}

	// Debug: print headers
	log.WithField("headers", headers).Debug("Read headers")

	// 4. Run Header Validation
	versionLine, ok := headers[HeaderFieldVersion]
	if !ok {
		if !ignoreVersionMismatch {
			return nil, errors.Join(errorParsingHeader, newFieldNotPresentError(HeaderFieldVersion))
		}

		log.Warn(Warning("PaperCrypt Version not present in header."))
	}

	// parse git-describe version, look for major version <= 1
	// releases are tagged as vX.Y.Z
	majorVersion := strings.Split(versionLine, ".")[0]
	majorVersion = strings.TrimPrefix(majorVersion, "v")
	if !ignoreVersionMismatch && !(majorVersion == "2" || majorVersion == "1" || majorVersion == "devel") {
		return nil, errors.Join(errorParsingHeader, fmt.Errorf("unsupported PaperCrypt version '%s'", versionLine))
	}

	// Validate Header checksum
	{
		headerCrc, ok := headers[HeaderFieldHeaderCRC32]
		if !ok {
			if !ignoreChecksumMismatch {
				return nil, errors.Join(errorParsingHeader, newFieldNotPresentError(HeaderFieldHeaderCRC32))
			}

			log.Warn(Warning("Header CRC-32 not present in header"))
		}

		headerCrc = strings.ToLower(headerCrc)
		headerCrc = strings.ReplaceAll(headerCrc, "0x", "")
		headerCrc = strings.ReplaceAll(headerCrc, " ", "")
		headerCrc32, err := ParseHexUint32(headerCrc)
		if err != nil {
			return nil, errors.Join(errorParsingHeader, errors.New("invalid CRC-32 format"), err)
		}

		headerWithoutCrc := bytes.ReplaceAll(paperCryptFileContentsSplit[0], []byte("# "), []byte{})
		headerWithoutCrc = bytes.ReplaceAll(headerWithoutCrc, []byte("\n"+HeaderFieldHeaderCRC32+": "+headers[HeaderFieldHeaderCRC32]), []byte{})

		if !ValidateCRC32(headerWithoutCrc, headerCrc32) {
			if !ignoreChecksumMismatch {
				return nil, errors.Join(errorParsingHeader, errorValidationFailure, errors.New("header CRC-32 mismatch: expected "+headers[HeaderFieldHeaderCRC32]+", got "+fmt.Sprintf("%x", crc32.ChecksumIEEE(headerWithoutCrc))))
			}

			log.Warn(Warning("Header CRC-32 mismatch!"))
		}
	}

	var pgpMessage *crypto.PGPMessage
	var body []byte
	body, err = DeserializeBinary(&paperCryptFileContentsSplit[1])
	if err != nil {
		return nil, errors.Join(errorParsingBody, err)
	}

	pgpMessage = crypto.NewPGPMessage(body)

	// 5. Verify Body Hashes
	body = pgpMessage.GetBinary()

	// 5.1 Verify Content Length
	bodyLength, ok := headers[HeaderFieldContentLength]
	if !ok {
		return nil, errors.Join(errorParsingBody, newFieldNotPresentError(HeaderFieldContentLength))
	}

	if fmt.Sprint(len(body)) != bodyLength {
		return nil, errors.Join(errorValidationFailure, fmt.Errorf("`%s` mismatch: expected %s, got %d", HeaderFieldContentLength, bodyLength, len(body)))
	}

	// 5.2 Verify CRC-32
	bodyCrc32, ok := headers[HeaderFieldCRC32]
	if !ok {
		return nil, errors.Join(errorValidationFailure, newFieldNotPresentError(HeaderFieldCRC32))
	}

	bodyCrc32Uint32, err := ParseHexUint32(bodyCrc32)
	if err != nil {
		return nil, errors.Join(errorParsingBody, err)
	}

	if !ValidateCRC32(body, bodyCrc32Uint32) {
		if !ignoreChecksumMismatch {
			return nil, errors.Join(errorValidationFailure, fmt.Errorf("`%s` mismatch", HeaderFieldCRC32))
		}

		log.Warn(Warning("Content CRC-32 mismatch!"))
	}

	// 5.3 Verify CRC-24
	bodyCrc24, ok := headers[HeaderFieldCRC24]
	if !ok {
		return nil, errors.Join(errorParsingBody, newFieldNotPresentError(HeaderFieldCRC24))
	}

	bodyCrc24Uint32, err := ParseHexUint32(bodyCrc24)
	if err != nil {
		return nil, errors.Join(errorParsingBody, err)
	}

	if !ValidateCRC24(body, bodyCrc24Uint32) {
		if !ignoreChecksumMismatch {
			return nil, errors.Join(errorValidationFailure, fmt.Errorf("`%s` mismatch", HeaderFieldCRC24))
		}

		log.Warn(Warning("Content CRC-24 mismatch!"))
	}

	// 5.4 Verify SHA-256
	bodySha256, ok := headers[HeaderFieldSHA256]
	if !ok {
		return nil, errors.Join(errorParsingBody, newFieldNotPresentError(HeaderFieldSHA256))
	}

	bodySha256Bytes, err := BytesFromBase64(bodySha256)
	if err != nil {
		return nil, errors.Join(errorParsingBody, err)
	}

	actualSha256 := sha256.Sum256(body)
	if !bytes.Equal(actualSha256[:], bodySha256Bytes) {
		if !ignoreChecksumMismatch {
			return nil, errors.Join(errorValidationFailure, fmt.Errorf("`%s` mismatch", HeaderFieldSHA256))
		}

		log.Warn(Warning("Content SHA-256 mismatch!"))
	}

	// 6. Construct PaperCrypt object
	headerDate, ok := headers[HeaderFieldDate]
	if !ok {
		log.Warn(Warning("Date not present in header!"))
	}

	timestamp, err := time.Parse("Mon, 02 Jan 2006 15:04:05.000000000 MST", headerDate)
	if err != nil {
		return nil, errors.Join(errors.New("invalid date format"), err)
	}

	// we don't need to pass the checksums, as they are already verified
	// and will just be recalculated
	paperCrypt := NewPaperCryptV1(
		versionLine,
		pgpMessage,
		headers[HeaderFieldSerial],
		headers[HeaderFieldPurpose],
		headers[HeaderFieldComment],
		timestamp,
	)

	// 7. Serialize PaperCrypt object for debugging purposes
	_, err = json.MarshalIndent(paperCrypt, "", "  ")
	if err != nil {
		return nil, errors.Join(errors.New("error encoding JSON"), err)
	}
	log.WithField("json", paperCrypt).Debug("Serialized PaperCrypt document")

	// upgrade to next version
	v2, err := paperCrypt.ToNextVersion()
	if err != nil {
		return nil, errors.Join(errors.New("error upgrading to next version"), err)
	}

	return v2, nil
}
