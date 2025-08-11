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

// PaperCryptDataFormat is an enum (uint8) of supported container formats
type PaperCryptDataFormat uint8

const (
	// PaperCryptDataFormatPGP marks that a container holds data enclosed in a PGP container
	PaperCryptDataFormatPGP PaperCryptDataFormat = 0
	// PaperCryptDataFormatRaw represents that the data encoded in the container is raw, i.e. has not been encrypted by papercrypt
	PaperCryptDataFormatRaw PaperCryptDataFormat = 1
)

// String serializes the enum value to a string deserializable by PaperCryptDataFormatFromString
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

// PaperCryptDataFormatFromString parses a container data format as a string, returning the corresponding enum value
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

// PaperCryptContainerVersion is an enum (uint32) of versions of the container format
type PaperCryptContainerVersion uint32

const (
	// PaperCryptContainerVersionUnknown represents any unknown version, which may be newer, or come from parsing invalid input
	PaperCryptContainerVersionUnknown PaperCryptContainerVersion = 0
	// PaperCryptContainerVersionMajor1 container format from PaperCryptV1, used for backwards compatibility
	PaperCryptContainerVersionMajor1 PaperCryptContainerVersion = 1
	// PaperCryptContainerVersionMajor2 container format for PaperCrypt
	PaperCryptContainerVersionMajor2 PaperCryptContainerVersion = 2
	// PaperCryptContainerVersionDevel is used instead of a set version number for development builds
	PaperCryptContainerVersionDevel PaperCryptContainerVersion = PaperCryptContainerVersion(
		0xFFFFFFFF,
	)
)

// String serializes the PaperCryptContainerVersion to a string of either a number corresponding to the major version, "devel" for a development build, or "unknown"
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

// PaperCryptContainerVersionFromString parses a version string to discover the major version of this software
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
