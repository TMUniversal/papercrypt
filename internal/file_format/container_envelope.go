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

package file_format

import (
	"errors"
	"strings"

	"github.com/tmuniversal/papercrypt/v3/internal/file_format/envelope"
)

// UnmarshalEnvelope unwraps an envelope string and deserializes the
// contained binary container into a PaperCrypt. The envelope encoding is
// determined from the envelope header, and the payload is decompressed
// transparently if the header marks it as gzipped.
func UnmarshalEnvelope(data string) (*PaperCrypt, error) {
	if !strings.HasPrefix(data, envelope.Magic) {
		return nil, errors.New("unsupported format: expected PC envelope")
	}

	hdr, _, err := envelope.ParseHeader(data)
	if err != nil {
		return nil, errors.Join(errors.New("error parsing envelope header"), err)
	}

	enc, err := envelope.NewEncoder(hdr.Encoding)
	if err != nil {
		return nil, err
	}

	content, err := envelope.Unwrap(data, enc)
	if err != nil {
		return nil, errors.Join(errors.New("error unwrapping envelope"), err)
	}

	pc, err := UnmarshalBinary(content)
	if err != nil {
		return nil, errors.Join(errors.New("error deserializing binary container"), err)
	}
	return pc, nil
}
