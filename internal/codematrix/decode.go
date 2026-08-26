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
	"fmt"
	"image"
	"io"

	"github.com/zxing-cpp/zxing-cpp/wrappers/go/zxingcpp"
)

// MaxDecodedPayloadSize is the maximum allowed size in bytes for a
// decompressed payload read by Decode. This guards against decompression
// bombs.
var MaxDecodedPayloadSize = 10 * 1024 * 1024 // 10 MiB

// limitDecodedPayload controls whether Decode enforces MaxDecodedPayloadSize.
// When true, payloads exceeding the limit are rejected.
// Set via SetLimitDecodedPayload; defaults to true.
var limitDecodedPayload = true

// SetLimitDecodedPayload sets whether Decode enforces the maximum decoded
// payload size. When disabled, decompression bombs are not guarded against.
func SetLimitDecodedPayload(enabled bool) {
	limitDecodedPayload = enabled
}

// Decode reads a single barcode image, decompresses the gzip payload,
// and returns the original data.
func Decode(img image.Image) ([]byte, error) {
	barcodes, err := zxingcpp.ReadBarcodes(
		img,
		zxingcpp.TryHarder(true),
		zxingcpp.WithMaxNumberOfSymbols(1),
		zxingcpp.WithFormats(zxingcpp.BarcodeFormatAztec),
	)
	if err != nil {
		return nil, errors.Join(errors.New("codematrix: barcode decode"), err)
	}
	if len(barcodes) == 0 {
		return nil, errors.New("codematrix: no barcode found")
	}
	defer barcodes[0].Close()

	raw := barcodes[0].Bytes()

	gz, err := gzip.NewReader(bytes.NewReader(raw))
	if err != nil {
		return nil, errors.Join(errors.New("codematrix: gzip reader"), err)
	}
	var data []byte
	if limitDecodedPayload {
		data, err = io.ReadAll(io.LimitReader(gz, int64(MaxDecodedPayloadSize)+1))
		if err != nil {
			return nil, errors.Join(errors.New("codematrix: gzip read"), err)
		}
		if len(data) > MaxDecodedPayloadSize {
			return nil, fmt.Errorf("codematrix: decoded payload exceeds maximum size (%d > %d)", len(data), MaxDecodedPayloadSize)
		}
	} else {
		data, err = io.ReadAll(gz)
		if err != nil {
			return nil, errors.Join(errors.New("codematrix: gzip read"), err)
		}
	}

	return data, nil
}
