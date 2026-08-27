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
	PDFSectionDescriptionContent = "This is a PaperCrypt recovery sheet. It stores your data together with its identifier, creation date, purpose, and comment. Keep it safe, so the original data can be recovered if it is ever lost or damaged."
	// PDFSectionRepresentationHeading holds the title of the section describing the data representation.
	PDFSectionRepresentationHeading = "Binary Data Representation"
	// PDFSectionRepresentationContentBase holds the content of the section describing the data representation.
	PDFSectionRepresentationContentBase = "The data is written as base 16 (hexadecimal) digits, each representing half a byte. Bytes appear in lines of %d, separated by spaces. Every line starts with its number and a colon and ends with its CRC-24 checksum; the final line holds the checksum of the whole block, computed with the polynomial mask %#x and initial value %#x."
	// PDFSectionRepresentationContentPGP is the PGP-specific suffix appended for encrypted data.
	PDFSectionRepresentationContentPGP = " The data on this sheet is gzipped and encrypted with the encryption passphrase, so it cannot be read directly."
	// PDFSectionRepresentationContentRaw is the raw-specific suffix appended for unencrypted data.
	PDFSectionRepresentationContentRaw = " The data on this sheet is stored exactly as-is, without compression or encryption, so it can be read directly from the hex digits."
	// PDFSectionRecoveryHeading holds the title of the section describing how to recover the data.
	PDFSectionRecoveryHeading = "Recovering the data"
	// PDFSectionRecoveryContent holds the content of the section describing how to recover the data.
	PDFSectionRecoveryContent = "Scan the QR code, or copy the data into a computer by typing it in or using OCR. Then decrypt it with the encryption passphrase."
	// PDFSectionRecoveryContentNo2D holds the content of the section describing how to recover the data, if no 2D code is present.
	PDFSectionRecoveryContentNo2D = "No QR code is printed on this sheet. Copy the data into a computer by typing it in or using OCR, then decrypt it with the encryption passphrase."
	// PDFSectionRecoveryContentRaw holds the content of the section describing how to recover raw data.
	PDFSectionRecoveryContentRaw = "Scan the QR code, or copy the data into a computer by typing it in or using OCR. The data is stored as-is, so reassembling the bytes reproduces the original file."
	// PDFSectionRecoveryContentRawNo2D holds the content of the section describing how to recover raw data, if no 2D code is present.
	PDFSectionRecoveryContentRawNo2D = "No QR code is printed on this sheet. Copy the data into a computer by typing it in or using OCR; the bytes reproduce the original file as-is, with no decryption needed."
	// PDFSectionDocumentationContent holds the content of the section on the final page pointing to the documentation.
	PDFSectionDocumentationContent = "This sheet was generated with PaperCrypt. For the documentation, source code, and more guidance on recovering the encoded data, scan the code to visit the project website."
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

	doc := pdf.GetPdf()
	p.renderHeader(doc, dm)
	p.renderFooter(doc)
	doc.AddPage()

	p.renderPage1Info(doc, no2D)
	if !no2D {
		p.renderQRCode(doc, data2D)
	}

	doc.AddPage()
	renderDataLines(doc, parts)

	if err := p.renderDocumentation(doc); err != nil {
		return nil, err
	}

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

// generateProductLinkQR produces the documentation link QR code PNG.
func generateProductLinkQR() (*bytes.Buffer, error) {
	// Uppercase the URL so every character is in the AlphaNumeric charset,
	// producing a denser, smaller QR code.
	value := strings.ToUpper(internal.VersionInfo.URL)

	qrSize := 709
	code, err := qr.Encode(value, qr.M, qr.AlphaNumeric)
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

// renderHeader configures the PDF header: sheet ID line and the data matrix code.
func (p *PaperCrypt) renderHeader(doc *gofpdf.Fpdf, dm *bytes.Buffer) {
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
	switch p.DataFormat {
	case PaperCryptDataFormatRaw:
		representationText += PDFSectionRepresentationContentRaw
	case PaperCryptDataFormatPGP:
		representationText += PDFSectionRepresentationContentPGP
	}
	doc.MultiCell(0, 5, representationText, "", "", false)
	doc.Ln(5)

	doc.SetFont(pdf.TextFont, "B", 10)
	doc.CellFormat(0, 5, PDFSectionRecoveryHeading, "", 0, "L", false, 0, "")
	doc.Ln(5)

	doc.SetFont(pdf.TextFont, "", 10)
	var recoverInstruction string
	if p.DataFormat == PaperCryptDataFormatRaw {
		if no2D {
			recoverInstruction = PDFSectionRecoveryContentRawNo2D
		} else {
			recoverInstruction = PDFSectionRecoveryContentRaw
		}
	} else if no2D {
		recoverInstruction = PDFSectionRecoveryContentNo2D
	} else {
		recoverInstruction = PDFSectionRecoveryContent
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

// renderDocumentation writes the documentation note and link QR code at the bottom
// of the final page. The QR code sits at the left, with the note rendered to its right.
func (p *PaperCrypt) renderDocumentation(doc *gofpdf.Fpdf) error {
	productLinkQr, err := generateProductLinkQR()
	if err != nil {
		return err
	}

	const (
		// A4 width is 210mm and height 297mm
		pageBottom = 283.5
		qrSize     = 15.0
		gap        = 3.5
		noteLineH  = 3.5
		leftMargin = 21.0
	)

	// Width available to the right of the QR code, up to the right margin.
	noteWidth := 210 - leftMargin - leftMargin - qrSize - gap - gap

	doc.SetFont(pdf.TextFont, "", 8)
	noteLines := len(doc.SplitLines([]byte(PDFSectionDocumentationContent), noteWidth))
	noteHeight := float64(noteLines) * noteLineH

	leftColHeight := qrSize + gap
	sectionHeight := leftColHeight
	if noteHeight > sectionHeight {
		sectionHeight = noteHeight
	}

	startY := pageBottom - sectionHeight

	if doc.GetY()+gap > startY {
		doc.AddPage()
	}

	doc.RegisterImageReader("product_link_qr.png", "PNG", productLinkQr)
	doc.ImageOptions(
		"product_link_qr.png", leftMargin, startY, qrSize, qrSize,
		false, gofpdf.ImageOptions{ImageType: "PNG"}, 0, "",
	)

	doc.SetXY(leftMargin+qrSize+gap, startY)
	doc.MultiCell(noteWidth, noteLineH, PDFSectionDocumentationContent, "", "", false)

	return nil
}
