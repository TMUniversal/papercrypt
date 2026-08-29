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

package pdf

const (
	pdfSectionRepresentationContentPGP = " The data is gzipped and encrypted."
	pdfSectionRecoveryContentPGP       = "Scan the QR code, or copy the data into a computer by typing it in or using OCR. Then decrypt it with the encryption passphrase."
	pdfSectionRecoveryContentPGPNoQR   = "No QR code is printed on this sheet. Copy the data into a computer by typing it in or using OCR, then decrypt it with the encryption passphrase."
)

func pgpQRLines() lineSet {
	lines := defaultLines()
	lines.representationBase = pdfSectionRepresentationContentBaseQR
	lines.representationSuffix = pdfSectionRepresentationContentPGP
	lines.recoveryContent = pdfSectionRecoveryContentPGP
	return lines
}

func pgpNoQRLines() lineSet {
	lines := defaultLines()
	lines.representationBase = pdfSectionRepresentationContentBaseNoQR
	lines.representationSuffix = pdfSectionRepresentationContentPGP
	lines.recoveryContent = pdfSectionRecoveryContentPGPNoQR
	return lines
}
