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
	PDFSectionRepresentationContent = "Data is written as base 16 (hexadecimal) digits, each representing a half-byte. Two half-bytes are grouped together as a byte, which are then grouped together in lines of %d bytes, where bytes are separated by a space. Each line begins with its line number and a colon, denoting its position and the beginning of the data. Each line is then followed by its CRC-24 checksum. The last line holds the checksum of the entire block. For the checksum algorithm, the polynomial mask 0x%x and initial value 0x%x are used."
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

// SerializeBinary returns the encrypted binary data,
// formatted for restoration
// lines will hold 22 bytes of data, prefaces by the line number, followed by the CRC-24 of the line,
// bytes are printed as two base16 (hex) digits, separated by a space.
// Example:
//
//	1: 00 01 02 03 04 05 06 07 08 09 0A 0B 0C 0D 0E 0F 10 11 12 13 14 15 <CRC-24 of this line>
//	2: ... <CRC-24 of this line>
//
// 10: ... <CRC-24 of this line>
// ...
// n-1: ... <CRC-24 of this line>
// n: <CRC-24 of the block>
//
// See [example.pdf](example.pdf) for an example.
func SerializeBinary(data *[]byte) string {
	lines := math.Ceil(float64(len(*data)) / BytesPerLine)
	lineNumberDigits := int(math.Floor(math.Log10(lines)))

	dataBlock := make([]byte, 0, len(*data)+int(lines)*(lineNumberDigits+1)+1)

	for i := 0; i < len(*data); i += BytesPerLine {
		lineNumber := (i / BytesPerLine) + 1
		lineNumberPadding := lineNumberDigits - int(math.Floor(math.Log10(float64(lineNumber))))

		line := fmt.Sprintf("%s%d: ", string(bytes.Repeat([]byte{' '}, lineNumberPadding)), lineNumber)

		dataLine := make([]byte, 0, BytesPerLine)

		for j := 0; j < BytesPerLine; j++ {
			if i+j >= len(*data) {
				break
			}

			dataLine = append(dataLine, (*data)[i+j])
			line += fmt.Sprintf("%02X ", (*data)[i+j])
		}

		lineCRC24 := Crc24Checksum(dataLine)
		line += fmt.Sprintf("%06X\n", lineCRC24)

		dataBlock = append(dataBlock, []byte(line)...)
	}

	dataCRC24 := Crc24Checksum(*data)
	finalLineNumber := int(math.Ceil(float64(len(*data))/BytesPerLine)) + 1
	finalLinePadding := lineNumberDigits - int(math.Floor(math.Log10(lines+1)))
	dataBlock = append(dataBlock, []byte(fmt.Sprintf("%s%d: %06X\n", string(bytes.Repeat([]byte{' '}, finalLinePadding)), finalLineNumber, dataCRC24))...)

	return string(dataBlock)
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
	pdf.SetCreator("PaperCrypt/"+p.Version, true)
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
		pdf.ImageOptions("qr.png", 20.5, 0, imageSize, imageSize, true, gofpdf.ImageOptions{ImageType: "PNG"}, 0, "")
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
%s: %x
%s: %x
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
%s: %x


%s
`,
			header,
			HeaderFieldHeaderCRC32,
			headerCRC32,
			serializedData)), nil
}
