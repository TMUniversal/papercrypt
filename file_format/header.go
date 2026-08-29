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
	"errors"
	"fmt"
)

const (
	DefaultBytesPerLine = 24
)

const (
	HeaderFieldVersion       = "PaperCrypt Version"
	HeaderFieldSerial        = "Content Serial"
	HeaderFieldPurpose       = "Purpose"
	HeaderFieldComment       = "Comment"
	HeaderFieldDate          = "Date"
	HeaderFieldDataFormat    = "Data Format"
	HeaderFieldContentLength = "Content Length"
	HeaderFieldSHA256        = "Content SHA-256"
	HeaderFieldHeaderCRC32   = "Header CRC-32"
)

var (
	errorParsingHeader     = errors.New("error parsing header")
	errorParsingBody       = errors.New("error parsing body")
	errorValidationFailure = errors.New("validation failure")
)

func newFieldNotPresentError(field string) error {
	return fmt.Errorf("`%s` not present in header", field)
}
