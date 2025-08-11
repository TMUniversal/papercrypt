/*
 * This file is part of PaperCrypt.
 *
 * PaperCrypt lets you prepare encrypted messages for printing on paper.
 * Copyright (C) 2023-2024 TMUniversal <me@tmuniversal.eu>.
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
	"fmt"
	"os"
)

// SensitivePrompt reads a password from the tty (if available) or stdin (if not).
func SensitivePrompt() ([]byte, error) {
	_, _ = fmt.Fprint(os.Stderr, "Passphrase: ")

	p, e := readTtyLine()

	_, _ = fmt.Fprint(os.Stderr, "\n")

	return p, e
}

func readTtyLine() ([]byte, error) {
	return readTtyLinePlatform()
}
