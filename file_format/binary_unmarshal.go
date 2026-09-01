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
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/tmuniversal/papercrypt/v3/file_format/envelope"
)

func UnmarshalBinary(data []byte) (*PaperCrypt, error) {
	if len(data) < BinaryHeaderSize {
		return nil, ErrBinaryTruncated
	}

	if [2]byte(data[0:2]) != BinaryMagic {
		return nil, ErrBinaryInvalidMagic
	}

	if data[2] != CurrentBinaryFormatVersion {
		return nil, fmt.Errorf("%w: %d", ErrBinaryUnsupportedVersion, data[2])
	}

	r := data[BinaryHeaderSize:]
	p := &PaperCrypt{}

	if len(r) < 3 {
		return nil, ErrBinaryTruncated
	}
	p.Version = formatVersion(r[0], r[1], r[2])
	r = r[3:]

	if len(r) < 1 {
		return nil, ErrBinaryTruncated
	}
	p.DataFormat = PaperCryptDataFormat(r[0])
	r = r[1:]

	if len(r) < 1 {
		return nil, ErrBinaryTruncated
	}
	serialLen := int(r[0])
	r = r[1:]
	if len(r) < serialLen {
		return nil, ErrBinaryTruncated
	}
	p.SerialNumber = string(r[:serialLen])
	r = r[serialLen:]

	if len(r) < 1 {
		return nil, ErrBinaryTruncated
	}
	purposeLen := int(r[0])
	r = r[1:]
	if len(r) < purposeLen {
		return nil, ErrBinaryTruncated
	}
	p.Purpose = string(r[:purposeLen])
	r = r[purposeLen:]

	if len(r) < 1 {
		return nil, ErrBinaryTruncated
	}
	commentLen := int(r[0])
	r = r[1:]
	if len(r) < commentLen {
		return nil, ErrBinaryTruncated
	}
	p.Comment = string(r[:commentLen])
	r = r[commentLen:]

	if len(r) < 8 {
		return nil, ErrBinaryTruncated
	}
	tsVal := binary.BigEndian.Uint64(r[:8])
	p.CreatedAt = time.Unix(0, int64(tsVal)) //nolint:gosec // Unix timestamps are non-negative
	r = r[8:]

	if len(r) < 32 {
		return nil, ErrBinaryTruncated
	}
	copy(p.DataSHA256[:], r[:32])
	r = r[32:]

	p.Data = r

	return p, nil
}

func UnmarshalBinaryFromReader(r io.Reader) (*PaperCrypt, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("binary: read: %w", err)
	}
	return UnmarshalBinary(data)
}

func UnmarshalEnvelope(data string, opts ...envelope.CompressorOption) (*PaperCrypt, error) {
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

	content, err := envelope.Unwrap(data, enc, opts...)
	if err != nil {
		return nil, errors.Join(errors.New("error unwrapping envelope"), err)
	}

	pc, err := UnmarshalBinary(content)
	if err != nil {
		return nil, errors.Join(errors.New("error deserializing binary container"), err)
	}
	return pc, nil
}
