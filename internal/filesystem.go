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
	"log"
	"os"

	"github.com/pkg/errors"
)

var l = log.New(os.Stderr, "", 0)

func GetFileHandleCarefully(path string, override bool) (*os.File, error) {
	if path == "" || path == "-" {
		return os.Stdout, nil
	}

	if _, err := os.Stat(path); err == nil {
		if !override {
			return nil, errors.Errorf("file %s already exists, use --force to override", path)
		} else {
			l.Printf("Overriding existing file \"%s\"!\n", path)
		}
	}

	out, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return nil, errors.Errorf("error opening file '%s': %s", path, err)
	}

	return out, nil
}
