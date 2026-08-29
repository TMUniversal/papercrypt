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
	"fmt"

	"github.com/tmuniversal/papercrypt/v3/crc24"
	"github.com/tmuniversal/papercrypt/v3/pdf"
)

// GetPDF renders the PaperCrypt document as a printable PDF, combining the
// human-readable text, the QR code (unless no2D) carrying the full binary
// container and the Data Matrix serial label.
func GetPDF(p *PaperCrypt, no2D bool, lowerCaseEncoding bool) ([]byte, error) {
	text, err := GetText(p, lowerCaseEncoding)
	if err != nil {
		return nil, fmt.Errorf("error getting text content: %s", err)
	}

	header, data := splitHeaderBody(text)
	if data == nil {
		return nil, fmt.Errorf("error splitting text content into header and data")
	}

	var qrImage []byte
	if !no2D {
		qrImage, err = GenerateQR(p)
		if err != nil {
			return nil, err
		}
	}

	dm, err := GenerateDataMatrix(p.SerialNumber)
	if err != nil {
		return nil, err
	}

	cfg := pdf.Config{
		HasQR:           !no2D,
		SheetSerial:     p.SerialNumber,
		CreatedAt:       p.CreatedAt,
		Purpose:         p.Purpose,
		DataQRImage:     qrImage,
		DataMatrixImage: dm,
		TextParts:       []string{string(header), string(data)},
		BytesPerLine:    DefaultBytesPerLine,
		CRC24Polynomial: crc24.CRC24Polynomial,
		CRC24Initial:    crc24.CRC24Initial,
	}

	return pdf.New(pdfMode(p, no2D)).Render(cfg)
}

// pdfMode picks the sheet layout for the document's data format.
func pdfMode(p *PaperCrypt, no2D bool) pdf.Mode {
	switch {
	case p.DataFormat == PaperCryptDataFormatRaw && no2D:
		return pdf.ModeRawNoQR
	case p.DataFormat == PaperCryptDataFormatRaw:
		return pdf.ModeRawQR
	case no2D:
		return pdf.ModePGPNoQR
	default:
		return pdf.ModePGPQR
	}
}
