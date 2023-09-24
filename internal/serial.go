/*
 * This file is part of PaperCrypt.
 *
 * PaperCrypt lets you prepare encrypted messages for printing on paper.
 * Copyright (C) 2023 TMUniversal <me@tmuniversal.eu>.
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
	"crypto/rand"
	"encoding/base32"
	"errors"
	"math"
	"math/big"
)

// GenerateSerial generates a random serial number of length `length`
func GenerateSerial(length uint8) (string, error) {
	// generate `length` random bytes,
	// encode them as base64,
	// and return the first `length` characters

	numbers := make([]*big.Int, length)

	for i := uint8(0); i < length; i++ {
		randInt, err := rand.Int(rand.Reader, big.NewInt(math.MaxInt64))
		if err != nil {
			return "", errors.Join(errors.New("error generating random bytes"), err)
		}

		numbers[i] = randInt
	}

	buf := new(bytes.Buffer)
	encoder := base32.NewEncoder(base32.StdEncoding, buf)
	for _, number := range numbers {
		_, err := encoder.Write(number.Bytes())
		if err != nil {
			return "", errors.Join(errors.New("error encoding bytes"), err)
		}
	}
	err := encoder.Close()
	if err != nil {
		return "", errors.Join(errors.New("error closing base64 encoder"), err)
	}

	return buf.String()[:length], nil
}

// DecodeSerial decodes a serial number
func DecodeSerial(serial string) ([]byte, error) {
	decoder := base32.NewDecoder(base32.StdEncoding, bytes.NewBufferString(serial))
	var decoded []byte
	_, err := decoder.Read(decoded)
	if err != nil {
		return nil, errors.Join(errors.New("error decoding serial"), err)
	}

	return decoded, nil
}
