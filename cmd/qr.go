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
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/tmuniversal/papercrypt/internal"
)

// qrCmd represents the data command
var qrCmd = &cobra.Command{
	Aliases: []string{"q"},
	Args:    cobra.MaximumNArgs(1),
	Use:     "qr <input>",
	Short:   "Decode a document from a QR code.",
	Long: `Decode a document from a QR code.

This command allows you to decode data saved by PaperCrypt.
The QR code in a PaperCrypt document contains a JSON serialized object
that contains the encrypted data and the PaperCrypt metadata.`,
	Example: `papercrypt qr ./qr.png | papercrypt decode -o ./out.json -P passphrase`,
	Run: func(cmd *cobra.Command, args []string) {
		// 1. get data from either argument or inFileName
		if len(args) != 0 {
			inFileName = args[0]
		}

		inFile := internal.PrintInputAndGetReader(cmd, inFileName)
		defer inFile.Close()

		img, _, err := image.Decode(inFile)
		if err != nil {
			cmd.Println("Error decoding image:", err)
			os.Exit(1)
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
		outFile := internal.GetFileHandleCarefully(cmd, outFileName, overrideOutFile)
		defer outFile.Close()

		data := result.GetText()

		// 3. Deserialize
		pc := internal.PaperCrypt{}
		err = json.Unmarshal([]byte(data), &pc)
		if err != nil {
			internal.Fatal(cmd, errors.Wrap(err, "error deserializing data"))
		}

		// 6. Write to file
		output, err := pc.GetText(false)
		if err != nil {
			internal.Fatal(cmd, errors.Wrap(err, "error deserializing data"))
		}
		n, err := outFile.Write(output)
		if err != nil {
			internal.Fatal(cmd, errors.Wrap(err, "error writing output"))
		}

		internal.PrintWrittenSize(cmd, n, outFile)
	},
}

func init() {
	rootCmd.AddCommand(qrCmd)
}
