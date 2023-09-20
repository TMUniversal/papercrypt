/*
 * This file is part of PaperCrypt.
 *
 * PaperCrypt lets you prepare encrypted messages for printing on paper.
 * Copyright (C) 2023 TMUniversal <me@tmuniversal.eu>.
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
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"hash/crc32"
	"image/png"
	"math"
	"math/big"
	"strings"
	"time"

	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/jung-kurt/gofpdf/v2"
	"github.com/makiuchi-d/gozxing"
	"github.com/makiuchi-d/gozxing/datamatrix"
	"github.com/makiuchi-d/gozxing/qrcode"
	"github.com/pkg/errors"
)

const (
	BytesPerLine = 22 // As is done in paperkey (https://www.jabberwocky.com/software/paperkey/)
	PdfTextFont  = "Times"
	PdfMonoFont  = "Courier"
)

const (
	HeaderFieldVersion              = "PaperCrypt Version"
	HeaderFieldSerial               = "Content Serial"
	HeaderFieldPurpose              = "Purpose"
	HeaderFieldComment              = "Comment"
	HeaderFieldDate                 = "Date"
	HeaderFieldLength               = "Content Length"
	HeaderFieldCRC24                = "Content CRC-24"
	HeaderFieldCRC32                = "Content CRC-32"
	HeaderFieldSHA256               = "Content SHA-256"
	HeaderFieldHeaderCRC32          = "Header CRC-32"
	PDFHeaderSheetId                = "Sheet ID"
	PDFHeading                      = "PaperCrypt Recovery Sheet"
	PDFSectionDescriptionHeading    = "What is this?"
	PDFSectionDescriptionContent    = "This is a PaperCrypt recovery sheet. It contains encrypted data, its own creation date, purpose, and a comment, as well as an identifier. This sheet is intended to help recover the original information, in case it is lost or destroyed."
	PDFSectionRepresentationHeading = "Binary Data Representation"
	PDFSectionRepresentationContent = "Data is written as base 16 (hexadecimal) digits, each representing a half-byte. Two half-bytes are grouped together as a byte, which are then grouped together in lines of %d bytes, where bytes are separated by a space. Each line begins with its line number and a colon, denoting its position and the beginning of the data. Each line is then followed by its CRC-24 checksum. The last line holds the checksum of the entire block. For the checksum algorithm, the polynomial mask %#x and initial value %#x are used."
	PDFSectionRecoveryHeading       = "Recovering the data"
	PDFSectionRecoveryContent       = "Firstly, scan the QR code, or copy (i.e. type it in, or use OCR) the encrypted data into a computer. Then decrypt it, either using the PaperCrypt CLI, or manually construct the data into a binary file, and decrypt it using OpenPGP-compatible software."
	PDFSectionRecoveryContentNoQR   = "Firstly, copy (i.e. type it in, or use OCR) the encrypted data into a computer. Then decrypt it, either using the PaperCrypt CLI, or manually construct the data into a binary file, and decrypt it using OpenPGP-compatible software."
)

type PaperCrypt struct {
	// Version is the version of papercrypt used to generate the document.
	Version string

	// Data is the encrypted data as a PGP message
	Data *crypto.PGPMessage

	// SerialNumber is the serial number of document, used to identify it. It is generated randomly if not provided.
	SerialNumber string

	// Purpose is the purpose of document
	Purpose string

	// Comment is the comment on document
	Comment string

	// CreatedAt is the creation timestamp
	CreatedAt time.Time

	// DataCRC24 is the CRC-24 checksum of the encrypted data
	DataCRC24 uint32

	// DataCRC32 is the CRC-32 checksum of the encrypted data
	DataCRC32 uint32

	// DataSHA256 is the SHA-256 checksum of the encrypted data
	DataSHA256 [32]byte
}

// NewPaperCrypt creates a new paper crypt
func NewPaperCrypt(version string, data *crypto.PGPMessage, serialNumber string, purpose string, comment string, createdAt time.Time) *PaperCrypt {
	binData := data.GetBinary()

	dataCRC24 := Crc24Checksum(binData)
	dataCRC32 := crc32.ChecksumIEEE(binData)
	dataSHA256 := sha256.Sum256(binData)

	return &PaperCrypt{
		Version:      version,
		Data:         data,
		SerialNumber: serialNumber,
		Purpose:      purpose,
		Comment:      comment,
		CreatedAt:    createdAt,
		DataCRC24:    dataCRC24,
		DataCRC32:    dataCRC32,
		DataSHA256:   dataSHA256,
	}
}

func (p *PaperCrypt) GetBinary() []byte {
	return p.Data.GetBinary()
}

func (p *PaperCrypt) GetBinarySerialized() string {
	data := p.GetBinary()
	return SerializeBinary(&data)
}

func (p *PaperCrypt) GetLength() int {
	return len(p.GetBinary())
}

// GetPDF returns the binary representation of the paper crypt
// The PDF will be generated to include some basic information about papercrypt,
// some metadata, optionally a QR-Code, and the encrypted data.
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
// and, next to the markdown information, a QR code containing the encrypted data.
func (p *PaperCrypt) GetPDF(noQR bool, lowerCaseEncoding bool) ([]byte, error) {
	text, err := p.GetText(lowerCaseEncoding)
	if err != nil {
		return nil, errors.Errorf("error getting text content: %s", err)
	}

	// split at 2 empty lines, to get the header and the data
	parts := strings.Split(string(text), "\n\n\n")
	if len(parts) != 2 {
		return nil, errors.Errorf("error splitting text content into header and data")
	}

	qr := new(bytes.Buffer)
	dm := new(bytes.Buffer)

	if !noQR {
		// for the qr-code, encode the *p as json, then base64 encode it
		qrDataJson, err := json.Marshal(p)
		if err != nil {
			return nil, errors.Errorf("error marshalling PaperCrypt to json: %s", err)
		}
		codeData := string(qrDataJson)

		enc := qrcode.NewQRCodeWriter()
		encoderHints := make(map[gozxing.EncodeHintType]interface{})
		encoderHints[gozxing.EncodeHintType_ERROR_CORRECTION] = "Q"
		encoderHints[gozxing.EncodeHintType_MARGIN] = 0
		qrSize := int(math.Ceil(165 * 9)) // 165 mm
		code, err := enc.Encode(codeData, gozxing.BarcodeFormat_QR_CODE, qrSize, qrSize, encoderHints)
		if err != nil {
			return nil, errors.Errorf("error generating QR code: %s", err)
		}

		err = png.Encode(qr, code)
		if err != nil {
			return nil, errors.Errorf("error generating QR code PNG: %s", err)
		}
	}

	{
		// generate a data matrix with the sheet id
		enc := datamatrix.NewDataMatrixWriter()
		code, err := enc.Encode(p.SerialNumber, gozxing.BarcodeFormat_DATA_MATRIX, 256, 256, nil)
		if err != nil {
			return nil, errors.Errorf("error generating Data Matrix code: %s", err)
		}

		err = png.Encode(dm, code)
		if err != nil {
			return nil, errors.Errorf("error generating Data Matrix code PNG: %s", err)
		}
	}

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetCreator("PaperCrypt/"+VersionInfo.Version, true)
	pdf.SetTopMargin(20)
	pdf.SetLeftMargin(20)
	pdf.SetRightMargin(20)
	pdf.SetAutoPageBreak(true, 15)
	pdf.SetHeaderFuncMode(func() {
		pdf.SetY(5)
		pdf.SetFont(PdfMonoFont, "", 10)
		headerLine := fmt.Sprintf("%s: %s - %s", PDFHeaderSheetId, p.SerialNumber, p.CreatedAt.Format("2006-01-02 15:04 -0700"))
		if p.Purpose != "" {
			headerLine += fmt.Sprintf(" - %s", p.Purpose)
		}
		pdf.CellFormat(0, 10, headerLine,
			"", 0, "C", false, 0, "")

		{
			// add the data matrix code
			pdf.RegisterImageReader("dm.png", "PNG", dm)
			imageSize := 5.0
			pdf.ImageOptions("dm.png", 195, 50, imageSize, imageSize, false, gofpdf.ImageOptions{ImageType: "PNG"}, 0, "")
		}

		pdf.Ln(10)
	}, true)
	pdf.SetFooterFunc(func() {
		pdf.SetY(-15)
		pdf.SetFont(PdfMonoFont, "", 10)
		pdf.CellFormat(0, 10, fmt.Sprintf("Page %d/{nb}", pdf.PageNo()), "", 0, "R", false, 0, "")
	})
	pdf.AliasNbPages("")
	pdf.AddPage()

	{
		// Info text
		pdf.SetFont(PdfTextFont, "", 16)
		pdf.CellFormat(0, 10, PDFHeading, "", 0, "C", false, 0, "")
		pdf.Ln(10)
		pdf.SetFont(PdfTextFont, "", 12)
		// enter the markdown information
		pdf.CellFormat(0, 5, PDFSectionDescriptionHeading, "", 0, "L", false, 0, "")
		pdf.Ln(5)
		pdf.SetFont(PdfTextFont, "", 10)
		pdf.MultiCell(0, 5, PDFSectionDescriptionContent, "", "", false)
		pdf.Ln(5)
		pdf.SetFont(PdfTextFont, "", 12)
		pdf.CellFormat(0, 5, PDFSectionRepresentationHeading, "", 0, "L", false, 0, "")
		pdf.Ln(5)
		pdf.SetFont(PdfTextFont, "", 10)
		pdf.MultiCell(0, 5, fmt.Sprintf(PDFSectionRepresentationContent, BytesPerLine, CRC24Polynomial, CRC24Initial), "", "", false)
		pdf.Ln(5)
		pdf.SetFont(PdfTextFont, "", 12)
		pdf.CellFormat(0, 5, PDFSectionRecoveryHeading, "", 0, "L", false, 0, "")
		pdf.Ln(5)
		pdf.SetFont(PdfTextFont, "", 10)
		recoverInstruction := PDFSectionRecoveryContent
		if noQR {
			recoverInstruction = PDFSectionRecoveryContentNoQR
		}
		pdf.MultiCell(0, 5, recoverInstruction, "", "", false)
		pdf.Ln(10)
	}

	// add the qr code
	if !noQR {
		pdf.RegisterImageReader("qr.png", "PNG", qr)
		imageSize := 167.0
		pdf.ImageOptions("qr.png", 20, 0, imageSize, imageSize, true, gofpdf.ImageOptions{ImageType: "PNG"}, 0, "")
		pdf.Ln(50)
	}

	// print header lines
	pdf.SetFont(PdfMonoFont, "B", 10)
	for _, line := range strings.Split(parts[0], "\n") {
		pdf.Cell(0, 5, "# "+line)
		pdf.Ln(5)
	}
	pdf.Ln(10)

	// print data lines
	dataLines := strings.Split(parts[1], "\n")
	pdf.SetFont(PdfMonoFont, "B", 10)
	for _, line := range dataLines {
		pdf.Cell(0, 5, line)
		pdf.Ln(5)
	}

	pdf.Close()

	var buf bytes.Buffer
	err = pdf.Output(&buf)
	if err != nil {
		return nil, errors.Errorf("error generating pdf: %s", err)
	}

	return buf.Bytes(), nil
}

// GetText returns the text representation of the paper crypt
func (p *PaperCrypt) GetText(lowerCaseEncoding bool) ([]byte, error) {
	header := fmt.Sprintf(
		`%s: %s
%s: %s
%s: %s
%s: %s
%s: %s
%s: %d
%s: %06x
%s: %08x
%s: %s`,
		HeaderFieldVersion,
		p.Version,
		HeaderFieldSerial,
		p.SerialNumber,
		HeaderFieldPurpose,
		p.Purpose,
		HeaderFieldComment,
		p.Comment,
		HeaderFieldDate,
		// format time with nanosecond precision
		// Sat, 12 Aug 2023 17:33:20.123456789
		p.CreatedAt.Format("Mon, 02 Jan 2006 15:04:05.000000000 MST"),
		HeaderFieldLength,
		p.GetLength(),
		HeaderFieldCRC24,
		p.DataCRC24,
		HeaderFieldCRC32,
		p.DataCRC32,
		HeaderFieldSHA256,
		base64.StdEncoding.EncodeToString(p.DataSHA256[:]))

	headerCRC32 := crc32.ChecksumIEEE([]byte(header))

	serializedData := p.GetBinarySerialized()
	if lowerCaseEncoding {
		serializedData = strings.ToLower(serializedData)
	}

	return []byte(
		fmt.Sprintf(`%s
%s: %08x


%s
`,
			header,
			HeaderFieldHeaderCRC32,
			headerCRC32,
			serializedData)), nil
}

func GeneratePassphraseSheetPDF(seed int64, words []string) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetCreator("PaperCrypt/"+VersionInfo.Version, true)
	pdf.SetTopMargin(20)
	pdf.SetLeftMargin(20)
	pdf.SetRightMargin(20)
	pdf.SetAutoPageBreak(true, 15)

	dm := new(bytes.Buffer)
	dmDims := [2]int{}
	encodedSeed := base64.StdEncoding.EncodeToString(big.NewInt(seed).Bytes())
	{
		// generate a data matrix with the seed
		enc := datamatrix.NewDataMatrixWriter()

		// create the code without dimensions to get the width and height required for the code
		initial, err := enc.Encode(encodedSeed, gozxing.BarcodeFormat_DATA_MATRIX, 0, 0, nil)
		if err != nil {
			return nil, errors.Wrap(err, "error generating Data Matrix code")
		}

		dmDims[0] = initial.GetWidth()
		dmDims[1] = initial.GetHeight()

		// create the code at 8x scale
		code, err := enc.Encode(encodedSeed, gozxing.BarcodeFormat_DATA_MATRIX, 8*dmDims[0], 8*dmDims[1], nil)
		if err != nil {
			return nil, errors.Wrap(err, "error generating Data Matrix code")
		}

		err = png.Encode(dm, code)
		if err != nil {
			return nil, errors.Wrap(err, "error generating Data Matrix code PNG")
		}
	}

	date := time.Now().Format("2006-01-02 15:04 -0700")

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

			pdf.ImageOptions("dm.png", 170, 7, imageWidth, imageHeight, false, gofpdf.ImageOptions{ImageType: "PNG"}, 0, "")
		}

		pdf.Ln(10)
	}, true)
	pdf.SetFooterFunc(func() {
		pdf.SetY(-15)
		pdf.SetFont(PdfMonoFont, "", 10)
		pdf.CellFormat(0, 10, fmt.Sprintf("Page %d/{nb}", pdf.PageNo()), "", 0, "R", false, 0, "")
	})
	pdf.AliasNbPages("")
	pdf.AddPage()

	{
		// Info text
		pdf.SetFont(PdfTextFont, "", 16)
		pdf.CellFormat(0, 10, "PaperCrypt Passphrase Sheet", "", 0, "C", false, 0, "")
		pdf.Ln(10)

		pdf.SetFont(PdfTextFont, "", 10)
		pdf.MultiCell(0, 5, `To create a passphrase or password with this sheet, start by choosing words on this sheet, preferably following these guidelines:
    1. Choose between 6 and 24 words,
    2. Do not choose words in order.`, "", "L", false)
		pdf.Ln(2)
		pdf.MultiCell(0, 5, `You can regenerate this sheet using the seed printed at the top of each page, which is also encoded in the Data Matrix at the top.`, "", "L", false)

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

		pdf.MultiCell(0, 5, fmt.Sprintf("This sheet contains %d words, giving %d (~2^%d) possible combinations of 6 distinct words, %d (~2^%d) of 12 words, and %d (~2^%d) of 24 words.",
			len(words),
			sixOf135,
			int(math.Round(sixOf135Power)),
			twelveOf135,
			int(math.Round(twelveOf135Power)),
			twentyFourOf135,
			int(math.Round(twentyFourOf135Power)),
		), "", "L", false)
	}

	pdf.Close()
	var buf bytes.Buffer
	err := pdf.Output(&buf)
	if err != nil {
		return nil, errors.Wrap(err, "error generating PDF")
	}
	return buf.Bytes(), nil
}
