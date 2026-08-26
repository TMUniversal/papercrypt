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
	"image"
	"image/png"
	"strings"

	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/qr"
	"github.com/jung-kurt/gofpdf/v2"
	"github.com/makiuchi-d/gozxing"
	"github.com/makiuchi-d/gozxing/datamatrix"
	"github.com/tmuniversal/papercrypt/v3/internal"
	"github.com/tmuniversal/papercrypt/v3/internal/codematrix"
	"github.com/tmuniversal/papercrypt/v3/internal/crc24"
	"github.com/tmuniversal/papercrypt/v3/internal/file_format/envelope"
	"github.com/tmuniversal/papercrypt/v3/internal/pdf"
)

const printProductQrCode = false

const (
	// DataLineFontSize sets the font size of data lines in the PDF [pt]
	DataLineFontSize = 11
	// PDFHeaderSheetID holds the text label displayed in the PDF header for the sheet ID.
	PDFHeaderSheetID = "Sheet ID"
	// PDFHeading holds the title of the PDF document, as shown on the first page.
	PDFHeading = "PaperCrypt Recovery Sheet"
	// PDFSectionDescriptionHeading holds the title of the section describing the document.
	PDFSectionDescriptionHeading = "What is this?"
	// PDFSectionDescriptionContent holds the content of the section describing the document.
	PDFSectionDescriptionContent = "This is a PaperCrypt recovery sheet. It contains encrypted data, its own creation date, purpose, and a comment, as well as an identifier. This sheet is intended to help recover the original information, in case it is lost or destroyed."
	// PDFSectionRepresentationHeading holds the title of the section describing the data representation.
	PDFSectionRepresentationHeading = "Binary Data Representation"
	// PDFSectionRepresentationContentBase holds the content of the section describing the data representation.
	PDFSectionRepresentationContentBase = "Data is written as base 16 (hexadecimal) digits, each representing a half-byte. Two half-bytes are grouped together as a byte, which are then grouped together in lines of %d bytes, where bytes are separated by a space. Each line begins with its line number and a colon, denoting its position and the beginning of the data. Each line is then followed by its CRC-24 checksum. The last line holds the checksum of the entire block. For the checksum algorithm, the polynomial mask %#x and initial value %#x are used."
	// PDFSectionRepresentationContentGzip is the gzip-specific suffix appended for PGP-format data.
	PDFSectionRepresentationContentGzip = " Data is compressed using the gzip algorithm."
	// PDFSectionRecoveryHeading holds the title of the section describing how to recover the data.
	PDFSectionRecoveryHeading = "Recovering the data"
	// PDFSectionRecoveryContent holds the content of the section describing how to recover the data.
	PDFSectionRecoveryContent = "Firstly, scan the 2D code, or copy (i.e. type in, or use OCR on) the encrypted data into a computer. Then decrypt it, either using the PaperCrypt CLI, or manually construct the data into a binary file, and decrypt it using OpenPGP-compatible software."
	// PDFSectionRecoveryContentNo2D holds the content of the section describing how to recover the data, if no 2D code is present.
	PDFSectionRecoveryContentNo2D = "Firstly, copy (i.e. type in, or use OCR on) the encrypted data into a computer. Then decrypt it, either using the PaperCrypt CLI, or manually construct the data into a binary file, and decrypt it using OpenPGP-compatible software."
)

// GetPDF returns the binary representation of the paper crypt
// The PDF will be generated to include some basic information about papercrypt,
// some metadata, optionally a 2D-Code, and the encrypted data.
//
// The data will be formatted as
//
//	a) ASCII armored OpenPGP data, if --armor is specified
//	b) Base16 (hex) encoded binary data, if --armor is not specified
//
// The PDF Document will have a header row, containing the following information:
//   - Serial Number
//   - Creation Date
//   - Purpose
//
// and, next to the markdown information, a 2D code containing the encrypted data.
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

	productLinkQr := new(bytes.Buffer)
	if printProductQrCode {
		qrSize := 709

		code, err := qr.Encode(internal.VersionInfo.URL, qr.M, qr.Auto)
		if err != nil {
			return nil, errors.Join(errors.New("error generating 2D code"), err)
		}

		code, err = barcode.Scale(code, qrSize, qrSize)
		if err != nil {
			return nil, errors.Join(errors.New("error scaling 2D code"), err)
		}

		converted := image.NewGray(code.Bounds())
		for y := 0; y < code.Bounds().Dy(); y++ {
			for x := 0; x < code.Bounds().Dx(); x++ {
				converted.Set(x, y, code.At(x, y))
			}
		}

		err = png.Encode(productLinkQr, converted)
		if err != nil {
			return nil, errors.Join(errors.New("error generating 2D code PNG"), err)
		}
	}

	data2D := new(bytes.Buffer)
	dm := new(bytes.Buffer)

	if !no2D {
		qrBin, err := MarshalBinary(p)
		if err != nil {
			return nil, errors.Join(errors.New("error marshalling PaperCrypt to binary"), err)
		}

		qrData := envelope.Wrap(qrBin)

		pngBytes, err := codematrix.EncodePNG(qrData)
		if err != nil {
			return nil, err
		}
		data2D.Write(pngBytes)
	}

	{
		// generate a data matrix with the sheet id
		enc := datamatrix.NewDataMatrixWriter()
		code, err := enc.Encode(p.SerialNumber, gozxing.BarcodeFormat_DATA_MATRIX, 384, 384, nil)
		if err != nil {
			return nil, errors.Join(errors.New("error generating Data Matrix code"), err)
		}

		err = png.Encode(dm, code)
		if err != nil {
			return nil, errors.Join(errors.New("error generating Data Matrix code PNG"), err)
		}
	}

	doc := pdf.GetPdf()
	doc.SetHeaderFuncMode(func() {
		doc.SetY(5)
		doc.SetFont(pdf.MonoFont, "", 10)
		headerLine := fmt.Sprintf(
			"%s: %s - %s",
			PDFHeaderSheetID,
			p.SerialNumber,
			p.CreatedAt.Format(internal.TimeStampFormatPDFHeader),
		)
		if p.Purpose != "" {
			headerLine += fmt.Sprintf(" - %s", p.Purpose)
		}
		doc.CellFormat(0, 10, headerLine,
			"", 0, "C", false, 0, "")

		{
			// add the data matrix code
			doc.RegisterImageReader("dm.png", "PNG", dm)
			imageSize := 5.0
			doc.ImageOptions(
				"dm.png",
				195,
				50,
				imageSize,
				imageSize,
				false,
				gofpdf.ImageOptions{ImageType: "PNG"},
				0,
				"",
			)
		}

		doc.Ln(10)

		if printProductQrCode {
			// add product qr code in upper left corner
			doc.RegisterImageReader("product_link_qr.png", "PNG", productLinkQr)
			imageSize := 15.0
			doc.ImageOptions(
				"product_link_qr.png",
				186,
				11,
				imageSize,
				imageSize,
				false,
				gofpdf.ImageOptions{ImageType: "PNG"},
				0,
				"",
			)
		}
	}, true)
	doc.SetFooterFunc(func() {
		doc.SetY(-15)
		doc.SetFont(pdf.MonoFont, "", 10)
		doc.CellFormat(0, 10, fmt.Sprintf("Page %d/{nb}", doc.PageNo()), "", 0, "R", false, 0, "")
	})
	doc.AddPage()

	{
		// Info text
		doc.SetFont(pdf.TextFont, "B", 16)
		doc.CellFormat(0, 10, PDFHeading, "", 0, "C", false, 0, "")
		doc.Ln(10)

		doc.SetFont(pdf.TextFont, "B", 10)
		doc.CellFormat(0, 5, PDFSectionDescriptionHeading, "", 0, "L", false, 0, "")
		doc.Ln(5)

		doc.SetFont(pdf.TextFont, "", 10)
		doc.MultiCell(0, 5, PDFSectionDescriptionContent, "", "", false)
		doc.Ln(5)

		doc.SetFont(pdf.TextFont, "B", 10)
		doc.CellFormat(0, 5, PDFSectionRepresentationHeading, "", 0, "L", false, 0, "")
		doc.Ln(5)

		doc.SetFont(pdf.TextFont, "", 10)
		representationText := fmt.Sprintf(
			PDFSectionRepresentationContentBase,
			BytesPerLine,
			crc24.CRC24Polynomial,
			crc24.CRC24Initial,
		)
		if p.DataFormat == PaperCryptDataFormatPGP {
			representationText += PDFSectionRepresentationContentGzip
		}
		doc.MultiCell(
			0,
			5,
			representationText,
			"",
			"",
			false,
		)
		doc.Ln(5)

		doc.SetFont(pdf.TextFont, "B", 10)
		doc.CellFormat(0, 5, PDFSectionRecoveryHeading, "", 0, "L", false, 0, "")
		doc.Ln(5)

		doc.SetFont(pdf.TextFont, "", 10)
		recoverInstruction := PDFSectionRecoveryContent
		if no2D {
			recoverInstruction = PDFSectionRecoveryContentNo2D
		}
		doc.MultiCell(0, 5, recoverInstruction, "", "", false)
	}

	// add the qr code
	if !no2D {
		doc.RegisterImageReader("data2D.png", "PNG", data2D)
		imageSize := 167.0
		doc.ImageOptions(
			"data2D.png",
			21,
			5,
			imageSize,
			imageSize,
			true,
			gofpdf.ImageOptions{ImageType: "PNG"},
			0,
			"",
		)
		doc.Ln(50)
	}

	doc.AddPage()
	// print header lines
	doc.SetFont(pdf.MonoFont, "B", DataLineFontSize)
	for _, line := range strings.Split(parts[0], "\n") {
		doc.Cell(0, 5, "# "+line)
		doc.Ln(5)
	}
	doc.Ln(10)

	// print data lines
	dataLines := strings.Split(parts[1], "\n")

	// cut empty lines (should be one at the end)
	filtered := dataLines[:0]
	for _, line := range dataLines {
		if line != "" {
			filtered = append(filtered, line)
		}
	}

	doc.SetFont(pdf.MonoFont, "B", DataLineFontSize)
	for n, line := range filtered {
		// mark every second line with a grey background
		if n%2 == 0 {
			doc.SetFillColor(240, 240, 240)
			doc.Rect(20, doc.GetY(), 166, 5, "F")
		}

		doc.Cell(0, 5, line)
		doc.Ln(5)
	}

	doc.Close()

	var buf bytes.Buffer
	err = doc.Output(&buf)
	if err != nil {
		return nil, errors.Join(errors.New("error generating pdf"), err)
	}

	return buf.Bytes(), nil
}
