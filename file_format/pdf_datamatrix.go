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
	"errors"
	"image/png"

	"github.com/makiuchi-d/gozxing"
	"github.com/makiuchi-d/gozxing/datamatrix"
)

func GenerateDataMatrix(serial string) ([]byte, error) {
	enc := datamatrix.NewDataMatrixWriter()
	code, err := enc.Encode(serial, gozxing.BarcodeFormat_DATA_MATRIX, 384, 384, nil)
	if err != nil {
		return nil, errors.Join(errors.New("error generating Data Matrix code"), err)
	}

	buf := new(bytes.Buffer)
	if err := png.Encode(buf, code); err != nil {
		return nil, errors.Join(errors.New("error generating Data Matrix code PNG"), err)
	}
	return buf.Bytes(), nil
}
