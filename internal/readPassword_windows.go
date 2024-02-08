//go:build windows

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
	"errors"
	"os"
	"syscall"

	"golang.org/x/term"

	"github.com/manifoldco/promptui"
)

func readTtyLine() ([]byte, error) {
	// if stdin is a terminal, use it with promptui
	if term.IsTerminal(int(syscall.Stdin)) {
		prompt := promptui.Prompt{
			Label:  "Passphrase",
			Mask:   '*',
			Stdout: os.Stderr,
		}

		result, err := prompt.Run()
		if err != nil {
			return nil, errors.Join(errors.New("could run prompt"), err)
		}

		return []byte(result), nil
	}

	return nil, errors.New("cannot access terminal outside stdin on Windows, if you must pass data through stdin, you can use the --passphrase flag")
}
