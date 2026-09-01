/*
 * This file is part of PaperCrypt.
 *
 * PaperCrypt lets you prepare encrypted messages for printing on paper.
 * Copyright (C) 2026 TMUniversal <me@tmuniversal.eu>.
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

package envelope

import (
	"fmt"

	"github.com/dasio/base45"
)

type EncodingType uint8

const (
	// EncodingTypeRaw is reserved for a future raw encoder; no encoder
	// currently registers it, and it must remain value 0 to keep the wire
	// header stable.
	EncodingTypeRaw EncodingType = iota
	EncodingTypeBase45
)

// ContentEncoder implementations must produce deterministic output for a
// given input.
type ContentEncoder interface {
	EncodeToString(data []byte) string
	DecodeString(data string) ([]byte, error)
	EncodedCRCSize() int
	EncodingType() EncodingType
}

type Base45Encoder struct{}

func (Base45Encoder) EncodeToString(data []byte) string {
	return base45.EncodeToString(data)
}

func (Base45Encoder) DecodeString(data string) ([]byte, error) {
	return base45.DecodeString(data)
}

// EncodedCRCSize is 6: base45 packs 4 bytes as 6 characters.
func (Base45Encoder) EncodedCRCSize() int {
	return 6
}

func (Base45Encoder) EncodingType() EncodingType {
	return EncodingTypeBase45
}

func NewEncoder(t EncodingType) (ContentEncoder, error) {
	switch t {
	case EncodingTypeBase45:
		return Base45Encoder{}, nil
	default:
		return nil, fmt.Errorf("unsupported envelope encoding type %d", t)
	}
}
