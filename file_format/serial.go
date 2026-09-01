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
	"crypto/rand"
	"encoding/base32"
	"errors"
)

func GenerateSerial(length uint8) (string, error) {
	// Encode length random bytes: base32 yields >= length characters for any
	// nonzero length, so the trailing slice never runs off the end even if a
	// byte is zero.
	random := make([]byte, length)
	if _, err := rand.Read(random); err != nil {
		return "", errors.Join(errors.New("error generating random bytes"), err)
	}

	buf := new(bytes.Buffer)
	encoder := base32.NewEncoder(base32.StdEncoding, buf)
	if _, err := encoder.Write(random); err != nil {
		return "", errors.Join(errors.New("error encoding bytes"), err)
	}
	if err := encoder.Close(); err != nil {
		return "", errors.Join(errors.New("error closing base32 encoder"), err)
	}

	return buf.String()[:length], nil
}
