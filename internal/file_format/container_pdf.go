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
	"compress/gzip"
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

	productLinkQr, err := generateProductLinkQR()
	if err != nil {
		return nil, err
	}

	doc := pdf.GetPdf()
	p.renderHeader(doc, dm, productLinkQr)
	p.renderFooter(doc)
	doc.AddPage()

	p.renderPage1Info(doc, no2D)
	if !no2D {
		p.renderQRCode(doc, data2D)
	}

	doc.AddPage()
	renderDataLines(doc, parts)

	doc.Close()

	var buf bytes.Buffer
	if err := doc.Output(&buf); err != nil {
		return nil, errors.Join(errors.New("error generating pdf"), err)
	}

	return buf.Bytes(), nil
}

// encodeDataQR produces the main data QR code PNG by marshalling, compressing, and encoding.
func (p *PaperCrypt) encodeDataQR(no2D bool) (*bytes.Buffer, error) {
	if no2D {
		return nil, nil
	}

	qrBin, err := MarshalBinary(p)
	if err != nil {
		return nil, errors.Join(errors.New("error marshalling PaperCrypt to binary"), err)
	}

	var gzBuf bytes.Buffer
	gz, err := gzip.NewWriterLevel(&gzBuf, gzip.BestCompression)
	if err != nil {
		return nil, errors.Join(errors.New("error creating gzip writer"), err)
	}
	if _, err := gz.Write(qrBin); err != nil {
		return nil, errors.Join(errors.New("error writing gzip data"), err)
	}
	if err := gz.Close(); err != nil {
		return nil, errors.Join(errors.New("error closing gzip writer"), err)
	}

	qrData := envelope.Wrap(gzBuf.Bytes(), envelope.Base45Encoder{})

	pngBytes, err := codematrix.EncodePNG(qrData)
	if err != nil {
		return nil, err
	}

	buf := new(bytes.Buffer)
	buf.Write(pngBytes)
	return buf, nil
}

// generateDataMatrix produces a Data Matrix code PNG encoding the sheet serial number.
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

// generateProductLinkQR produces the product link QR code PNG, or nil if disabled.
func generateProductLinkQR() (*bytes.Buffer, error) {
	if !printProductQrCode {
		return nil, nil
	}

	qrSize := 709
	code, err := qr.Encode(internal.VersionInfo.URL, qr.M, qr.Auto)
	if err != nil {
		return nil, errors.Join(errors.New("error generating product link QR code"), err)
	}

	code, err = barcode.Scale(code, qrSize, qrSize)
	if err != nil {
		return nil, errors.Join(errors.New("error scaling product link QR code"), err)
	}

	converted := image.NewGray(code.Bounds())
	for y := 0; y < code.Bounds().Dy(); y++ {
		for x := 0; x < code.Bounds().Dx(); x++ {
			converted.Set(x, y, code.At(x, y))
		}
	}

	buf := new(bytes.Buffer)
	if err := png.Encode(buf, converted); err != nil {
		return nil, errors.Join(errors.New("error encoding product link QR code PNG"), err)
	}
	return buf, nil
}

// renderHeader configures the PDF header: sheet ID line, data matrix, and optional product QR.
func (p *PaperCrypt) renderHeader(doc *gofpdf.Fpdf, dm, productLinkQr *bytes.Buffer) {
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
		doc.CellFormat(0, 10, headerLine, "", 0, "C", false, 0, "")

		doc.RegisterImageReader("dm.png", "PNG", dm)
		doc.ImageOptions(
			"dm.png", 195, 50, 5, 5,
			false, gofpdf.ImageOptions{ImageType: "PNG"}, 0, "",
		)

		doc.Ln(10)

		if productLinkQr != nil {
			doc.RegisterImageReader("product_link_qr.png", "PNG", productLinkQr)
			doc.ImageOptions(
				"product_link_qr.png", 186, 11, 15, 15,
				false, gofpdf.ImageOptions{ImageType: "PNG"}, 0, "",
			)
		}
	}, true)
}

// renderFooter configures the PDF footer: program name + version (left), page number (right).
func (p *PaperCrypt) renderFooter(doc *gofpdf.Fpdf) {
	doc.SetFooterFunc(func() {
		doc.SetY(-15)
		doc.SetFont(pdf.MonoFont, "", 10)
		doc.CellFormat(
			0, 10, fmt.Sprintf("PaperCrypt %s", internal.VersionInfo.GitVersion),
			"", 0, "L", false, 0, "",
		)
		doc.CellFormat(
			0, 10, fmt.Sprintf("Page %d/{nb}", doc.PageNo()),
			"", 0, "R", false, 0, "",
		)
	})
}

// renderPage1Info writes the title, description, representation, and recovery sections.
func (p *PaperCrypt) renderPage1Info(doc *gofpdf.Fpdf, no2D bool) {
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
	doc.MultiCell(0, 5, representationText, "", "", false)
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

// renderQRCode places the main QR code image on the page.
func (p *PaperCrypt) renderQRCode(doc *gofpdf.Fpdf, data2D *bytes.Buffer) {
	doc.RegisterImageReader("data2D.png", "PNG", data2D)
	doc.ImageOptions(
		"data2D.png", 21, 5, 167, 167,
		true, gofpdf.ImageOptions{ImageType: "PNG"}, 0, "",
	)

	doc.Ln(50)
}

// renderDataLines writes the header lines and hex data lines on page 2.
func renderDataLines(doc *gofpdf.Fpdf, parts []string) {
	doc.SetFont(pdf.MonoFont, "B", DataLineFontSize)
	for _, line := range strings.Split(parts[0], "\n") {
		doc.Cell(0, 5, "# "+line)
		doc.Ln(5)
	}
	doc.Ln(10)

	dataLines := strings.Split(parts[1], "\n")
	filtered := dataLines[:0]
	for _, line := range dataLines {
		if line != "" {
			filtered = append(filtered, line)
		}
	}

	doc.SetFont(pdf.MonoFont, "B", DataLineFontSize)
	for n, line := range filtered {
		if n%2 == 0 {
			doc.SetFillColor(240, 240, 240)
			doc.Rect(20, doc.GetY(), 166, 5, "F")
		}
		doc.Cell(0, 5, line)
		doc.Ln(5)
	}
}
