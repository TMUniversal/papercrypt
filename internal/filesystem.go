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
	"log"
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var l = log.New(os.Stderr, "", 0)

// GetFileHandleCarefully returns a file handle for the given path.
// will warn if the file already exists, and crash if override is false.
// if path is empty, returns os.Stdout.
// will exit with code 1 if the file cannot be opened.
// must be closed by the caller.
func GetFileHandleCarefully(cmd *cobra.Command, path string, override bool) *os.File {
	if path == "" || path == "-" {
		return os.Stdout
	}

	if _, err := os.Stat(path); err == nil {
		if !override {
			Fatal(cmd, errors.Errorf("file %s already exists, use --force to override", path))
		} else {
			l.Printf("Overriding existing file \"%s\"!\n", path)
		}
	}

	out, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		Fatal(cmd, errors.Errorf("error opening file '%s': %s", path, err))
	}

	return out
}

// PrintInputAndGetReader prints the input source and returns the reader.
// if path is empty, returns os.Stdin.
// will exit with code 1 if the file cannot be opened.
// must be closed by the caller.
func PrintInputAndGetReader(cmd *cobra.Command, inFileName string) *os.File {
	PrintInputSource(cmd, inFileName)
	var err error
	var inFile *os.File
	if inFileName == "" || inFileName == "-" {
		inFile = os.Stdin
	} else {
		inFile, err = os.Open(inFileName)
		if err != nil {
			Fatal(cmd, err)
		}
	}
	return inFile
}

// PrintInputAndRead prints the input source and returns the contents of the file.
// if path is empty, returns os.Stdin.
// will exit with code 1 if the file cannot be opened or read.
func PrintInputAndRead(cmd *cobra.Command, inFileName string) []byte {
	inFile := PrintInputAndGetReader(cmd, inFileName)
	defer inFile.Close()
	contents, err := io.ReadAll(inFile)
	if err != nil && err != io.EOF {
		Fatal(cmd, errors.Wrap(err, "error reading file"))
	}
	return contents
}
