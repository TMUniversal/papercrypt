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
	"compress/gzip"
	"errors"
	"image"
	"image/png"

	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/qr"
	"github.com/dasio/base45"
)

// outputSize is the barcode output size in pixels (165mm at 1200dpi).
const outputSize = 7795

// Encode compresses data with gzip, base45-encodes it, and encodes it
// into a single QR code image using alphanumeric mode.
func Encode(data []byte) (image.Image, error) {
	var buf bytes.Buffer
	gz, err := gzip.NewWriterLevel(&buf, gzip.BestCompression)
	if err != nil {
		return nil, errors.Join(errors.New("codematrix: gzip writer"), err)
	}
	if _, err := gz.Write(data); err != nil {
		return nil, errors.Join(errors.New("codematrix: gzip write"), err)
	}
	if err := gz.Close(); err != nil {
		return nil, errors.Join(errors.New("codematrix: gzip close"), err)
	}

	encoded := base45.EncodeToString(buf.Bytes())

	code, err := qr.Encode(encoded, qr.H, qr.AlphaNumeric)
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

// EncodePNG encodes data into a single QR code and returns the PNG-encoded bytes.
func EncodePNG(data []byte) ([]byte, error) {
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
