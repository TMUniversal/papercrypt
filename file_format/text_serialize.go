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
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"hash/crc32"
	"math"
	"sort"
	"strconv"
	"strings"

	"github.com/tmuniversal/papercrypt/v3/crc24"
	"github.com/tmuniversal/papercrypt/v3/internal"
)

const hexDigits = "0123456789ABCDEF"

type lineData struct {
	LineNumber uint32
	Data       []byte
	CRC24      uint32
}

// Lines hold 22 bytes of data, prefaced by the line number, followed by the
// CRC-24 of the line; bytes are printed as two base16 (hex) digits, separated
// by a space. The last line carries the block CRC-24.
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
func SerializeBinary(data *[]byte, bytesPerLine int) string {
	lines := math.Ceil(float64(len(*data)) / float64(bytesPerLine))
	lineNumberDigits := int(math.Floor(math.Log10(lines + 1)))

	// two hex digits plus a space per byte, line-number prefixes, CRCs and newlines
	dataBlock := make([]byte, 0, len(*data)*3+int(lines)*15+8)

	for i := 0; i < len(*data); i += bytesPerLine {
		lineNumber := (i / bytesPerLine) + 1
		lineNumberPadding := lineNumberDigits - int(math.Floor(math.Log10(float64(lineNumber))))

		dataBlock = append(dataBlock, bytes.Repeat([]byte{' '}, lineNumberPadding)...)
		dataBlock = strconv.AppendInt(dataBlock, int64(lineNumber), 10)
		dataBlock = append(dataBlock, ':', ' ')

		dataLine := (*data)[i:min(len(*data), i+bytesPerLine)]
		for _, b := range dataLine {
			dataBlock = append(dataBlock, hexDigits[b>>4], hexDigits[b&0x0f], ' ')
		}

		lineCRC24 := crc24.Checksum(dataLine)
		dataBlock = append(dataBlock,
			hexDigits[lineCRC24>>20&0x0f],
			hexDigits[lineCRC24>>16&0x0f],
			hexDigits[lineCRC24>>12&0x0f],
			hexDigits[lineCRC24>>8&0x0f],
			hexDigits[lineCRC24>>4&0x0f],
			hexDigits[lineCRC24&0x0f],
			'\n',
		)
	}

	dataCRC24 := crc24.Checksum(*data)
	finalLineNumber := max(int(lines+1), min(1, int(lines)))
	dataBlock = strconv.AppendInt(dataBlock, int64(finalLineNumber), 10)
	dataBlock = append(dataBlock, ':', ' ',
		hexDigits[dataCRC24>>20&0x0f],
		hexDigits[dataCRC24>>16&0x0f],
		hexDigits[dataCRC24>>12&0x0f],
		hexDigits[dataCRC24>>8&0x0f],
		hexDigits[dataCRC24>>4&0x0f],
		hexDigits[dataCRC24&0x0f],
		'\n',
	)

	return string(dataBlock)
}

func DeserializeBinary(data *[]byte) ([]byte, error) {
	rawLines := bytes.Split(*data, []byte{'\n'})
	lines := make([][]byte, 0, len(rawLines))
	for _, line := range rawLines {
		if len(line) > 0 {
			lines = append(lines, line)
		}
	}

	result := make([]lineData, 0, len(lines))

	blockCrc := uint32(0)

	for lineIdx := 0; lineIdx < len(lines); lineIdx++ {
		line := lines[lineIdx]
		sep := bytes.Index(line, []byte(": "))
		if sep < 0 {
			return nil, fmt.Errorf("invalid line format: %s", line)
		}

		lineNumber := line[:sep]
		lineNumber = bytes.ReplaceAll(lineNumber, []byte(" "), nil)
		lineNumber = bytes.ReplaceAll(lineNumber, []byte("\t"), nil)

		lineNum, err := strconv.ParseUint(string(lineNumber), 10, 32)
		if err != nil {
			return nil, fmt.Errorf("invalid line number: %s", lineNumber)
		}

		// last line, contains the CRC24 of the data block
		if int64(lineNum) == int64(len(lines)) {
			blockCrc, err = ParseHexUint32(string(line[sep+2:]))
			if err != nil {
				return nil, fmt.Errorf("error parsing block CRC24: %s", line[sep+2:])
			}
			continue
		}

		lineParts := bytes.Split(line[sep+2:], []byte(" "))
		// as lineParts contains sub-arrays of encoded bytes, the length of lineParts is equal to the number of bytes in the line + 1 (for the checksum)
		// a line must never contain no data, this a line must contain at least two parts, one byte and the checksum
		// (the last line, containing only the block checksum, is already handled above)
		if len(lineParts) > DefaultBytesPerLine+1 || len(lineParts) < 2 {
			return nil, fmt.Errorf("unexpected line length: line %d: %s", lineNum, line[sep+2:])
		}

		hexBytes := make([]byte, 0, len(line)-sep-2)
		for _, hb := range lineParts[:len(lineParts)-1] {
			hexBytes = append(hexBytes, hb...)
		}

		decoded := make([]byte, len(hexBytes)/2)
		if _, err := hex.Decode(decoded, hexBytes); err != nil {
			return nil, err
		}

		checksumHex := lineParts[len(lineParts)-1]
		checksumData, err := ParseHexUint32(string(checksumHex))
		if err != nil {
			return nil, fmt.Errorf("error parsing line checksum: %s", checksumHex)
		}

		lineEntry := lineData{
			LineNumber: uint32(lineNum),
			Data:       decoded,
			CRC24:      checksumData,
		}

		if crc24.ValidateCRC24(lineEntry.Data, lineEntry.CRC24) {
			result = append(result, lineEntry)
		} else {
			return nil, fmt.Errorf(
				"invalid line checksum: line %d has checksum %06X, expected %06X",
				lineEntry.LineNumber,
				crc24.Checksum(lineEntry.Data),
				lineEntry.CRC24,
			)
		}
	}

	sort.SliceStable(result, func(i, j int) bool {
		return result[i].LineNumber < result[j].LineNumber
	})

	// Ensure that lines are consecutive, starting at 1: as we sorted the
	// lines, we can just check the first and last line.
	if len(result) == 0 {
		return nil, errors.New("no lines found")
	}

	if result[0].LineNumber != 1 {
		return nil, fmt.Errorf("invalid first line number: %d", result[0].LineNumber)
	}

	// this also ensures that we have all lines, as the last line number must equal the number of lines
	if int64(result[len(result)-1].LineNumber) != int64(len(result)) {
		return nil, fmt.Errorf(
			"invalid last line number: %d",
			result[len(result)-1].LineNumber,
		)
	}

	resultData := make([]byte, 0, len(result)*DefaultBytesPerLine)
	for _, line := range result {
		resultData = append(resultData, line.Data...)
	}

	if !crc24.ValidateCRC24(resultData, blockCrc) {
		return nil, fmt.Errorf(
			"invalid block checksum: expected %06X, found %06X (%d bytes)",
			blockCrc,
			crc24.Checksum(resultData),
			len(resultData),
		)
	}

	return resultData, nil
}

func MarshalBinaryForText(p *PaperCrypt) (string, error) {
	if p.Data == nil {
		return "", errors.New("no data to serialize")
	}

	if len(p.Data) == 0 {
		return "", errors.New("no data to serialize")
	}

	return SerializeBinary(&p.Data, DefaultBytesPerLine), nil
}

func GetText(p *PaperCrypt, lowerCaseEncoding bool) ([]byte, error) {
	header := fmt.Sprintf(
		`%s: %s
%s: %s
%s: %s
%s: %s
%s: %s
%s: %s
%s: %d
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
		p.CreatedAt.Format(internal.TimeStampFormatLong),
		HeaderFieldDataFormat,
		p.DataFormat,
		HeaderFieldContentLength,
		len(p.Data),
		HeaderFieldSHA256,
		base64.StdEncoding.EncodeToString(p.DataSHA256[:]))

	headerCRC32 := crc32.ChecksumIEEE([]byte(header))

	serializedData, err := MarshalBinaryForText(p)
	if err != nil {
		return nil, errors.Join(errors.New("failed to get serialized data"), err)
	}
	if lowerCaseEncoding {
		serializedData = strings.ToLower(serializedData)
	}

	return fmt.Appendf(nil, `%s
%s: %08x


%s
`,
		header,
		HeaderFieldHeaderCRC32,
		headerCRC32,
		serializedData), nil
}

func BytesFromBase64(data string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(data)
}
