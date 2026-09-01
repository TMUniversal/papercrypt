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
	"errors"
	"time"

	"github.com/ProtonMail/gopenpgp/v3/crypto"
	"github.com/caarlos0/log"
	"github.com/tmuniversal/papercrypt/v3/internal"
)

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

	versionLine, err := validateVersion(headers, ignoreVersionMismatch)
	if err != nil {
		return nil, err
	}

	if err := validateHeaderCRC32(headers, headersSection, ignoreChecksumMismatch); err != nil {
		return nil, err
	}

	dataFormat, err := validateDataFormat(headers)
	if err != nil {
		return nil, err
	}

	body, err := DeserializeBinary(&bodySection)
	if err != nil {
		return nil, errors.Join(errorParsingBody, err)
	}

	switch dataFormat {
	case PaperCryptDataFormatPGP:
		body = crypto.NewPGPMessage(body).Bytes()
	case PaperCryptDataFormatRaw:
		// raw data is stored as-is
	default:
		return nil, errors.Join(errorParsingBody, errors.New("unsupported data format"))
	}

	if err := validateContentLength(body, headers); err != nil {
		return nil, err
	}

	if err := validateSHA256(body, headers, ignoreChecksumMismatch); err != nil {
		return nil, err
	}

	headerDate, ok := headers[HeaderFieldDate]
	if !ok {
		return nil, errors.Join(errorParsingHeader, newFieldNotPresentError(HeaderFieldDate))
	}

	timestamp, err := time.Parse(internal.TimeStampFormatLong, headerDate)
	if err != nil {
		return nil, errors.Join(errors.New("invalid date format"), err)
	}

	// checksums are already verified and recalculated by NewPaperCrypt
	paperCrypt := NewPaperCrypt(
		versionLine,
		body,
		headers[HeaderFieldSerial],
		headers[HeaderFieldPurpose],
		headers[HeaderFieldComment],
		timestamp,
		dataFormat,
	)

	return paperCrypt, nil
}
