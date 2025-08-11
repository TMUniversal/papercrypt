/*
 * This file is part of PaperCrypt.
 *
 * PaperCrypt lets you prepare encrypted messages for printing on paper.
 * Copyright (C) 2023-2024 TMUniversal <me@tmuniversal.eu>.
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

// Package internal implements file reading and processing, as well as generation.
package internal

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"hash/crc32"
	"image"
	"image/png"
	"strings"
	"time"

	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/aztec"
	"github.com/boombuler/barcode/qr"
	"github.com/caarlos0/log"
	"github.com/jung-kurt/gofpdf/v2"
	"github.com/makiuchi-d/gozxing"
	"github.com/makiuchi-d/gozxing/datamatrix"
)

const (
	// BytesPerLine denominates the amount of bytes to be encoded per line of the serialized output
	BytesPerLine = 24
	// PdfTextFont gives the typeface as it is named in the PDF
	PdfTextFont = "Text"
	// PdfMonoFont gives the typeface for monospace passages as it is named in the PDF
	PdfMonoFont = "Mono"
	// PdfDataLineFontSize sets the font size of data lines in the PDF [pt]
	PdfDataLineFontSize = 11
)

const printProductQrCode = false

var (
	// PdfTextFontRegularBytes holds the font data for the text typeface, as embedded at compile time. For regular text.
	PdfTextFontRegularBytes []byte
	// PdfTextFontBoldBytes holds the font data for the text typeface, as embedded at compile time. For bold text.
	PdfTextFontBoldBytes []byte
	// PdfTextFontItalicBytes holds the font data for the text typeface, as embedded at compile time. For italic text.
	PdfTextFontItalicBytes []byte
)

var (
	// PdfMonoFontRegularBytes holds the font data for the monospace typeface, as embedded at compile time. For regular text.
	PdfMonoFontRegularBytes []byte
	// PdfMonoFontBoldBytes holds the font data for the monospace typeface, as embedded at compile time. For bold text.
	PdfMonoFontBoldBytes []byte
	// PdfMonoFontItalicBytes holds the font data for the monospace typeface, as embedded at compile time. For italic text.
	PdfMonoFontItalicBytes []byte
)

const (
	// HeaderFieldVersion holds the name of the header field Version. Constant to avoid parsing issues.
	HeaderFieldVersion = "PaperCrypt Version"
	// HeaderFieldSerial holds the name of the header field for the serial number. Constant to avoid parsing issues.
	HeaderFieldSerial = "Content Serial"
	// HeaderFieldPurpose holds the name of the header field Purpose. Constant to avoid parsing issues.
	HeaderFieldPurpose = "Purpose"
	// HeaderFieldComment holds the name of the header field Comment. Constant to avoid parsing issues.
	HeaderFieldComment = "Comment"
	// HeaderFieldDate holds the name of the header field Date. Constant to avoid parsing issues.
	HeaderFieldDate = "Date"
	// HeaderFieldDataFormat holds the name of the header field Data Format. Constant to avoid parsing issues.
	HeaderFieldDataFormat = "Data Format"
	// HeaderFieldContentLength holds the name of the header field Content Length. Constant to avoid parsing issues.
	HeaderFieldContentLength = "Content Length"
	// HeaderFieldCRC24 holds the name of the header field for the CRC-24 checksum. Constant to avoid parsing issues.
	HeaderFieldCRC24 = "Content CRC-24"
	// HeaderFieldCRC32 holds the name of the header field for the CRC-32 checksum. Constant to avoid parsing issues.
	HeaderFieldCRC32 = "Content CRC-32"
	// HeaderFieldSHA256 holds the name of the header field for the SHA-256 checksum. Constant to avoid parsing issues.
	HeaderFieldSHA256 = "Content SHA-256"
	// HeaderFieldHeaderCRC32 holds the name of the header field for the CRC-32 checksum of the header. Constant to avoid parsing issues.
	HeaderFieldHeaderCRC32 = "Header CRC-32"
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
	// PDFSectionRepresentationContent holds the content of the section describing the data representation.
	PDFSectionRepresentationContent = "Data is written as base 16 (hexadecimal) digits, each representing a half-byte. Two half-bytes are grouped together as a byte, which are then grouped together in lines of %d bytes, where bytes are separated by a space. Each line begins with its line number and a colon, denoting its position and the beginning of the data. Each line is then followed by its CRC-24 checksum. The last line holds the checksum of the entire block. For the checksum algorithm, the polynomial mask %#x and initial value %#x are used. Data is compressed using the gzip algorithm."
	// PDFSectionRecoveryHeading holds the title of the section describing how to recover the data.
	PDFSectionRecoveryHeading = "Recovering the data"
	// PDFSectionRecoveryContent holds the content of the section describing how to recover the data.
	PDFSectionRecoveryContent = "Firstly, scan the 2D code, or copy (i.e. type in, or use OCR on) the encrypted data into a computer. Then decrypt it, either using the PaperCrypt CLI, or manually construct the data into a binary file, and decrypt it using OpenPGP-compatible software."
	// PDFSectionRecoveryContentNo2D holds the content of the section describing how to recover the data, if no 2D code is present.
	PDFSectionRecoveryContentNo2D = "Firstly, copy (i.e. type in, or use OCR on) the encrypted data into a computer. Then decrypt it, either using the PaperCrypt CLI, or manually construct the data into a binary file, and decrypt it using OpenPGP-compatible software."
)

var (
	errorParsingHeader     = errors.New("error parsing header")
	errorParsingBody       = errors.New("error parsing body")
	errorValidationFailure = errors.New("validation failure")
)

// PaperCrypt represents a PaperCrypt document.
// It contains metadata about the document, such as its version, serial number, purpose, comment, creation date, and the data itself.
type PaperCrypt struct {
	// Version is the version of papercrypt used to generate the document.
	Version string `json:"v"`

	// DataFormat determines whether the data is raw (although still gzipped), or follows the PGP message format (gzipped).
	DataFormat PaperCryptDataFormat `json:"f"`

	// SerialNumber is the serial number of document, used to identify it. It is generated randomly if not provided.
	SerialNumber string `json:"sn"`

	// Purpose is the purpose of document
	Purpose string `json:"p"`

	// Comment is the comment on document
	Comment string `json:"cm"`

	// CreatedAt is the creation timestamp
	CreatedAt time.Time `json:"ct"`

	// DataCRC24 is the CRC-24 checksum of the encrypted data
	DataCRC24 uint32 `json:"d_c24"`

	// DataCRC32 is the CRC-32 checksum of the encrypted data
	DataCRC32 uint32 `json:"d_c32"`

	// DataSHA256 is the SHA-256 checksum of the encrypted data
	DataSHA256 [32]byte `json:"d_s256"`

	// Data is the contents of the document
	// it can be either of two formats:
	//   a) ASCII armored OpenPGP data, if DataFormat is PGP
	//      the contained message is gzipped before encryption
	//   b) Raw data of any kind, if DataFormat is Raw
	// either way, data is always gzipped after processing
	Data []byte `json:"d"`
}

// MarshalJSON implements the json.Marshaler interface for PaperCrypt.
func (p *PaperCrypt) MarshalJSON() ([]byte, error) { // nosemgrep
	type Alias PaperCrypt
	return json.Marshal(&struct {
		CreatedAt  string `json:"ct"`
		DataSHA256 string `json:"d_s256"`
		*Alias
	}{
		CreatedAt:  p.CreatedAt.Format(TimeStampFormatLong),
		DataSHA256: base64.StdEncoding.EncodeToString(p.DataSHA256[:]),
		Alias:      (*Alias)(p),
	})
}

// UnmarshalJSON implements the json.Unmarshaler interface for PaperCrypt.
func (p *PaperCrypt) UnmarshalJSON(data []byte) error {
	type Alias PaperCrypt
	aux := &struct {
		CreatedAt  string `json:"ct"`
		DataSHA256 string `json:"d_s256"`
		*Alias
	}{
		Alias: (*Alias)(p),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	createdAt, err := time.Parse(TimeStampFormatLong, aux.CreatedAt)
	if err != nil {
		return err
	}
	p.CreatedAt = createdAt

	dataSHA256, err := BytesFromBase64(aux.DataSHA256)
	if err != nil {
		return err
	}
	copy(p.DataSHA256[:], dataSHA256)

	return nil
}

// NewPaperCrypt creates a new paper crypt.
func NewPaperCrypt(
	version string,
	data []byte,
	serialNumber string,
	purpose string,
	comment string,
	createdAt time.Time,
	format PaperCryptDataFormat,
) *PaperCrypt {
	dataCRC24 := Crc24Checksum(data)
	dataCRC32 := crc32.ChecksumIEEE(data)
	dataSHA256 := sha256.Sum256(data)

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
		DataFormat:   format,
	}
}

// GetBinarySerialized returns the binary serialized representation of the PaperCrypt document as a string.
func (p *PaperCrypt) GetBinarySerialized() (string, error) {
	if p.Data == nil {
		return "", errors.New("no data to serialize")
	}

	if len(p.Data) == 0 {
		return "", errors.New("no data to serialize")
	}

	return SerializeBinaryV2(&p.Data), nil
}

// GetDataLength returns the length of the data in bytes as an integer.
func (p *PaperCrypt) GetDataLength() int {
	return len(p.Data)
}

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

		code, err := qr.Encode(VersionInfo.URL, qr.M, qr.Auto)
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
		// for the qr-code, encode the *p as json, then base64 encode it
		qrDataJSON, err := json.Marshal(p)
		if err != nil {
			return nil, errors.Join(errors.New("error marshalling PaperCrypt to JSON"), err)
		}

		// qrSize := 1949 // 165 mm at 300 dpi
		qrSize := 7795 // 165 mm at 1200 dpi
		code, err := aztec.Encode(qrDataJSON, 35, 0)
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

		err = png.Encode(data2D, converted)
		if err != nil {
			return nil, errors.Join(errors.New("error generating 2D code PNG"), err)
		}
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

	pdf := getPdf()
	pdf.SetHeaderFuncMode(func() {
		pdf.SetY(5)
		pdf.SetFont(PdfMonoFont, "", 10)
		headerLine := fmt.Sprintf(
			"%s: %s - %s",
			PDFHeaderSheetID,
			p.SerialNumber,
			p.CreatedAt.Format(TimeStampFormatPDFHeader),
		)
		if p.Purpose != "" {
			headerLine += fmt.Sprintf(" - %s", p.Purpose)
		}
		pdf.CellFormat(0, 10, headerLine,
			"", 0, "C", false, 0, "")

		{
			// add the data matrix code
			pdf.RegisterImageReader("dm.png", "PNG", dm)
			imageSize := 5.0
			pdf.ImageOptions(
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

		pdf.Ln(10)

		if printProductQrCode {
			// add product qr code in upper left corner
			pdf.RegisterImageReader("product_link_qr.png", "PNG", productLinkQr)
			imageSize := 15.0
			pdf.ImageOptions(
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
	pdf.SetFooterFunc(func() {
		pdf.SetY(-15)
		pdf.SetFont(PdfMonoFont, "", 10)
		pdf.CellFormat(0, 10, fmt.Sprintf("Page %d/{nb}", pdf.PageNo()), "", 0, "R", false, 0, "")
	})
	pdf.AddPage()

	{
		// Info text
		pdf.SetFont(PdfTextFont, "B", 16)
		pdf.CellFormat(0, 10, PDFHeading, "", 0, "C", false, 0, "")
		pdf.Ln(10)

		pdf.SetFont(PdfTextFont, "B", 10)
		pdf.CellFormat(0, 5, PDFSectionDescriptionHeading, "", 0, "L", false, 0, "")
		pdf.Ln(5)

		pdf.SetFont(PdfTextFont, "", 10)
		pdf.MultiCell(0, 5, PDFSectionDescriptionContent, "", "", false)
		pdf.Ln(5)

		pdf.SetFont(PdfTextFont, "B", 10)
		pdf.CellFormat(0, 5, PDFSectionRepresentationHeading, "", 0, "L", false, 0, "")
		pdf.Ln(5)

		pdf.SetFont(PdfTextFont, "", 10)
		pdf.MultiCell(
			0,
			5,
			fmt.Sprintf(
				PDFSectionRepresentationContent,
				BytesPerLine,
				CRC24Polynomial,
				CRC24Initial,
			),
			"",
			"",
			false,
		)
		pdf.Ln(5)

		pdf.SetFont(PdfTextFont, "B", 10)
		pdf.CellFormat(0, 5, PDFSectionRecoveryHeading, "", 0, "L", false, 0, "")
		pdf.Ln(5)

		pdf.SetFont(PdfTextFont, "", 10)
		recoverInstruction := PDFSectionRecoveryContent
		if no2D {
			recoverInstruction = PDFSectionRecoveryContentNo2D
		}
		pdf.MultiCell(0, 5, recoverInstruction, "", "", false)
	}

	// add the qr code
	if !no2D {
		pdf.RegisterImageReader("data2D.png", "PNG", data2D)
		imageSize := 167.0
		pdf.ImageOptions(
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
		pdf.Ln(50)
	}

	pdf.AddPage()
	// print header lines
	pdf.SetFont(PdfMonoFont, "B", PdfDataLineFontSize)
	for _, line := range strings.Split(parts[0], "\n") {
		pdf.Cell(0, 5, "# "+line)
		pdf.Ln(5)
	}
	pdf.Ln(10)

	// print data lines
	dataLines := strings.Split(parts[1], "\n")

	// cut empty lines (should be one at the end)
	filtered := dataLines[:0]
	for _, line := range dataLines {
		if line != "" {
			filtered = append(filtered, line)
		}
	}

	pdf.SetFont(PdfMonoFont, "B", PdfDataLineFontSize)
	for n, line := range filtered {
		// mark every second line with a grey background
		if n%2 == 0 {
			pdf.SetFillColor(240, 240, 240)
			pdf.Rect(20, pdf.GetY(), 166, 5, "F")
		}

		pdf.Cell(0, 5, line)
		pdf.Ln(5)
	}

	pdf.Close()

	var buf bytes.Buffer
	err = pdf.Output(&buf)
	if err != nil {
		return nil, errors.Join(errors.New("error generating pdf"), err)
	}

	return buf.Bytes(), nil
}

// GetText returns the text representation of the paper crypt.
func (p *PaperCrypt) GetText(lowerCaseEncoding bool) ([]byte, error) {
	header := fmt.Sprintf(
		`%s: %s
%s: %s
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
		p.CreatedAt.Format(TimeStampFormatLong),
		HeaderFieldDataFormat,
		p.DataFormat,
		HeaderFieldContentLength,
		p.GetDataLength(),
		HeaderFieldCRC24,
		p.DataCRC24,
		HeaderFieldCRC32,
		p.DataCRC32,
		HeaderFieldSHA256,
		base64.StdEncoding.EncodeToString(p.DataSHA256[:]))

	headerCRC32 := crc32.ChecksumIEEE([]byte(header))

	serializedData, err := p.GetBinarySerialized()
	if err != nil {
		return nil, errors.Join(errors.New("failed to get serialized data"), err)
	}
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

func getPdf() *gofpdf.Fpdf {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetCreator("PaperCrypt/"+VersionInfo.GitVersion, true)
	pdf.SetTextRenderingMode(4)
	pdf.SetTopMargin(20)
	pdf.SetLeftMargin(20)
	pdf.SetRightMargin(20)
	pdf.SetAutoPageBreak(true, 15)
	pdf.AliasNbPages("")

	pdf.AddUTF8FontFromBytes(PdfTextFont, "", PdfTextFontRegularBytes)
	pdf.AddUTF8FontFromBytes(PdfTextFont, "B", PdfTextFontBoldBytes)
	pdf.AddUTF8FontFromBytes(PdfTextFont, "I", PdfTextFontItalicBytes)

	pdf.AddUTF8FontFromBytes(PdfMonoFont, "", PdfMonoFontRegularBytes)
	pdf.AddUTF8FontFromBytes(PdfMonoFont, "B", PdfMonoFontBoldBytes)
	pdf.AddUTF8FontFromBytes(PdfMonoFont, "I", PdfMonoFontItalicBytes)

	return pdf
}

func newFieldNotPresentError(field string) error {
	return fmt.Errorf("`%s` not present in header", field)
}

// Decode decodes and, if the data was encrypted with PaperCrypt (data format is PaperCryptDataFormatPGP),
// decrypts the data, returning the original binary data.
func (p *PaperCrypt) Decode(passphrase []byte) ([]byte, error) {
	data := p.Data
	if p.DataFormat == PaperCryptDataFormatPGP {
		// 1. Decompress
		gzipReader, err := gzip.NewReader(bytes.NewReader(p.Data))
		if err != nil {
			return nil, errors.Join(errors.New("error creating gzip reader"), err)
		}

		decompressed := new(bytes.Buffer)
		if _, err := decompressed.ReadFrom(gzipReader); err != nil {
			return nil, errors.Join(errors.New("error reading from gzip reader"), err)
		}
		if err := gzipReader.Close(); err != nil {
			return nil, errors.Join(errors.New("error closing gzip reader"), err)
		}

		pgpMessage := crypto.NewPGPMessage(decompressed.Bytes())

		// 9. Decrypt secretContents
		decryptedMessage, err := crypto.DecryptMessageWithPassword(pgpMessage, passphrase)
		if err != nil {
			return nil, errors.Join(errors.New("error decrypting secret contents"), err)
		}

		data = decryptedMessage.GetBinary()
	}

	// 10. Decompress content
	gzipReader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, errors.Join(errors.New("error creating gzip reader"), err)
	}

	decompressed := new(bytes.Buffer)
	if _, err := decompressed.ReadFrom(gzipReader); err != nil {
		return nil, errors.Join(errors.New("error reading from gzip reader"), err)
	}
	if err := gzipReader.Close(); err != nil {
		return nil, errors.Join(errors.New("error closing gzip reader"), err)
	}

	return decompressed.Bytes(), nil
}

// TextToHeaderMap converts a byte slice containing text headers into a map of header fields.
// Each header line should be in the format "Key: Value", with the key being the header field name
// and the value being the header field value.
// The function trims the "# " prefix from header lines, which is present in the serialized text format.
func TextToHeaderMap(text []byte) (map[string]string, error) {
	headers := make(map[string]string)

	headerLines := bytes.Split(text, []byte("\n"))
	for _, headerLine := range headerLines {
		headerLineSplit := bytes.SplitN(headerLine, []byte(": "), 2)
		if len(headerLineSplit) != 2 {
			return nil, errors.Join(
				errorParsingHeader,
				fmt.Errorf("error parsing header line: %s", headerLine),
			)
		}

		key := string(headerLineSplit[0])
		key = strings.TrimPrefix(key, "# ")

		headers[key] = string(headerLineSplit[1])
	}

	return headers, nil
}

// SplitTextHeaderAndBody splits the given byte slice, which should be a PaperCrypt document, into a header and body section.
func SplitTextHeaderAndBody(data []byte) ([]byte, []byte, error) {
	dataSplit := bytes.SplitN(data, []byte("\n\n\n"), 2)
	if len(dataSplit) != 2 {
		return nil, nil, errors.New(
			"header not discernible, header and content should be separated by two empty lines",
		)
	}
	return dataSplit[0], dataSplit[1], nil
}

// DeserializeV2Text deserializes a PaperCrypt document from a byte slice containing text.
// It expects the text to be in the format defined by PaperCrypt version 2. (PaperCryptContainerVersionMajor2).
func DeserializeV2Text(
	data []byte,
	ignoreVersionMismatch bool,
	ignoreChecksumMismatch bool,
) (*PaperCrypt, error) {
	paperCryptFileContents := NormalizeLineEndings(data)

	headersSection, bodySection, err := SplitTextHeaderAndBody(paperCryptFileContents)
	if err != nil {
		return nil, errors.Join(errorParsingHeader, err)
	}

	headers, err := TextToHeaderMap(headersSection)
	if err != nil {
		return nil, errors.Join(errorParsingHeader, err)
	}

	// Debug: print headers
	log.WithField("headers", headers).Debug("Read headers")

	// 4. Run Header Validation
	versionLine, ok := headers[HeaderFieldVersion]
	if !ok {
		if !ignoreVersionMismatch {
			return nil, errors.Join(errorParsingHeader, newFieldNotPresentError(HeaderFieldVersion))
		}

		log.Warn(Warning("PaperCrypt Version not present in header."))
	}

	majorVersion := PaperCryptContainerVersionFromString(versionLine)
	if !ignoreVersionMismatch &&
		(majorVersion != PaperCryptContainerVersionMajor2 && majorVersion != PaperCryptContainerVersionDevel) {
		return nil, errors.Join(
			errorParsingHeader,
			fmt.Errorf("unsupported PaperCrypt version '%s'", versionLine),
		)
	}

	// Validate Header checksum
	{
		headerCrc, ok := headers[HeaderFieldHeaderCRC32]
		if !ok {
			if !ignoreChecksumMismatch {
				return nil, errors.Join(
					errorParsingHeader,
					newFieldNotPresentError(HeaderFieldHeaderCRC32),
				)
			}

			log.Warn(Warning("Header CRC-32 not present in header"))
		}

		headerCrc = strings.ToLower(headerCrc)
		headerCrc = strings.ReplaceAll(headerCrc, "0x", "")
		headerCrc = strings.ReplaceAll(headerCrc, " ", "")
		headerCrc32, err := ParseHexUint32(headerCrc)
		if err != nil {
			return nil, errors.Join(errorParsingHeader, errors.New("invalid CRC-32 format"), err)
		}

		headerWithoutCrc := bytes.ReplaceAll(headersSection, []byte("# "), []byte{})
		headerWithoutCrc = bytes.ReplaceAll(
			headerWithoutCrc,
			[]byte("\n"+HeaderFieldHeaderCRC32+": "+headers[HeaderFieldHeaderCRC32]),
			[]byte{},
		)

		if !ValidateCRC32(headerWithoutCrc, headerCrc32) {
			if !ignoreChecksumMismatch {
				return nil, errors.Join(
					errorParsingHeader,
					errorValidationFailure,
					errors.New(
						"header CRC-32 mismatch: expected "+headers[HeaderFieldHeaderCRC32]+", got "+fmt.Sprintf(
							"%x",
							crc32.ChecksumIEEE(headerWithoutCrc),
						),
					),
				)
			}

			log.Warn(Warning("Header CRC-32 mismatch!"))
		}
	}

	var dataFormat PaperCryptDataFormat
	{
		dataFormatString, ok := headers[HeaderFieldDataFormat]
		if !ok {
			return nil, errors.Join(
				errorParsingHeader,
				newFieldNotPresentError(HeaderFieldDataFormat),
			)
		}

		log.Debugf("Data Format: %s", dataFormatString)

		dataFormat = PaperCryptDataFormatFromString(dataFormatString)
	}

	var pgpMessage *crypto.PGPMessage
	var body []byte
	body, err = DeserializeBinary(&bodySection)
	if err != nil {
		return nil, errors.Join(errorParsingBody, err)
	}

	switch dataFormat {
	case PaperCryptDataFormatPGP:
		pgpMessage = crypto.NewPGPMessage(body)
		body = pgpMessage.GetBinary()
	case PaperCryptDataFormatRaw:
		// do nothing
	default:
		return nil, errors.Join(errorParsingBody, errors.New("unsupported data format"))
	}

	// 5. Verify Body Hashes

	// 5.1 Verify Content Length
	bodyLength, ok := headers[HeaderFieldContentLength]
	if !ok {
		return nil, errors.Join(errorParsingBody, newFieldNotPresentError(HeaderFieldContentLength))
	}

	if fmt.Sprint(len(body)) != bodyLength {
		return nil, errors.Join(
			errorValidationFailure,
			fmt.Errorf(
				"`%s` mismatch: expected %s, got %d",
				HeaderFieldContentLength,
				bodyLength,
				len(body),
			),
		)
	}

	// 5.2 Verify CRC-32
	bodyCrc32, ok := headers[HeaderFieldCRC32]
	if !ok {
		return nil, errors.Join(errorValidationFailure, newFieldNotPresentError(HeaderFieldCRC32))
	}

	bodyCrc32Uint32, err := ParseHexUint32(bodyCrc32)
	if err != nil {
		return nil, errors.Join(errorParsingBody, err)
	}

	if !ValidateCRC32(body, bodyCrc32Uint32) {
		if !ignoreChecksumMismatch {
			return nil, errors.Join(
				errorValidationFailure,
				fmt.Errorf("`%s` mismatch", HeaderFieldCRC32),
			)
		}

		log.Warn(Warning("Content CRC-32 mismatch!"))
	}

	// 5.3 Verify CRC-24
	bodyCrc24, ok := headers[HeaderFieldCRC24]
	if !ok {
		return nil, errors.Join(errorParsingBody, newFieldNotPresentError(HeaderFieldCRC24))
	}

	bodyCrc24Uint32, err := ParseHexUint32(bodyCrc24)
	if err != nil {
		return nil, errors.Join(errorParsingBody, err)
	}

	if !ValidateCRC24(body, bodyCrc24Uint32) {
		if !ignoreChecksumMismatch {
			return nil, errors.Join(
				errorValidationFailure,
				fmt.Errorf("`%s` mismatch", HeaderFieldCRC24),
			)
		}

		log.Warn(Warning("Content CRC-24 mismatch!"))
	}

	// 5.4 Verify SHA-256
	bodySha256, ok := headers[HeaderFieldSHA256]
	if !ok {
		return nil, errors.Join(errorParsingBody, newFieldNotPresentError(HeaderFieldSHA256))
	}

	bodySha256Bytes, err := BytesFromBase64(bodySha256)
	if err != nil {
		return nil, errors.Join(errorParsingBody, err)
	}

	actualSha256 := sha256.Sum256(body)
	if !bytes.Equal(actualSha256[:], bodySha256Bytes) {
		if !ignoreChecksumMismatch {
			return nil, errors.Join(
				errorValidationFailure,
				fmt.Errorf("`%s` mismatch", HeaderFieldSHA256),
			)
		}

		log.Warn(Warning("Content SHA-256 mismatch!"))
	}

	// 6. Construct PaperCrypt object
	headerDate, ok := headers[HeaderFieldDate]
	if !ok {
		log.Warn(Warning("Date not present in header!"))
	}

	timestamp, err := time.Parse(TimeStampFormatLong, headerDate)
	if err != nil {
		return nil, errors.Join(errors.New("invalid date format"), err)
	}

	// we don't need to pass the checksums, as they are already verified
	// and will just be recalculated
	paperCrypt := NewPaperCrypt(
		versionLine,
		body,
		headers[HeaderFieldSerial],
		headers[HeaderFieldPurpose],
		headers[HeaderFieldComment],
		timestamp,
		dataFormat,
	)

	// 7. Serialize PaperCrypt object
	_, err = json.MarshalIndent(paperCrypt, "", "  ")
	if err != nil {
		return nil, errors.Join(errors.New("error encoding JSON"), err)
	}
	log.WithField("json", paperCrypt).Debug("Serialized PaperCrypt document")

	return paperCrypt, nil
}
