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
	"errors"

	"github.com/tmuniversal/papercrypt/v3/codematrix"
	"github.com/tmuniversal/papercrypt/v3/file_format/envelope"
)

func GenerateQR(p *PaperCrypt) ([]byte, error) {
	qrBin, err := MarshalBinary(p)
	if err != nil {
		return nil, errors.Join(errors.New("error marshalling PaperCrypt to binary"), err)
	}

	qrData := envelope.Wrap(qrBin, envelope.Base45Encoder{})

	qrImage, err := codematrix.EncodePNG(qrData)
	if err != nil {
		return nil, errors.Join(errors.New("error generating QR code"), err)
	}
	return qrImage, nil
}
