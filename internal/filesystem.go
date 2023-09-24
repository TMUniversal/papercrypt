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
	"io"
	"os"

	"github.com/caarlos0/log"
	"github.com/pkg/errors"
)

// GetFileHandleCarefully returns a file handle for the given path.
// will warn if the file already exists, and error if override is false.
// if path is empty, returns os.Stdout.
func GetFileHandleCarefully(path string, override bool) (*os.File, error) {
	if path == "" || path == "-" {
		return os.Stdout, nil
	}

	if _, err := os.Stat(path); err == nil {
		if !override {
			return nil, errors.Errorf("file %s already exists, use --force to override", path)
		} else {
			log.WithField("path", path).Warn("Overriding existing file!")
		}
	}

	out, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return nil, errors.Errorf("error opening file '%s': %s", path, err)
	}

	return out, nil
}

// PrintInputAndGetReader prints the input source and returns the reader.
// if path is empty, returns os.Stdin.
// must be closed by the caller.
func PrintInputAndGetReader(inFileName string) (*os.File, error) {
	var err error
	var inFile *os.File
	if inFileName == "" || inFileName == "-" {
		inFile = os.Stdin
	} else {
		inFile, err = os.Open(inFileName)
		if err != nil {
			return nil, errors.Wrap(err, "error opening file")
		}
	}

	log.WithField("input", inFileName).Debug("Reading from input")

	return inFile, nil
}

// PrintInputAndRead prints the input source and returns the contents of the file.
// if path is empty, returns os.Stdin.
func PrintInputAndRead(inFileName string) ([]byte, error) {
	inFile, err := PrintInputAndGetReader(inFileName)
	if err != nil {
		return nil, err
	}

	contents, err := io.ReadAll(inFile)
	if err != nil && err != io.EOF {
		return nil, errors.Wrap(err, "error reading file")
	}

	if err := inFile.Close(); err != nil {
		return nil, errors.Wrap(err, "error closing file")
	}

	return contents, nil
}
