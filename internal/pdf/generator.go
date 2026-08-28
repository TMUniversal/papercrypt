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
	dataLineFontSize                        = 11
	pdfSectionRepresentationContentBaseQR   = "The data is contained in a QR code for programmatic recovery, and in text form for recovery without the original software. Text mode prints data in lines of %d bytes, ending with its CRC-24 checksum; the final line holds the checksum of the whole block (polynomial %#x, initial %#x)."
	pdfSectionRepresentationContentBaseNoQR = "The data is printed in text form for manual recovery. Text mode prints data in lines of %d bytes, ending with its CRC-24 checksum; the final line holds the checksum of the whole block (polynomial %#x, initial %#x)."
)

type Mode int

const (
	ModePGPQR Mode = iota
	ModePGPNoQR
	ModeRawQR
	ModeRawNoQR
)

type Config struct {
	HasQR           bool
	SheetSerial     string
	CreatedAt       time.Time
	Purpose         string
	DataQRImage     []byte
	DataMatrixImage []byte
	TextParts       []string

	// BytesPerLine, CRC24Polynomial and CRC24Initial describe the printed text
	// layout and are used when rendering the representation section.
	BytesPerLine    int
	CRC24Polynomial uint32
	CRC24Initial    uint32
}

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

type Generator struct {
	lines lineSet
}

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

func (g *Generator) renderQRCode(doc *gofpdf.Fpdf, cfg Config) {
	doc.RegisterImageReader("data2D.png", "PNG", bytes.NewReader(cfg.DataQRImage))
	doc.ImageOptions(
		"data2D.png", 21, 5, 167, 167,
		true, gofpdf.ImageOptions{ImageType: "PNG"}, 0, "",
	)

	doc.Ln(50)
}

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
