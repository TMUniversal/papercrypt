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
	"image/color"
	"image/png"

	"github.com/zxing-cpp/zxing-cpp/wrappers/go/zxingcpp"
)

// aztecSize is the barcode output size in pixels (165mm at 1200dpi).
const aztecSize = 7795

// Encode compresses data with gzip and encodes it into a single Aztec code image.
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
	compressed := buf.Bytes()

	bc, err := zxingcpp.CreateBarcode(
		compressed,
		zxingcpp.BarcodeFormatAztec,
		zxingcpp.WithEcLevel("30%"),
	)
	if err != nil {
		return nil, errors.Join(errors.New("codematrix: barcode encode"), err)
	}
	defer func() {
		_ = bc.Close()
	}()

	nativeImg, err := bc.ToImage()
	if err != nil {
		return nil, errors.Join(errors.New("codematrix: barcode to image"), err)
	}

	return scaleToGray(nativeImg, aztecSize, aztecSize), nil
}

// scaleToGray scales an image to the target dimensions and returns
// a grayscale image with a white background.
func scaleToGray(src image.Image, width, height int) *image.Gray {
	srcBounds := src.Bounds()
	srcW := srcBounds.Dx()
	srcH := srcBounds.Dy()

	scaleX := float64(width) / float64(srcW)
	scaleY := float64(height) / float64(srcH)
	scale := scaleX
	if scaleY < scale {
		scale = scaleY
	}

	outW := int(float64(srcW) * scale)
	outH := int(float64(srcH) * scale)
	offsetX := (width - outW) / 2
	offsetY := (height - outH) / 2

	out := image.NewGray(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			out.SetGray(x, y, color.Gray{Y: 255})
		}
	}

	for dy := 0; dy < outH; dy++ {
		sy := srcBounds.Min.Y + int(float64(dy)/scale)
		oy := offsetY + dy
		for dx := 0; dx < outW; dx++ {
			sx := srcBounds.Min.X + int(float64(dx)/scale)
			ox := offsetX + dx
			r, g, b, _ := src.At(sx, sy).RGBA()
			// luminance is always 0-255
			lum := uint8((r*299 + g*587 + b*114 + 500) / 1000) //nolint:gosec
			if lum < 128 {
				out.SetGray(ox, oy, color.Gray{Y: 0})
			}
		}
	}

	return out
}

// EncodePNG encodes data into a barcode and returns the PNG-encoded bytes.
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
