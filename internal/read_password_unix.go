//go:build !windows

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

	"github.com/manifoldco/promptui"
	"golang.org/x/term"
)

func readTtyLinePlatform() ([]byte, error) {
	// if stdin is a terminal, use it with promptui
	if term.IsTerminal(syscall.Stdin) {
		prompt := promptui.Prompt{
			Label:  "Passphrase (hidden)",
			Mask:   '*',
			Stdout: os.Stderr,
		}

		result, err := prompt.Run()
		if err != nil {
			return nil, errors.Join(errors.New("could run prompt"), err)
		}

		return []byte(result), nil
	}

	// otherwise, try /dev/tty
	tty, err := os.Open("/dev/tty")
	if err != nil {
		return nil, errors.Join(errors.New("could not open /dev/tty"), err)
	}

	password, err := term.ReadPassword(int(tty.Fd()))
	if err != nil {
		return nil, errors.Join(errors.New("could not read password from /dev/tty"),
			err)
	}
	if password == nil {
		return nil, errors.New("could not read password from /dev/tty")
	}

	if err = tty.Close(); err != nil {
		return nil, errors.Join(errors.New("could not close /dev/tty"), err)
	}

	return password, nil
}
