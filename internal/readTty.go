//go:build !windows

/*
 * This file is part of PaperCrypt.
 *
 * PaperCrypt lets you prepare encrypted messages for printing on paper.
 * Copyright (C) 2023 TMUniversal <me@tmuniversal.eu>.
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
	"bufio"
	"os"
	"strings"

	"github.com/pkg/errors"
)

func ReadTtyLine() (string, error) {
	tty, err := os.Open("/dev/tty")
	if err != nil {
		return "", errors.Wrap(err, "could not open /dev/tty")
	}
	defer tty.Close()

	reader := bufio.NewReader(tty)
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", errors.Wrap(err, "could not read from /dev/tty")
	}

	input = strings.ReplaceAll(input, "\r", "")
	input = strings.ReplaceAll(input, "\n", "")

	return input, nil
}
