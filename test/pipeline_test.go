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

package test

import (
	"bytes"
	"image/png"
	"os"
	"os/exec"
	"path"
	"strings"
	"testing"
)

func TestFullPipeline(t *testing.T) {
	message := "{\"message\":\"Hello World!\"}"
	passphrase := "test"

	t.Run("generate > extract qr code > extract message > decode", func(t *testing.T) {
		tmpDir := t.TempDir()

		// generate
		pdfPath := tmpDir + string(os.PathSeparator) + "t.pdf"

		genCmd := exec.Command("go", "run", "../main.go", "generate", "--purpose", "Test", "--comment", "Test", "--date", "2023-09-20 12:00:00", "--passphrase", passphrase, "-o", pdfPath)
		genCmd.Stdin = bytes.NewBufferString(message)
		_, err := genCmd.Output()
		if err != nil {
			t.Logf("exit error: %s", err.(*exec.ExitError).Stderr)
			t.Fatal(err)
		}

		// extract qr code
		extractCmd := exec.Command("pdfcpu", "extract", "-m", "image", pdfPath, tmpDir)
		_, err = extractCmd.Output()
		if err != nil {
			t.Logf("exit error: %s", err.(*exec.ExitError).Stderr)
			t.Fatal(err)
		}

		// find qr code file (there should be two files beginning in output_1_, one being the s/n data matrix, the other being the qr code)
		var qrCodeFile string

		files, err := os.ReadDir(tmpDir)
		if err != nil {
			t.Fatal(err)
		}

		for _, file := range files {
			if file.IsDir() {
				continue
			}

			t.Logf("found file %s", file.Name())

			if strings.HasSuffix(file.Name(), ".png") && strings.HasPrefix(file.Name(), "t_1_") {
				// create a reader
				fileName := path.Join(tmpDir, file.Name())
				reader, err := os.Open(fileName)
				if err != nil {
					t.Fatal(err)
				}

				// check the png dimensions
				decode, err := png.Decode(reader)
				if err != nil {
					t.Fatal(err)
				}

				err = reader.Close()
				if err != nil {
					t.Fatal(err)
				}

				if decode.Bounds().Dx() <= 256 || decode.Bounds().Dy() <= 256 {
					// this is the data matrix
					t.Logf("%s has smaller dimensions than expected for qr code, skipping...", fileName)
					continue
				}

				t.Logf("chose %s as qr code file", fileName)
				qrCodeFile = fileName
				break
			}
		}

		if qrCodeFile == "" {
			t.Fatal("could not find qr code file")
		}

		// extract message
		qrCodeFileReader, err := os.Open(qrCodeFile)
		if err != nil {
			t.Fatal(err)
		}
		defer qrCodeFileReader.Close()

		qrCmd := exec.Command("go", "run", "../main.go", "qr")
		qrCmd.Stdin = qrCodeFileReader
		var out bytes.Buffer
		qrCmd.Stdout = &out
		err = qrCmd.Run()
		if err != nil {

			t.Fatal(err)
		}

		// decode
		decodeCmd := exec.Command("go", "run", "../main.go", "decode", "--passphrase", passphrase)
		decodeCmd.Stdin = bytes.NewBuffer(out.Bytes())
		out.Truncate(0)
		decodeCmd.Stdout = &out
		err = decodeCmd.Run()
		if err != nil {
			t.Fatal(err)
		}

		if out.String() != message {
			t.Fatalf("expected %s, got %s", message, out.String())
		}
	})
}
