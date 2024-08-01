/*
 * This file is part of PaperCrypt.
 *
 * PaperCrypt lets you prepare encrypted messages for printing on paper.
 * Copyright (C) 2024 TMUniversal <me@tmuniversal.eu>.
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
	"strings"

	"github.com/caarlos0/log"
)

type PaperCryptDataFormat uint8

const (
	PaperCryptDataFormatPGP PaperCryptDataFormat = 0
	PaperCryptDataFormatRaw PaperCryptDataFormat = 1
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
		return PaperCryptDataFormat(0xFF)
	}
}

type PaperCryptContainerVersion uint32

const (
	PaperCryptContainerVersionUnknown PaperCryptContainerVersion = 0
	PaperCryptContainerVersionMajor1  PaperCryptContainerVersion = 1
	PaperCryptContainerVersionMajor2  PaperCryptContainerVersion = 2
	PaperCryptContainerVersionDevel   PaperCryptContainerVersion = PaperCryptContainerVersion(0xFFFFFFFF)
)

func (v PaperCryptContainerVersion) String() string {
	switch v {
	case PaperCryptContainerVersionMajor1:
		return "1"
	case PaperCryptContainerVersionMajor2:
		return "2"
	case PaperCryptContainerVersionDevel:
		return "devel"
	default:
		return "unknown"
	}
}

func PaperCryptContainerVersionFromString(s string) PaperCryptContainerVersion {
	major := strings.Split(s, ".")[0]
	log.Debugf("PaperCrypt Version: %s", major)

	switch major {
	case "1":
		return PaperCryptContainerVersionMajor1
	case "2":
		return PaperCryptContainerVersionMajor2
	case "devel":
		return PaperCryptContainerVersionDevel
	default:
		return PaperCryptContainerVersionUnknown
	}
}
