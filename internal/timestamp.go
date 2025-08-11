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

const (
	// TimeStampFormatLong shows the full date and time precisely for humans. It is used for the container file, as well as the timestamp command-line parameter.
	TimeStampFormatLong = "Mon, 02 Jan 2006 15:04:05.000000000 -0700"
	// TimeStampFormatLongTZ is used for v1 backwards compatibility.
	TimeStampFormatLongTZ = "Mon, 02 Jan 2006 15:04:05.000000000 MST"
	// TimeStampFormatShort is used in parsing the timestamp command-line parameter [cmd/generateCmd].
	TimeStampFormatShort = "2006-01-02 15:04:05"
	// TimeStampFormatDate is used as an alternative format for the command-line parameter.
	TimeStampFormatDate = "2006-01-02"
	// TimeStampFormatPDFHeader is used exclusively for the timestamps in displayed PDF headers. Both for the normal papercrypt sheet and the phrase sheet.
	TimeStampFormatPDFHeader = "2006-01-02 15:04 -0700"
)
