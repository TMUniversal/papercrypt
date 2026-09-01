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
	"errors"
	"fmt"
	"strconv"
	"strings"
)

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

// splitHeaderBody splits text at the two empty lines separating the
// header section from the serialized body, returning text unchanged as the
// header when the separator is absent.
func splitHeaderBody(text []byte) (header, body []byte) {
	parts := bytes.SplitN(text, []byte("\n\n\n"), 2)
	if len(parts) != 2 {
		return text, nil
	}
	return parts[0], parts[1]
}

func SplitTextHeaderAndBody(data []byte) ([]byte, []byte, error) {
	header, body := splitHeaderBody(data)
	if body == nil {
		return nil, nil, errors.New(
			"header not discernible, header and content should be separated by two empty lines",
		)
	}
	return header, body, nil
}

func ParseHexUint32(hex string) (uint32, error) {
	s := strings.TrimPrefix(strings.ToLower(hex), "0x")
	s = strings.ReplaceAll(s, " ", "")

	n, err := strconv.ParseUint(s, 16, 32)
	if err != nil {
		return 0, errors.Join(errors.New("error parsing hexadecimal value"), err)
	}
	return uint32(n), nil
}
