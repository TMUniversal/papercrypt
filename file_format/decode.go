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

import "errors"

// DecodeData decodes and, if the data was encrypted with PaperCrypt (data
// format is PaperCryptDataFormatPGP), decrypts the data, returning the
// original binary data.
func DecodeData(p *PaperCrypt, passphrase []byte) ([]byte, error) {
	if p == nil {
		return nil, errors.New("decode: nil PaperCrypt")
	}
	handler, err := getHandler(p.DataFormat)
	if err != nil {
		return nil, err
	}
	return handler.process(p.Data, passphrase)
}
