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

package internal

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"strings"
)

type LineData struct {
	LineNumber uint32
	Data       []byte
	CRC24      uint32
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
	lineNumberDigits := int(math.Floor(math.Log10(lines + 1)))

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
	finalLineNumber := max(int(lines+1), min(1, int(lines)))
	dataBlock = append(dataBlock, []byte(fmt.Sprintf("%d: %06X\n", finalLineNumber, dataCRC24))...)

	return string(dataBlock)
}

func DeserializeBinary(data *[]byte) ([]byte, error) {
	rawLines := bytes.Split(*data, []byte{'\n'})
	lines := make([][]byte, 0)

	// filter out empty lines
	for _, line := range rawLines {
		if len(line) > 0 {
			lines = append(lines, line)
		}
	}

	result := make([]LineData, 0)

	blockCrc := uint32(0)

	// 1. Parse lines, validate line checksums
	for _, line := range lines {
		parts := bytes.SplitN(line, []byte(": "), 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid line format: %s", line)
		}

		lineNumber := strings.ReplaceAll(string(parts[0]), " ", "")
		lineNumber = strings.ReplaceAll(lineNumber, "\t", "")

		if lineNumber == fmt.Sprint(len(lines)) {
			// last line, contains CRC24 of data
			var err error
			blockCrc, err = ParseHexUint32(string(parts[1]))
			if err != nil {
				return nil, fmt.Errorf("error parsing block CRC24: %s", parts[1])
			}
			continue
		}

		lineParts := bytes.Split(parts[1], []byte(" "))
		// as lineParts contains sub-arrays of encoded bytes, the length of lineParts is equal to the number of bytes in the line + 1 (for the checksum)
		// a line must never contain no data, this a line must contain at least two parts, one byte and the checksum
		// (the last line, containing only the block checksum, is already handled above)
		if len(lineParts) > BytesPerLine+1 || len(lineParts) < 2 {
			return nil, fmt.Errorf("unexpected line length: line %s: %s", lineNumber, parts[1])
		}

		// lineParts[0] - lineParts[last-1] contain the data
		bytesHex := bytes.Join(lineParts[0:len(lineParts)-1], []byte(""))
		// while the last part contains the checksum
		checksumHex := lineParts[len(lineParts)-1]

		bytesData, err := hex.DecodeString(string(bytesHex))
		if err != nil {
			return nil, err
		}

		checksumData, err := ParseHexUint32(string(checksumHex))
		if err != nil {
			return nil, fmt.Errorf("error parsing line checksum: %s", checksumHex)
		}

		lineNum := 0
		_, err = fmt.Sscanf(lineNumber, "%d", &lineNum)
		if err != nil {
			return nil, err
		}

		lineData := LineData{
			LineNumber: uint32(lineNum),
			Data:       bytesData,
			CRC24:      checksumData,
		}

		if ValidateCRC24(lineData.Data, lineData.CRC24) {
			result = append(result, lineData)
		} else {
			return nil, fmt.Errorf("invalid line checksum: line %d has checksum %06X, expected %06X", lineData.LineNumber, Crc24Checksum(lineData.Data), lineData.CRC24)
		}
	}

	// 2. Assemble data

	// 2.1. Sort lines
	for i := 0; i < len(result); i++ {
		for j := i + 1; j < len(result); j++ {
			if result[i].LineNumber > result[j].LineNumber {
				tmp := result[i]
				result[i] = result[j]
				result[j] = tmp
			}
		}
	}

	// 2.2. Ensure that lines are consecutive, starting at 1
	// as we sorted the lines, we can just check the first and last line

	if len(result) == 0 {
		return nil, errors.New("no lines found")
	}

	if result[0].LineNumber != 1 {
		return nil, fmt.Errorf("invalid first line number: %d", result[0].LineNumber)
	}

	// this also ensures that we have all lines, as the last line number must equal the number of lines
	if result[len(result)-1].LineNumber != uint32(len(result)) {
		return nil, fmt.Errorf("invalid last line number: %d", result[len(result)-1].LineNumber)
	}

	var resultData []byte
	for _, line := range result {
		resultData = append(resultData, line.Data...)
	}

	// 3. Validate data checksum
	if !ValidateCRC24(resultData, blockCrc) {
		return nil, errors.New("invalid block checksum")
	}

	return resultData, nil
}

func ParseHexUint32(hex string) (uint32, error) {
	h := strings.ToLower(hex)
	h = strings.ReplaceAll(h, "0x", "")
	h = strings.ReplaceAll(h, " ", "")

	var n uint32
	_, err := fmt.Sscanf(h, "%x", &n)
	if err != nil {
		return 0, errors.Join(errors.New("error parsing hexadecimal value"), err)
	}
	return n, nil
}

func BytesFromBase64(data string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(data)
}
