/*
 * This file is part of PaperCrypt.
 *
 * PaperCrypt lets you prepare encrypted messages for printing on paper.
 * Copyright (C) 2024-2026 TMUniversal <me@tmuniversal.eu>.
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
	"strings"

	"github.com/caarlos0/log"
)

type PaperCryptDataFormat uint8

const (
	PaperCryptDataFormatPGP     PaperCryptDataFormat = 0
	PaperCryptDataFormatRaw     PaperCryptDataFormat = 1
	PaperCryptDataFormatUnknown PaperCryptDataFormat = 0xFF
)

func (f PaperCryptDataFormat) String() string {
	switch f {
	case PaperCryptDataFormatPGP:
		return "PGP"
	case PaperCryptDataFormatRaw:
		return "Raw"
	default:
		return "Unknown"
	}
}

func PaperCryptDataFormatFromString(s string) PaperCryptDataFormat {
	switch s {
	case "PGP":
		return PaperCryptDataFormatPGP
	case "Raw":
		return PaperCryptDataFormatRaw
	default:
		return PaperCryptDataFormatUnknown
	}
}

type PaperCryptContainerVersion uint32

const (
	PaperCryptContainerVersionUnknown PaperCryptContainerVersion = 0
	// PaperCryptContainerVersionMajor1 container format from PaperCryptV1, used for backwards compatibility
	PaperCryptContainerVersionMajor1 PaperCryptContainerVersion = 1
	PaperCryptContainerVersionMajor2 PaperCryptContainerVersion = 2
	PaperCryptContainerVersionMajor3 PaperCryptContainerVersion = 3
	PaperCryptContainerVersionDevel  PaperCryptContainerVersion = PaperCryptContainerVersion(
		0xFFFFFFFF,
	)
)

func (v PaperCryptContainerVersion) String() string {
	switch v {
	case PaperCryptContainerVersionMajor1:
		return "1"
	case PaperCryptContainerVersionMajor2:
		return "2"
	case PaperCryptContainerVersionMajor3:
		return "3"
	case PaperCryptContainerVersionDevel:
		return "devel"
	default:
		return "unknown"
	}
}

func PaperCryptContainerVersionFromString(s string) PaperCryptContainerVersion {
	major := strings.TrimPrefix(s, "v")
	major = strings.Split(major, ".")[0]
	log.Debugf("PaperCrypt Version: %s", major)

	switch major {
	case "0":
		return PaperCryptContainerVersionDevel
	case "1":
		return PaperCryptContainerVersionMajor1
	case "2":
		return PaperCryptContainerVersionMajor2
	case "3":
		return PaperCryptContainerVersionMajor3
	case "devel":
		return PaperCryptContainerVersionDevel
	default:
		return PaperCryptContainerVersionUnknown
	}
}
