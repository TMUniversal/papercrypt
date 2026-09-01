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
	"compress/gzip"
	"errors"
	"fmt"

	"github.com/ProtonMail/gopenpgp/v3/crypto"
	"github.com/tmuniversal/papercrypt/v3/internal/decompression"
)

type formatHandler struct {
	decode func(maxDecompressedSize int, data, passphrase []byte) ([]byte, error)
}

var formatHandlers = map[PaperCryptDataFormat]formatHandler{
	PaperCryptDataFormatPGP: {decode: decodePGPData},
	PaperCryptDataFormatRaw: {decode: decodeRawData},
}

func getHandler(format PaperCryptDataFormat) (formatHandler, error) {
	handler, ok := formatHandlers[format]
	if !ok {
		return formatHandler{}, fmt.Errorf("unsupported data format %v", format)
	}
	return handler, nil
}

func decodeRawData(_ int, data, _ []byte) ([]byte, error) {
	return data, nil
}

func decodePGPData(maxDecompressedSize int, data, passphrase []byte) ([]byte, error) {
	gzipReader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, errors.Join(errors.New("error creating gzip reader"), err)
	}

	decompressed, err := decompression.ReadAll(gzipReader, maxDecompressedSize)
	if err != nil {
		return nil, errors.Join(errors.New("error reading from gzip reader"), err)
	}
	if err := gzipReader.Close(); err != nil {
		return nil, errors.Join(errors.New("error closing gzip reader"), err)
	}

	pgpMessage := crypto.NewPGPMessage(decompressed)

	pgp := crypto.PGP()
	decryptionHandler, err := pgp.Decryption().Password(passphrase).New()
	if err != nil {
		return nil, errors.Join(errors.New("error creating decryption handler"), err)
	}

	decrypted, err := decryptionHandler.Decrypt(pgpMessage.Bytes(), crypto.Bytes)
	if err != nil {
		return nil, errors.Join(errors.New("error decrypting data"), err)
	}

	return decrypted.Bytes(), nil
}
