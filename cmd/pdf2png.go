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

package cmd

import (
	"bytes"
	"image/png"
	"os"

	"github.com/karmdip-mi/go-fitz"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/tmuniversal/papercrypt/internal"
)

// pdf2pngCmd represents the pdf2png command
var pdf2pngCmd = &cobra.Command{
	Use:   "pdf2png",
	Short: "pdf2png takes the first page of a pdf and outputs it as a png",
	Args:  cobra.MaximumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		var inFile *os.File
		if len(args) != 0 {
			inFileName = args[0]
		}
		if inFileName == "" || inFileName == "-" {
			cmd.Println("Reading from stdin")
			inFile = os.Stdin
		} else {
			cmd.Printf("Reading from %s\n", inFileName)
			var err error
			inFile, err = os.Open(inFileName)
			if err != nil {
				cmd.Println("Error opening input file:", err)
				os.Exit(1)
			}
			defer inFile.Close()
		}

		if len(args) > 1 {
			outFileName = args[1]
		}
		outFile, err := internal.GetFileHandleCarefully(outFileName, overrideOutFile)
		if err != nil {
			cmd.Println("Error opening output file:", err)
			os.Exit(1)
		}
		defer outFile.Close()

		n, err := pdf2png(inFile, outFile)
		if err != nil {
			cmd.Println("Error converting pdf to png:", err)
			os.Exit(1)
		}

		cmd.Printf("Successfully wrote %s bytes to %s\n", internal.SprintBinarySize(n), outFile.Name())
	},
}

func pdf2png(inFile *os.File, outFile *os.File) (int, error) {
	doc, err := fitz.NewFromReader(inFile)
	if err != nil {
		return 0, errors.Wrap(err, "error opening pdf file")
	}
	defer doc.Close()

	// Convert the first page to a png
	img, err := doc.Image(0)
	if err != nil {
		return 0, errors.Wrap(err, "error converting pdf to png")
	}

	// Write the png to the output file
	tmp := new(bytes.Buffer)
	err = png.Encode(tmp, img)
	if err != nil {
		return 0, errors.Wrap(err, "error encoding png")
	}

	return outFile.Write(tmp.Bytes())
}

func init() {
	rootCmd.AddCommand(pdf2pngCmd)
}
