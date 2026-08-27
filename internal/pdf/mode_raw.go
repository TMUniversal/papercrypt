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
	// pdfSectionRepresentationContentRaw is the suffix appended for raw, unencrypted data.
	pdfSectionRepresentationContentRaw = " The data is stored as-is, unencrypted and uncompressed, so it can be read directly from the hex digits."
	// pdfSectionRecoveryContentRaw is the recovery instruction for raw data with a QR code.
	pdfSectionRecoveryContentRaw = "Scan the QR code, or copy the data into a computer by typing it in or using OCR. The data is stored as-is, so reassembling the bytes reproduces the original file."
	// pdfSectionRecoveryContentRawNoQR is the recovery instruction for raw data without a QR code.
	pdfSectionRecoveryContentRawNoQR = "No QR code is printed on this sheet. Copy the data into a computer by typing it in or using OCR; the bytes reproduce the original file as-is, with no decryption needed."
)

// rawQRLines returns the text lines for the raw recovery sheet with a QR code.
func rawQRLines() lineSet {
	lines := defaultLines()
	lines.representationBase = pdfSectionRepresentationContentBaseQR
	lines.representationSuffix = pdfSectionRepresentationContentRaw
	lines.recoveryContent = pdfSectionRecoveryContentRaw
	return lines
}

// rawNoQRLines returns the text lines for the raw recovery sheet without a QR code.
func rawNoQRLines() lineSet {
	lines := defaultLines()
	lines.representationBase = pdfSectionRepresentationContentBaseNoQR
	lines.representationSuffix = pdfSectionRepresentationContentRaw
	lines.recoveryContent = pdfSectionRecoveryContentRawNoQR
	return lines
}
