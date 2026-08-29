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
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"hash/crc32"
	"strings"
	"time"

	"github.com/ProtonMail/gopenpgp/v3/crypto"
	"github.com/caarlos0/log"
	"github.com/tmuniversal/papercrypt/v3/crc24"
	"github.com/tmuniversal/papercrypt/v3/internal"
	"github.com/tmuniversal/papercrypt/v3/terminal"
)

func (p *PaperCrypt) GetText(lowerCaseEncoding bool) ([]byte, error) {
	header := fmt.Sprintf(
		`%s: %s
%s: %s
%s: %s
%s: %s
%s: %s
%s: %s
%s: %d
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
		p.CreatedAt.Format(internal.TimeStampFormatLong),
		HeaderFieldDataFormat,
		p.DataFormat,
		HeaderFieldContentLength,
		p.GetDataLength(),
		HeaderFieldSHA256,
		base64.StdEncoding.EncodeToString(p.DataSHA256[:]))

	headerCRC32 := crc32.ChecksumIEEE([]byte(header))

	serializedData, err := p.GetBinarySerialized()
	if err != nil {
		return nil, errors.Join(errors.New("failed to get serialized data"), err)
	}
	if lowerCaseEncoding {
		serializedData = strings.ToLower(serializedData)
	}

	return fmt.Appendf(nil, `%s
%s: %08x


%s
`,
		header,
		HeaderFieldHeaderCRC32,
		headerCRC32,
		serializedData), nil
}

// TextToHeaderMap expects "Key: Value" header lines; the "# " prefix is stripped from keys.
func TextToHeaderMap(text []byte) (map[string]string, error) {
	headers := make(map[string]string)

	headerLines := bytes.Split(text, []byte("\n"))
	for _, headerLine := range headerLines {
		headerLineSplit := bytes.SplitN(headerLine, []byte(": "), 2)
		if len(headerLineSplit) != 2 {
			return nil, errors.Join(
				errorParsingHeader,
				fmt.Errorf("error parsing header line: %s", headerLine),
			)
		}

		key := string(headerLineSplit[0])
		key = strings.TrimPrefix(key, "# ")

		headers[key] = string(headerLineSplit[1])
	}

	return headers, nil
}

func SplitTextHeaderAndBody(data []byte) ([]byte, []byte, error) {
	dataSplit := bytes.SplitN(data, []byte("\n\n\n"), 2)
	if len(dataSplit) != 2 {
		return nil, nil, errors.New(
			"header not discernible, header and content should be separated by two empty lines",
		)
	}
	return dataSplit[0], dataSplit[1], nil
}

func DeserializeText(
	data []byte,
	ignoreVersionMismatch bool,
	ignoreChecksumMismatch bool,
) (*PaperCrypt, error) {
	paperCryptFileContents := internal.NormalizeLineEndings(data)

	headersSection, bodySection, err := SplitTextHeaderAndBody(paperCryptFileContents)
	if err != nil {
		return nil, errors.Join(errorParsingHeader, err)
	}

	headers, err := TextToHeaderMap(headersSection)
	if err != nil {
		return nil, errors.Join(errorParsingHeader, err)
	}

	log.WithField("headers", headers).Debug("Read headers")

	versionLine, ok := headers[HeaderFieldVersion]
	if !ok {
		if !ignoreVersionMismatch {
			return nil, errors.Join(errorParsingHeader, newFieldNotPresentError(HeaderFieldVersion))
		}

		log.Warn(terminal.Warning("PaperCrypt Version not present in header."))
	}

	majorVersion := PaperCryptContainerVersionFromString(versionLine)
	if !ignoreVersionMismatch &&
		(majorVersion != PaperCryptContainerVersionMajor3 && majorVersion != PaperCryptContainerVersionDevel) {
		return nil, errors.Join(
			errorParsingHeader,
			fmt.Errorf("unsupported PaperCrypt version '%s'", versionLine),
		)
	}

	{
		headerCrc, ok := headers[HeaderFieldHeaderCRC32]
		if !ok {
			if !ignoreChecksumMismatch {
				return nil, errors.Join(
					errorParsingHeader,
					newFieldNotPresentError(HeaderFieldHeaderCRC32),
				)
			}

			log.Warn(terminal.Warning("Header CRC-32 not present in header"))
		}

		headerCrc = strings.ToLower(headerCrc)
		headerCrc = strings.ReplaceAll(headerCrc, "0x", "")
		headerCrc = strings.ReplaceAll(headerCrc, " ", "")
		headerCrc32, err := ParseHexUint32(headerCrc)
		if err != nil {
			return nil, errors.Join(errorParsingHeader, errors.New("invalid CRC-32 format"), err)
		}

		headerWithoutCrc := bytes.ReplaceAll(headersSection, []byte("# "), []byte{})
		headerWithoutCrc = bytes.ReplaceAll(
			headerWithoutCrc,
			[]byte("\n"+HeaderFieldHeaderCRC32+": "+headers[HeaderFieldHeaderCRC32]),
			[]byte{},
		)

		if !crc24.ValidateCRC32(headerWithoutCrc, headerCrc32) {
			if !ignoreChecksumMismatch {
				return nil, errors.Join(
					errorParsingHeader,
					errorValidationFailure,
					errors.New(
						"header CRC-32 mismatch: expected "+headers[HeaderFieldHeaderCRC32]+", got "+fmt.Sprintf(
							"%x",
							crc32.ChecksumIEEE(headerWithoutCrc),
						),
					),
				)
			}

			log.Warn(terminal.Warning("Header CRC-32 mismatch!"))
		}
	}

	var dataFormat PaperCryptDataFormat
	{
		dataFormatString, ok := headers[HeaderFieldDataFormat]
		if !ok {
			return nil, errors.Join(
				errorParsingHeader,
				newFieldNotPresentError(HeaderFieldDataFormat),
			)
		}

		log.Debugf("Data Format: %s", dataFormatString)

		dataFormat = PaperCryptDataFormatFromString(dataFormatString)
	}

	var pgpMessage *crypto.PGPMessage
	var body []byte
	body, err = DeserializeBinary(&bodySection)
	if err != nil {
		return nil, errors.Join(errorParsingBody, err)
	}

	switch dataFormat {
	case PaperCryptDataFormatPGP:
		pgpMessage = crypto.NewPGPMessage(body)
		body = pgpMessage.Bytes()
	case PaperCryptDataFormatRaw:
		// do nothing
	default:
		return nil, errors.Join(errorParsingBody, errors.New("unsupported data format"))
	}

	bodyLength, ok := headers[HeaderFieldContentLength]
	if !ok {
		return nil, errors.Join(errorParsingBody, newFieldNotPresentError(HeaderFieldContentLength))
	}

	if fmt.Sprint(len(body)) != bodyLength {
		return nil, errors.Join(
			errorValidationFailure,
			fmt.Errorf(
				"`%s` mismatch: expected %s, got %d",
				HeaderFieldContentLength,
				bodyLength,
				len(body),
			),
		)
	}

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
			return nil, errors.Join(
				errorValidationFailure,
				fmt.Errorf(
					"`%s` mismatch: expected %s, found %s (content length %d)",
					HeaderFieldSHA256,
					bodySha256,
					base64.StdEncoding.EncodeToString(actualSha256[:]),
					len(body),
				),
			)
		}

		log.Warn(terminal.Warning("Content SHA-256 mismatch!"))
	}

	headerDate, ok := headers[HeaderFieldDate]
	if !ok {
		log.Warn(terminal.Warning("Date not present in header!"))
	}

	timestamp, err := time.Parse(internal.TimeStampFormatLong, headerDate)
	if err != nil {
		return nil, errors.Join(errors.New("invalid date format"), err)
	}

	// we don't need to pass the checksums, as they are already verified
	// and will just be recalculated
	paperCrypt := NewPaperCrypt(
		versionLine,
		body,
		headers[HeaderFieldSerial],
		headers[HeaderFieldPurpose],
		headers[HeaderFieldComment],
		timestamp,
		dataFormat,
	)

	_, err = json.MarshalIndent(paperCrypt, "", "  ")
	if err != nil {
		return nil, errors.Join(errors.New("error encoding JSON"), err)
	}
	log.WithField("json", paperCrypt).Debug("Serialized PaperCrypt document")

	return paperCrypt, nil
}
