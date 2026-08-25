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

// Package pdf provides PDF generation utilities for PaperCrypt documents.
package pdf

import (
	"github.com/jung-kurt/gofpdf/v2"
	"github.com/tmuniversal/papercrypt/v3/internal"
)

const (
	// TextFont gives the typeface as it is named in the PDF
	TextFont = "Text"
	// MonoFont gives the typeface for monospace passages as it is named in the PDF
	MonoFont = "Mono"
)

var (
	// TextFontRegularBytes holds the font data for the text typeface, as embedded at compile time. For regular text.
	TextFontRegularBytes []byte
	// TextFontBoldBytes holds the font data for the text typeface, as embedded at compile time. For bold text.
	TextFontBoldBytes []byte
	// TextFontItalicBytes holds the font data for the text typeface, as embedded at compile time. For italic text.
	TextFontItalicBytes []byte
)

var (
	// MonoFontRegularBytes holds the font data for the monospace typeface, as embedded at compile time. For regular text.
	MonoFontRegularBytes []byte
	// MonoFontBoldBytes holds the font data for the monospace typeface, as embedded at compile time. For bold text.
	MonoFontBoldBytes []byte
	// MonoFontItalicBytes holds the font data for the monospace typeface, as embedded at compile time. For italic text.
	MonoFontItalicBytes []byte
)

// GetPdf returns a new PDF instance configured with PaperCrypt fonts and layout settings.
func GetPdf() *gofpdf.Fpdf {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetCreator("PaperCrypt/"+internal.VersionInfo.GitVersion, true)
	pdf.SetTextRenderingMode(4)
	pdf.SetTopMargin(20)
	pdf.SetLeftMargin(20)
	pdf.SetRightMargin(20)
	pdf.SetAutoPageBreak(true, 15)
	pdf.AliasNbPages("")

	pdf.AddUTF8FontFromBytes(TextFont, "", TextFontRegularBytes)
	pdf.AddUTF8FontFromBytes(TextFont, "B", TextFontBoldBytes)
	pdf.AddUTF8FontFromBytes(TextFont, "I", TextFontItalicBytes)

	pdf.AddUTF8FontFromBytes(MonoFont, "", MonoFontRegularBytes)
	pdf.AddUTF8FontFromBytes(MonoFont, "B", MonoFontBoldBytes)
	pdf.AddUTF8FontFromBytes(MonoFont, "I", MonoFontItalicBytes)

	return pdf
}
