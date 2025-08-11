/*
 * This file is part of PaperCrypt.
 *
 * PaperCrypt lets you prepare encrypted messages for printing on paper.
 * Copyright (C) 2024 TMUniversal <me@tmuniversal.eu>.
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

package internal

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"image/png"
	"math"
	"math/big"
	"math/rand"
	"time"

	"github.com/caarlos0/log"
	"github.com/jung-kurt/gofpdf/v2"
	"github.com/makiuchi-d/gozxing"
	"github.com/makiuchi-d/gozxing/datamatrix"
)

// GenerateFromSeed selects a number of words from the given list
// using a seeded, non-cryptographic pseudo-random generator.
func GenerateFromSeed(seed int64, amount int, wordList *[]string) ([]string, error) {
	if amount < 1 {
		return nil, errors.New("amount must be greater than 0")
	}
	// 2. Generate random numbers
	gen := rand.New(rand.NewSource(seed))

	words := make([]string, amount)
	for i := 0; i < amount; i++ {
		random := gen.Intn(len(*wordList)) // Intn returns [0, n) (excludes n)
		w := (*wordList)[random]

		if SliceHasString(words, w) {
			// if the word is already in the slice, try again
			log.WithField("word", w).
				WithField("index", i).
				Warn("Duplicate word appeared, trying again...")
			i--
			continue
		}

		words[i] = w
	}
	return words, nil
}

// GeneratePassphraseSheetPDF creates a PDF file displaying the given words in three columns, the seed in the header.
func GeneratePassphraseSheetPDF(seed int64, words []string) ([]byte, error) {
	pdf := getPdf()

	dm := new(bytes.Buffer)
	dmDims := [2]int{}
	encodedSeed := base64.StdEncoding.EncodeToString(big.NewInt(seed).Bytes())
	{
		// generate a data matrix with the seed
		enc := datamatrix.NewDataMatrixWriter()

		// create the code without dimensions to get the width and height required for the code
		initial, err := enc.Encode(encodedSeed, gozxing.BarcodeFormat_DATA_MATRIX, 0, 0, nil)
		if err != nil {
			return nil, errors.Join(errors.New("error generating Data Matrix code"), err)
		}

		dmDims[0] = initial.GetWidth()
		dmDims[1] = initial.GetHeight()

		// create the code at 8x scale
		code, err := enc.Encode(
			encodedSeed,
			gozxing.BarcodeFormat_DATA_MATRIX,
			8*dmDims[0],
			8*dmDims[1],
			nil,
		)
		if err != nil {
			return nil, errors.Join(errors.New("error generating Data Matrix code"), err)
		}

		err = png.Encode(dm, code)
		if err != nil {
			return nil, errors.Join(errors.New("error generating Data Matrix code PNG"), err)
		}
	}

	date := time.Now().Format(TimeStampFormatPDFHeader)

	pdf.SetHeaderFuncMode(func() {
		pdf.SetY(5)
		pdf.SetFont(PdfMonoFont, "", 10)
		headerLine := fmt.Sprintf("Seed: %s - %s", encodedSeed, date)
		pdf.CellFormat(0, 10, headerLine,
			"", 0, "C", false, 0, "")

		{
			// add the data matrix code
			pdf.RegisterImageReader("dm.png", "PNG", dm)
			width := float64(dmDims[0])
			height := float64(dmDims[1])

			// code is like to come out as 16px*16px (2x2 modules), but can also be 8px*32px (1x4 modules)
			scale := 0.5 // so we choose scale = 0.5 to get 8mm*8mm, or 4mm*16mm

			imageWidth := width * scale
			imageHeight := height * scale

			pdf.ImageOptions(
				"dm.png",
				170,
				7,
				imageWidth,
				imageHeight,
				false,
				gofpdf.ImageOptions{ImageType: "PNG"},
				0,
				"",
			)
		}

		pdf.Ln(10)
	}, true)
	pdf.SetFooterFunc(func() {
		pdf.SetY(-15)
		pdf.SetFont(PdfMonoFont, "", 10)
		pdf.CellFormat(0, 10, fmt.Sprintf("Page %d/{nb}", pdf.PageNo()), "", 0, "R", false, 0, "")
	})
	pdf.AddPage()

	{
		// Info text
		pdf.SetFont(PdfTextFont, "B", 16)
		pdf.CellFormat(0, 10, "PaperCrypt Passphrase Sheet", "", 0, "C", false, 0, "")
		pdf.Ln(10)

		pdf.SetFont(PdfTextFont, "", 10)
		pdf.MultiCell(
			0,
			5,
			`To create a passphrase or password with this sheet, start by choosing words on this sheet, preferably following these guidelines:
    1. Choose between 6 and 24 words,
    2. Do not choose words in order.`,
			"",
			"L",
			false,
		)
		pdf.Ln(2)
		pdf.MultiCell(
			0,
			5,
			`You can regenerate this sheet using the seed printed at the top of each page, which is also encoded in the Data Matrix at the top.`,
			"",
			"L",
			false,
		)

		pdf.Ln(3)
	}

	tableWidth := 170.0 // 210mm - 20mm left margin - 20mm right margin
	columnWidth := tableWidth / 3

	// Print table data
	for i := 0; i < len(words); i += 3 {
		for j := 0; j < 3; j++ {
			if i+j < len(words) {
				// print index
				pdf.SetFont(PdfMonoFont, "", 10)
				pdf.CellFormat(10, 10, fmt.Sprintf("%d", i+j+1), "", 0, "R", false, 0, "")
				// print word
				pdf.SetFont(PdfMonoFont, "B", 14)
				pdf.CellFormat(columnWidth, 10, words[i+j], "", 0, "L", false, 0, "")
			}
		}
		pdf.Ln(-1)
	}

	{
		// amount of possible combinations
		pdf.Ln(10)

		// calculate n choose k (n! / (k! * (n-k)!)
		// for 6 words, 12, and 24 of 135 words
		sixOf135 := big.NewInt(0).Binomial(int64(len(words)), 6)
		twelveOf135 := big.NewInt(0).Binomial(int64(len(words)), 12)
		twentyFourOf135 := big.NewInt(0).Binomial(int64(len(words)), 24)

		// find the nearest power of 2
		sixOf135Power := math.Log2(float64(sixOf135.Int64()))
		twelveOf135Power := math.Log2(float64(twelveOf135.Int64()))
		twentyFourOf135Power := math.Log2(float64(twentyFourOf135.Int64()))

		pdf.SetFont(PdfTextFont, "", 10)

		pdf.MultiCell(
			0,
			5,
			fmt.Sprintf(
				"This sheet contains %d words, giving %d (~2^%d) possible combinations of 6 distinct words, %d (~2^%d) of 12 words, and %d (~2^%d) of 24 words.",
				len(words),
				sixOf135,
				int(math.Round(sixOf135Power)),
				twelveOf135,
				int(math.Round(twelveOf135Power)),
				twentyFourOf135,
				int(math.Round(twentyFourOf135Power)),
			),
			"",
			"L",
			false,
		)
	}

	pdf.Close()
	var buf bytes.Buffer
	err := pdf.Output(&buf)
	if err != nil {
		return nil, errors.Join(errors.New("error generating PDF"), err)
	}
	return buf.Bytes(), nil
}
