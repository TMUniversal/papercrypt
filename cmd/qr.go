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
	"encoding/json"
	"image"
	"os"

	"github.com/makiuchi-d/gozxing"
	"github.com/makiuchi-d/gozxing/qrcode"
	"github.com/spf13/cobra"
	"github.com/tmuniversal/papercrypt/internal"
)

// qrCmd represents the data command
var qrCmd = &cobra.Command{
	Use:   "qr <path>",
	Short: "Decode a document from a QR code.",
	Long: `Decode a document from a QR code.

This command allows you to decode data saved by PaperCrypt.
The QR code in a PaperCrypt document contains a JSON serialized object
that contains the encrypted data and the PaperCrypt metadata.`,
	Example: `papercrypt qr ./qr.png | papercrypt decode -o ./out.json -P passphrase`,
	Args:    cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// 1. get data from either argument or inFileName
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
		}

		img, _, err := image.Decode(inFile)
		if err != nil {
			cmd.Println("Error decoding image:", err)
			os.Exit(1)
		}

		if inFileName != "" && inFileName != "-" {
			err = inFile.Close()
			if err != nil {
				cmd.Println("Error closing input file:", err)
				os.Exit(1)
			}
		}

		bmp, err := gozxing.NewBinaryBitmapFromImage(img)
		if err != nil {
			cmd.Println("Error creating binary bitmap:", err)
			os.Exit(1)
		}

		qrReader := qrcode.NewQRCodeReader()
		result, err := qrReader.Decode(bmp, nil)
		if err != nil {
			cmd.Println("Error decoding QR code:", err)
			os.Exit(1)
		}

		// 2. Open output file
		outFile, err := internal.GetFileHandleCarefully(outFileName, overrideOutFile)
		if err != nil {
			cmd.Println("Error opening output file:", err)
			os.Exit(1)
		}
		defer outFile.Close()

		data := result.GetText()

		// 3. Deserialize
		pc := internal.PaperCrypt{}
		err = json.Unmarshal([]byte(data), &pc)
		if err != nil {
			cmd.Println("Error deserializing data:", err)
			os.Exit(1)
		}

		// 6. Write to file
		// binData := pc.Data.GetBinary()
		// output := internal.SerializeBinary(&binData)
		output, err := pc.GetText(false)
		if err != nil {
			cmd.Println("Error serializing data:", err)
			os.Exit(1)
		}
		n, err := outFile.Write(output)
		if err != nil {
			cmd.Println("Error writing to file:", err)
			os.Exit(1)
		}

		cmd.Printf("Wrote %d bytes to %s\n", n, outFile.Name())
	},
}

func init() {
	rootCmd.AddCommand(qrCmd)
}
