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
	"compress/gzip"
	"errors"

	"github.com/ProtonMail/gopenpgp/v3/crypto"
)

// Decode decodes and, if the data was encrypted with PaperCrypt (data format is PaperCryptDataFormatPGP),
// decrypts the data, returning the original binary data.
func (p *PaperCrypt) Decode(passphrase []byte) ([]byte, error) {
	data := p.Data
	if p.DataFormat == PaperCryptDataFormatPGP {
		// 1. Decompress ciphertext
		gzipReader, err := gzip.NewReader(bytes.NewReader(p.Data))
		if err != nil {
			return nil, errors.Join(errors.New("error creating gzip reader"), err)
		}

		decompressed := new(bytes.Buffer)
		if _, err := decompressed.ReadFrom(gzipReader); err != nil {
			return nil, errors.Join(errors.New("error reading from gzip reader"), err)
		}
		if err := gzipReader.Close(); err != nil {
			return nil, errors.Join(errors.New("error closing gzip reader"), err)
		}

		pgpMessage := crypto.NewPGPMessage(decompressed.Bytes())

		// 2. Decrypt
		pgp := crypto.PGP()
		decHandle, err := pgp.Decryption().Password(passphrase).New()
		if err != nil {
			return nil, errors.Join(errors.New("error creating decryption handle"), err)
		}

		decrypted, err := decHandle.Decrypt(pgpMessage.Bytes(), crypto.Bytes)
		if err != nil {
			return nil, errors.Join(errors.New("error decrypting data"), err)
		}

		return decrypted.Bytes(), nil
	}

	// Raw mode: data is stored as-is
	return data, nil
}
