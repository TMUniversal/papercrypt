package internal

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/pkg/errors"
)

type LineData struct {
	LineNumber uint32
	Data       []byte
	CRC24      uint32
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
			return nil, errors.Errorf("invalid line format: %s", line)
		}

		lineNumber := strings.ReplaceAll(string(parts[0]), " ", "")
		lineNumber = strings.ReplaceAll(lineNumber, "\t", "")

		if lineNumber == fmt.Sprint(len(lines)) {
			// last line, contains CRC24 of data
			var err error
			blockCrc, err = ParseHexUint32(string(parts[1]))
			if err != nil {
				return nil, errors.Errorf("error parsing block CRC24: %s", parts[1])
			}
			continue
		}

		lineParts := bytes.Split(parts[1], []byte(" "))
		// as lineParts contains sub-arrays of encoded bytes, the length of lineParts is equal to the number of bytes in the line + 1 (for the checksum)
		// a line must never contain no data, this a line must contain at least two parts, one byte and the checksum
		// (the last line, containing only the block checksum, is already handled above)
		if len(lineParts) > BytesPerLine+1 || len(lineParts) < 2 {
			return nil, errors.Errorf("unexpected line length: line %s: %s", lineNumber, parts[1])
		}

		// lineParts[0] - lineParts[last-1] contain the data
		bytesHex := bytes.Join(lineParts[0:len(lineParts)-1], []byte(""))
		// while the last part contains the checksum
		checksumHex := lineParts[len(lineParts)-1]

		//// debug
		//fmt.Printf("line: %s\n", lineParts)
		//fmt.Printf("bytesHex: %s\n", bytesHex)
		//fmt.Printf("checksumHex: %s\n", checksumHex)

		bytesData, err := hex.DecodeString(string(bytesHex))
		if err != nil {
			return nil, err
		}

		checksumData, err := ParseHexUint32(string(checksumHex))
		if err != nil {
			return nil, errors.Errorf("error parsing line checksum: %s", checksumHex)
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
			return nil, errors.Errorf("invalid line checksum: line %d has checksum %06X, expected %06X.", lineData.LineNumber, Crc24Checksum(lineData.Data), lineData.CRC24)
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
		return nil, errors.Errorf("no lines found")
	}

	if result[0].LineNumber != 1 {
		return nil, errors.Errorf("invalid first line number: %d", result[0].LineNumber)
	}

	// this also ensures that we have all lines, as the last line number must equal the number of lines
	if result[len(result)-1].LineNumber != uint32(len(result)) {
		return nil, errors.Errorf("invalid last line number: %d", result[len(result)-1].LineNumber)
	}

	var resultData []byte
	for _, line := range result {
		resultData = append(resultData, line.Data...)
	}

	// 3. Validate data checksum
	if !ValidateCRC24(resultData, blockCrc) {
		return nil, errors.Errorf("invalid block checksum")
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
		return 0, errors.Errorf("error parsing hexadecimal value: %s", err)
	}
	return n, nil
}

func BytesFromBase64(data string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(data)
}
