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
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"hash/crc32"
	"strings"

	"github.com/caarlos0/log"
	"github.com/tmuniversal/papercrypt/v3/terminal"
)

func validateVersion(headers map[string]string, ignoreVersionMismatch bool) (string, error) {
	versionLine, ok := headers[HeaderFieldVersion]
	if !ok {
		if !ignoreVersionMismatch {
			return "", errors.Join(errorParsingHeader, newFieldNotPresentError(HeaderFieldVersion))
		}

		log.Warn(terminal.Warning("PaperCrypt Version not present in header."))
	}

	majorVersion := PaperCryptContainerVersionFromString(versionLine)
	if !ignoreVersionMismatch &&
		(majorVersion != PaperCryptContainerVersionMajor3 && majorVersion != PaperCryptContainerVersionDevel) {
		return "", errors.Join(
			errorParsingHeader,
			fmt.Errorf("unsupported PaperCrypt version '%s'", versionLine),
		)
	}

	return versionLine, nil
}

func validateHeaderCRC32(
	headers map[string]string,
	headersSection []byte,
	ignoreChecksumMismatch bool,
) error {
	headerCrc, ok := headers[HeaderFieldHeaderCRC32]
	if !ok {
		return errors.Join(
			errorParsingHeader,
			newFieldNotPresentError(HeaderFieldHeaderCRC32),
		)
	}

	headerCrc = strings.ToLower(headerCrc)
	headerCrc = strings.ReplaceAll(headerCrc, "0x", "")
	headerCrc = strings.ReplaceAll(headerCrc, " ", "")
	headerCrc32, err := ParseHexUint32(headerCrc)
	if err != nil {
		return errors.Join(errorParsingHeader, errors.New("invalid CRC-32 format"), err)
	}

	headerWithoutCrc := bytes.ReplaceAll(headersSection, []byte("# "), []byte{})
	headerWithoutCrc = bytes.ReplaceAll(
		headerWithoutCrc,
		[]byte("\n"+HeaderFieldHeaderCRC32+": "+headers[HeaderFieldHeaderCRC32]),
		[]byte{},
	)

	if crc32.ChecksumIEEE(headerWithoutCrc) != headerCrc32 {
		if !ignoreChecksumMismatch {
			return errors.Join(
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

	return nil
}

func validateDataFormat(headers map[string]string) (PaperCryptDataFormat, error) {
	dataFormatString, ok := headers[HeaderFieldDataFormat]
	if !ok {
		return 0, errors.Join(
			errorParsingHeader,
			newFieldNotPresentError(HeaderFieldDataFormat),
		)
	}

	log.Debugf("Data Format: %s", dataFormatString)

	return PaperCryptDataFormatFromString(dataFormatString), nil
}

func validateContentLength(body []byte, headers map[string]string) error {
	bodyLength, ok := headers[HeaderFieldContentLength]
	if !ok {
		return errors.Join(errorParsingBody, newFieldNotPresentError(HeaderFieldContentLength))
	}

	if fmt.Sprint(len(body)) != bodyLength {
		return errors.Join(
			errorValidationFailure,
			fmt.Errorf(
				"`%s` mismatch: expected %s, got %d",
				HeaderFieldContentLength,
				bodyLength,
				len(body),
			),
		)
	}

	return nil
}

func validateSHA256(body []byte, headers map[string]string, ignoreChecksumMismatch bool) error {
	bodySha256, ok := headers[HeaderFieldSHA256]
	if !ok {
		return errors.Join(errorParsingBody, newFieldNotPresentError(HeaderFieldSHA256))
	}

	bodySha256Bytes, err := BytesFromBase64(bodySha256)
	if err != nil {
		return errors.Join(errorParsingBody, err)
	}

	actualSha256 := sha256.Sum256(body)
	if !bytes.Equal(actualSha256[:], bodySha256Bytes) {
		if !ignoreChecksumMismatch {
			return errors.Join(
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

	return nil
}
