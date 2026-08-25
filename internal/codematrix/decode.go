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
	"encoding/base64"
	"errors"
	"image"
	"io"

	"github.com/makiuchi-d/gozxing"
	"github.com/makiuchi-d/gozxing/aztec"
)

// MaxDecodedPayloadSize is the maximum allowed size in bytes for a
// decompressed payload read by Decode. This guards against decompression
// bombs.
const MaxDecodedPayloadSize = 10 * 1024 * 1024 // 10 MiB

// Decode reads a single Aztec code image, base64-decodes the gzip payload,
// and returns the original data.
func Decode(img image.Image) ([]byte, error) {
	bmp, err := gozxing.NewBinaryBitmapFromImage(img)
	if err != nil {
		return nil, errors.Join(errors.New("codematrix: create bitmap"), err)
	}

	reader := aztec.NewAztecReader()
	result, err := reader.Decode(bmp, nil)
	if err != nil {
		return nil, errors.Join(errors.New("codematrix: aztec decode"), err)
	}

	b64, err := base64.StdEncoding.DecodeString(result.GetText())
	if err != nil {
		return nil, errors.Join(errors.New("codematrix: base64 decode"), err)
	}

	gz, err := gzip.NewReader(bytes.NewReader(b64))
	if err != nil {
		return nil, errors.Join(errors.New("codematrix: gzip reader"), err)
	}
	data, err := io.ReadAll(io.LimitReader(gz, MaxDecodedPayloadSize+1))
	if err != nil {
		return nil, errors.Join(errors.New("codematrix: gzip read"), err)
	}
	if len(data) > MaxDecodedPayloadSize {
		return nil, errors.New("codematrix: decoded payload exceeds maximum size")
	}

	return data, nil
}
