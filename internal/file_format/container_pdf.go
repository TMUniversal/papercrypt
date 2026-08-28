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
	"fmt"
	"image/png"
	"strings"

	"github.com/makiuchi-d/gozxing"
	"github.com/makiuchi-d/gozxing/datamatrix"
	"github.com/tmuniversal/papercrypt/v3/internal/codematrix"
	"github.com/tmuniversal/papercrypt/v3/internal/crc24"
	"github.com/tmuniversal/papercrypt/v3/internal/file_format/envelope"
	"github.com/tmuniversal/papercrypt/v3/internal/pdf"
)

// GetPDF returns the binary representation of the paper crypt
// The PDF will be generated to include some basic information about papercrypt,
// some metadata, optionally a 2D-Code, and the encrypted data.
func (p *PaperCrypt) GetPDF(no2D bool, lowerCaseEncoding bool) ([]byte, error) {
	text, err := p.GetText(lowerCaseEncoding)
	if err != nil {
		return nil, fmt.Errorf("error getting text content: %s", err)
	}

	// split at 2 empty lines, to get the header and the data
	parts := strings.Split(string(text), "\n\n\n")
	if len(parts) != 2 {
		return nil, fmt.Errorf("error splitting text content into header and data")
	}

	data2D, err := p.encodeDataQR(no2D)
	if err != nil {
		return nil, err
	}

	dm, err := p.generateDataMatrix()
	if err != nil {
		return nil, err
	}

	var qrImage []byte
	if data2D != nil {
		qrImage = data2D.Bytes()
	}

	cfg := pdf.Config{
		HasQR:           !no2D,
		SheetSerial:     p.SerialNumber,
		CreatedAt:       p.CreatedAt,
		Purpose:         p.Purpose,
		DataQRImage:     qrImage,
		DataMatrixImage: dm.Bytes(),
		TextParts:       parts,
		BytesPerLine:    BytesPerLine,
		CRC24Polynomial: crc24.CRC24Polynomial,
		CRC24Initial:    crc24.CRC24Initial,
	}

	return pdf.New(pdfMode(p, no2D)).Render(cfg)
}

// pdfMode returns the recovery-sheet mode matching the data format and whether a QR code is printed.
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

func (p *PaperCrypt) encodeDataQR(no2D bool) (*bytes.Buffer, error) {
	if no2D {
		return nil, nil
	}

	qrBin, err := MarshalBinary(p)
	if err != nil {
		return nil, errors.Join(errors.New("error marshalling PaperCrypt to binary"), err)
	}

	qrData := envelope.Wrap(qrBin, envelope.Base45Encoder{})

	pngBytes, err := codematrix.EncodePNG(qrData)
	if err != nil {
		return nil, err
	}

	buf := new(bytes.Buffer)
	buf.Write(pngBytes)
	return buf, nil
}

func (p *PaperCrypt) generateDataMatrix() (*bytes.Buffer, error) {
	enc := datamatrix.NewDataMatrixWriter()
	code, err := enc.Encode(p.SerialNumber, gozxing.BarcodeFormat_DATA_MATRIX, 384, 384, nil)
	if err != nil {
		return nil, errors.Join(errors.New("error generating Data Matrix code"), err)
	}

	buf := new(bytes.Buffer)
	if err := png.Encode(buf, code); err != nil {
		return nil, errors.Join(errors.New("error generating Data Matrix code PNG"), err)
	}
	return buf, nil
}
