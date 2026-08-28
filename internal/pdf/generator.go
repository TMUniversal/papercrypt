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

package pdf

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/png"
	"strings"
	"time"

	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/qr"
	"github.com/jung-kurt/gofpdf/v2"
	"github.com/tmuniversal/papercrypt/v3/internal"
)

const (
	// dataLineFontSize sets the font size of data lines in the PDF [pt]
	dataLineFontSize = 11
	// pdfSectionRepresentationContentBaseQR describes the data representation for
	// sheets that carry a QR code. The data is contained in the QR code, and also
	// printed in text form for manual recovery.
	pdfSectionRepresentationContentBaseQR = "The data is contained in a QR code for programmatic recovery, and in text form for recovery without the original software. Text mode prints data in lines of %d bytes, ending with its CRC-24 checksum; the final line holds the checksum of the whole block (polynomial %#x, initial %#x)."
	// pdfSectionRepresentationContentBaseNoQR describes the data representation for
	// sheets without a QR code: the data is only available in printed text form.
	pdfSectionRepresentationContentBaseNoQR = "The data is printed in text form for manual recovery. Text mode prints data in lines of %d bytes, ending with its CRC-24 checksum; the final line holds the checksum of the whole block (polynomial %#x, initial %#x)."
)

// Mode identifies the kind of recovery sheet a Generator produces.
type Mode int

const (
	// ModePGPQR produces a recovery sheet for encrypted (PGP) data with a QR code.
	ModePGPQR Mode = iota
	// ModePGPNoQR produces a recovery sheet for encrypted (PGP) data without a QR code.
	ModePGPNoQR
	// ModeRawQR produces a recovery sheet for raw, unencrypted data with a QR code.
	ModeRawQR
	// ModeRawNoQR produces a recovery sheet for raw, unencrypted data without a QR code.
	ModeRawNoQR
)

// Config carries the sheet content that is independent of the generator mode.
type Config struct {
	// HasQR reports whether a data QR code should be rendered on the first page.
	HasQR bool
	// SheetSerial is the identifier printed in the header.
	SheetSerial string
	// CreatedAt is the creation timestamp printed in the header.
	CreatedAt time.Time
	// Purpose is an optional short description printed in the header.
	Purpose string

	// DataQRImage is the PNG of the data QR code; only used when HasQR is set.
	DataQRImage []byte
	// DataMatrixImage is the PNG of the header Data Matrix code.
	DataMatrixImage []byte

	// TextParts holds the header and data text lines, as split from the text representation.
	TextParts []string

	// BytesPerLine, CRC24Polynomial and CRC24Initial describe the printed text layout
	// and are used when rendering the representation section.
	BytesPerLine    int
	CRC24Polynomial uint32
	CRC24Initial    uint32
}

// lineSet holds all text lines that vary between recovery-sheet modes.
type lineSet struct {
	headerSheetID         string
	heading               string
	descriptionHeading    string
	descriptionContent    string
	representationHeading string
	representationBase    string
	representationSuffix  string
	recoveryHeading       string
	recoveryContent       string
	documentationContent  string
}

// defaultLines returns the text lines that are shared by all recovery-sheet modes.
func defaultLines() lineSet {
	return lineSet{
		headerSheetID:         "Sheet ID",
		heading:               "PaperCrypt Recovery Sheet",
		descriptionHeading:    "What is this?",
		descriptionContent:    "This is a PaperCrypt recovery sheet. It stores your data together with its identifier, creation date, purpose, and comment. Keep it safe, so the original data can be recovered if it is ever lost or damaged.",
		representationHeading: "Binary Data Representation",
		recoveryHeading:       "Recovering the data",
		documentationContent:  "This sheet was generated with PaperCrypt. For the documentation, source code, and more guidance on recovering the encoded data, scan the code to visit the project website.",
	}
}

// Generator renders a PaperCrypt recovery sheet PDF using mode-specific text lines.
type Generator struct {
	lines lineSet
}

// New returns a Generator for the requested recovery-sheet mode.
func New(mode Mode) *Generator {
	g := &Generator{}
	switch mode {
	case ModePGPQR:
		g.lines = pgpQRLines()
	case ModePGPNoQR:
		g.lines = pgpNoQRLines()
	case ModeRawQR:
		g.lines = rawQRLines()
	case ModeRawNoQR:
		g.lines = rawNoQRLines()
	}
	return g
}

// Render produces the PDF bytes for the configured sheet.
func (g *Generator) Render(cfg Config) ([]byte, error) {
	if len(cfg.TextParts) != 2 {
		return nil, errors.New("error splitting text content into header and data")
	}

	doc := GetPdf()
	g.renderHeader(doc, cfg)
	g.renderFooter(doc)
	doc.AddPage()

	g.renderPage1Info(doc, cfg)
	if cfg.HasQR {
		g.renderQRCode(doc, cfg)
	}

	doc.AddPage()
	renderDataLines(doc, cfg)

	if err := g.renderDocumentation(doc); err != nil {
		return nil, err
	}

	doc.Close()

	var buf bytes.Buffer
	if err := doc.Output(&buf); err != nil {
		return nil, errors.Join(errors.New("error generating pdf"), err)
	}

	return buf.Bytes(), nil
}

// renderHeader configures the PDF header: sheet ID line and the data matrix code.
func (g *Generator) renderHeader(doc *gofpdf.Fpdf, cfg Config) {
	doc.SetHeaderFuncMode(func() {
		doc.SetY(5)
		doc.SetFont(MonoFont, "", 10)

		headerLine := fmt.Sprintf(
			"%s: %s - %s",
			g.lines.headerSheetID,
			cfg.SheetSerial,
			cfg.CreatedAt.Format(internal.TimeStampFormatPDFHeader),
		)
		if cfg.Purpose != "" {
			headerLine += fmt.Sprintf(" - %s", cfg.Purpose)
		}

		doc.CellFormat(0, 10, headerLine, "", 0, "C", false, 0, "")

		doc.RegisterImageReader("dm.png", "PNG", bytes.NewReader(cfg.DataMatrixImage))
		doc.ImageOptions(
			"dm.png", 195, 50, 5, 5,
			false, gofpdf.ImageOptions{ImageType: "PNG"}, 0, "",
		)

		doc.Ln(10)
	}, true)
}

// renderFooter configures the PDF footer: program name + version (left), page number (right).
func (g *Generator) renderFooter(doc *gofpdf.Fpdf) {
	doc.SetFooterFunc(func() {
		doc.SetY(-15)
		doc.SetFont(MonoFont, "", 10)
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
func (g *Generator) renderPage1Info(doc *gofpdf.Fpdf, cfg Config) {
	doc.SetFont(TextFont, "B", 16)
	doc.CellFormat(0, 10, g.lines.heading, "", 0, "C", false, 0, "")
	doc.Ln(10)

	doc.SetFont(TextFont, "B", 10)
	doc.CellFormat(0, 5, g.lines.descriptionHeading, "", 0, "L", false, 0, "")
	doc.Ln(5)

	doc.SetFont(TextFont, "", 10)
	doc.MultiCell(0, 5, g.lines.descriptionContent, "", "", false)
	doc.Ln(5)

	doc.SetFont(TextFont, "B", 10)
	doc.CellFormat(0, 5, g.lines.representationHeading, "", 0, "L", false, 0, "")
	doc.Ln(5)

	doc.SetFont(TextFont, "", 10)
	representationText := fmt.Sprintf(
		g.lines.representationBase,
		cfg.BytesPerLine,
		cfg.CRC24Polynomial,
		cfg.CRC24Initial,
	)
	representationText += g.lines.representationSuffix
	doc.MultiCell(0, 5, representationText, "", "", false)
	doc.Ln(5)

	doc.SetFont(TextFont, "B", 10)
	doc.CellFormat(0, 5, g.lines.recoveryHeading, "", 0, "L", false, 0, "")
	doc.Ln(5)

	doc.SetFont(TextFont, "", 10)
	doc.MultiCell(0, 5, g.lines.recoveryContent, "", "", false)
}

// renderQRCode places the main QR code image on the page.
func (g *Generator) renderQRCode(doc *gofpdf.Fpdf, cfg Config) {
	doc.RegisterImageReader("data2D.png", "PNG", bytes.NewReader(cfg.DataQRImage))
	doc.ImageOptions(
		"data2D.png", 21, 5, 167, 167,
		true, gofpdf.ImageOptions{ImageType: "PNG"}, 0, "",
	)

	doc.Ln(50)
}

// renderDataLines writes the header lines and hex data lines on page 2.
func renderDataLines(doc *gofpdf.Fpdf, cfg Config) {
	doc.SetFont(MonoFont, "B", dataLineFontSize)
	for _, line := range strings.Split(cfg.TextParts[0], "\n") {
		doc.Cell(0, 5, "# "+line)
		doc.Ln(5)
	}
	doc.Ln(10)

	dataLines := strings.Split(cfg.TextParts[1], "\n")
	filtered := dataLines[:0]
	for _, line := range dataLines {
		if line != "" {
			filtered = append(filtered, line)
		}
	}

	doc.SetFont(MonoFont, "B", dataLineFontSize)
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
func (g *Generator) renderDocumentation(doc *gofpdf.Fpdf) error {
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

	doc.SetFont(TextFont, "", 8)
	noteLines := len(doc.SplitLines([]byte(g.lines.documentationContent), noteWidth))
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
	doc.MultiCell(noteWidth, noteLineH, g.lines.documentationContent, "", "", false)

	doc.SetXY(leftMargin+qrSize+gap, startY+noteHeight+5.0)
	doc.MultiCell(noteWidth, noteLineH, internal.VersionInfo.URL, "", "", false)

	return nil
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
