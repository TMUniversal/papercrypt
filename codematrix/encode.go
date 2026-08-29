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

package codematrix

import (
	"bytes"
	"errors"
	"image"
	"image/png"

	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/qr"
)

// outputSize is a 165mm square at 1200dpi.
const outputSize = 7795

func Encode(data string) (image.Image, error) {
	code, err := qr.Encode(data, qr.H, qr.AlphaNumeric)
	if err != nil {
		return nil, errors.Join(errors.New("codematrix: qr encode"), err)
	}

	code, err = barcode.Scale(code, outputSize, outputSize)
	if err != nil {
		return nil, errors.Join(errors.New("codematrix: scale"), err)
	}

	converted := image.NewGray(code.Bounds())
	for y := 0; y < code.Bounds().Dy(); y++ {
		for x := 0; x < code.Bounds().Dx(); x++ {
			converted.Set(x, y, code.At(x, y))
		}
	}

	return converted, nil
}

func EncodePNG(data string) ([]byte, error) {
	img, err := Encode(data)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, errors.Join(errors.New("codematrix: png encode"), err)
	}
	return buf.Bytes(), nil
}
