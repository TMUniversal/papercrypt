package util

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"hash/crc32"
	"math"
	"strings"
	"time"

	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/jung-kurt/gofpdf/v2"
	"github.com/pkg/errors"
	"github.com/skip2/go-qrcode"
)

const (
	BytesPerLine = 22 // As is done in paperkey (https://www.jabberwocky.com/software/paperkey/)
	PdfTextFont  = "Times"
	PdfMonoFont  = "Courier"
)

type PaperCrypt struct {
	// Version is the version of the paper crypt
	Version string

	// Data is the encrypted data
	Data *crypto.PGPMessage

	// SerialNumber is the serial number of the paper crypt
	SerialNumber string

	// Purpose is the purpose of the paper crypt
	Purpose string

	// Comment is the comment on the paper crypt
	Comment string

	// CreatedAt is the creation timestamp
	CreatedAt time.Time
}

// NewPaperCrypt creates a new paper crypt
func NewPaperCrypt(version string, data *crypto.PGPMessage, serialNumber string, purpose string, comment string, createdAt time.Time) *PaperCrypt {
	return &PaperCrypt{
		Version:      version,
		Data:         data,
		SerialNumber: serialNumber,
		Purpose:      purpose,
		Comment:      comment,
		CreatedAt:    createdAt,
	}
}

// SerializeBinary returns the encrypted binary data,
// formatted for restoration
// lines will hold 22 bytes of data, prefaces by the line number, followed by the CRC-24 of the line,
// bytes are printed as two base16 (hex) digits, separated by a space.
// Example:
//
//	1: 00 01 02 03 04 05 06 07 08 09 0A 0B 0C 0D 0E 0F 10 11 12 13 <CRC-24 of this line>
//	2: ...
//
// 10: ...
// ...
// n-1: ... <CRC-24 of this line>
// n: <CRC-24 of the block>
func SerializeBinary(data *[]byte) string {
	lines := math.Ceil(float64(len(*data)) / BytesPerLine)
	lineNumberDigits := int(math.Floor(math.Log10(lines))) + 1

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
	finalLineNumber := int(math.Ceil(float64(len(*data)) / BytesPerLine))
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
func (p *PaperCrypt) GetPDF(asciiArmor, noQR bool, lowerCaseEncoding bool) ([]byte, error) {
	text, err := p.GetText(asciiArmor, lowerCaseEncoding)
	if err != nil {
		return nil, errors.Errorf("error getting text content: %s", err)
	}

	// split at 2 empty lines, to get the header and the data
	parts := strings.Split(string(text), "\n\n\n")
	if len(parts) != 2 {
		return nil, errors.Errorf("error splitting text content into header and data")
	}

	var qr []byte

	if !noQR {

		// for the qr-code, encode the *p as json, then base64 encode it
		// finally format it as a URL: papercrypt://d/?data=<base64-encoded-json>
		qrDataJson, err := json.Marshal(p)
		if err != nil {
			return nil, errors.Errorf("error marshalling PaperCrypt to json: %s", err)
		}

		qrDataBase64 := new(bytes.Buffer)
		base64encoder := base64.NewEncoder(base64.URLEncoding, qrDataBase64)
		_, err = base64encoder.Write(qrDataJson)
		if err != nil {
			return nil, errors.Errorf("error base64-encoding PaperCrypt json: %s", err)
		}

		qrData := fmt.Sprintf("papercrypt://d/?data=%s", qrDataBase64.String())

		qr, err = qrcode.Encode(qrData, qrcode.Highest, 1024)
		if err != nil {
			return nil, errors.Errorf("error generating QR code: %s", err)
		}
	}

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetCreator("PaperCrypt/"+p.Version, true)
	pdf.SetTopMargin(30)
	pdf.SetHeaderFuncMode(func() {
		pdf.SetY(5)
		pdf.SetFont(PdfMonoFont, "", 10)
		pdf.Cell(80, 0, "")
		pdf.CellFormat(30, 10, fmt.Sprintf("Sheet ID: %s - %s - Purpose: %s", p.SerialNumber, p.CreatedAt.Format("2006-01-02 15:04 -0700"), p.Purpose),
			"", 0, "C", false, 0, "")
		pdf.Ln(10)
	}, true)
	pdf.SetFooterFunc(func() {
		pdf.SetY(-15)
		pdf.SetFont(PdfMonoFont, "", 10)
		pdf.CellFormat(0, 10, fmt.Sprintf("Page %d/{nb}", pdf.PageNo()), "", 0, "R", false, 0, "")
	})
	pdf.AliasNbPages("")
	pdf.AddPage()

	pdf.SetFont(PdfTextFont, "", 16)
	pdf.CellFormat(0, 10, "PaperCrypt Recovery Sheet", "", 0, "C", false, 0, "")
	pdf.Ln(10)
	pdf.SetFont(PdfTextFont, "", 12)
	// enter the markdown information
	pdf.CellFormat(0, 5, "What is this?", "", 0, "L", false, 0, "")
	pdf.Ln(5)
	pdf.SetFont(PdfTextFont, "", 10)
	pdf.MultiCell(0, 5, `This is a PaperCrypt recovery sheet. It contains encrypted data, its own creation date, purpose, and a comment, as well as an identifier. This sheet is intended to help recover the original information, in case it is lost or destroyed.`, "", "", false)
	pdf.Ln(5)
	pdf.SetFont(PdfTextFont, "", 12)
	pdf.CellFormat(0, 5, "Recovering the data", "", 0, "L", false, 0, "")
	pdf.Ln(5)
	pdf.SetFont(PdfTextFont, "", 10)
	pdf.MultiCell(0, 5, `1. Scan the QR code, or copy (i.e. type it in, or use OCR) the encrypted data into a computer.
2. Decrypt using the PaperCrypt CLI, or manually construct the data into a binary file, and decrypt it using OpenPGP-compatible software.`, "", "", false)
	pdf.Ln(10)

	// add the qr code
	if !noQR {
		pdf.RegisterImageReader("qr.png", "PNG", bytes.NewReader(qr))
		pdf.ImageOptions("qr.png", 30, 0, 150, 150, true, gofpdf.ImageOptions{ImageType: "PNG"}, 0, "")
		pdf.Ln(50)
	}

	pdf.SetFont(PdfMonoFont, "", 10)
	pdf.Cell(0, 5, fmt.Sprintf("PaperCrypt %s", p.Version))
	pdf.Ln(5)
	pdf.Cell(0, 5, fmt.Sprintf("Creation Date: %s", p.CreatedAt.Format("Mon, 02 Jan 2006 15:04:05.000000000 MST")))
	pdf.Ln(5)
	pdf.Cell(0, 5, fmt.Sprintf("Comment: %s", p.Comment))
	pdf.Ln(10)

	// loop over the data lines, and add them to the pdf
	dataLines := strings.Split(parts[1], "\n")
	pdf.SetFont(PdfMonoFont, "", 10)
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
func (p *PaperCrypt) GetText(asciiArmor bool, lowerCaseEncoding bool) ([]byte, error) {
	var data []byte

	if asciiArmor {
		//stringData, err := p.Data.GetArmoredWithCustomHeaders(fmt.Sprintf("PaperCrypt/%s (https://github.com/TMUniversal/PaperCrypt), https://openpgp.org/", p.Version), constants.ArmorHeaderVersion)
		stringData, err := p.Data.GetArmored()
		if err != nil {
			return nil, errors.Errorf("error getting armored data: %s", err)
		}

		data = []byte(stringData)
	} else {
		data = p.Data.GetBinary()
	}

	dataCRC32 := crc32.ChecksumIEEE(data)
	dataSHA256 := sha256.Sum256(data)

	header := fmt.Sprintf(
		`PaperCrypt/%s
Content Serial: %s (Base32)
Purpose: %s
Comment: %s
Date: %s
Content CRC-32: %x
Content SHA-256: %x
Content Length: %d bytes`,
		p.Version,
		p.SerialNumber,
		p.Purpose,
		p.Comment,
		// format time with nanosecond precision
		// Sat, 12 Aug 2023 17:33:20.123456789
		p.CreatedAt.Format("Mon, 02 Jan 2006 15:04:05.000000000 MST"),
		dataCRC32,
		dataSHA256,
		len(data))

	headerCRC32 := crc32.ChecksumIEEE([]byte(header))

	serializedData := string(data)
	if !asciiArmor {
		serializedData = SerializeBinary(&data)
		if lowerCaseEncoding {
			serializedData = strings.ToLower(serializedData)
		}
	}

	return []byte(
		fmt.Sprintf(`%s
Header CRC-32: %x


%s
`,
			header,
			headerCRC32,
			serializedData)), nil
}
