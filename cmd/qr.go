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

package cmd

import (
	"encoding/json"
	"errors"
	"image"
	"os"

	"github.com/caarlos0/log"
	"github.com/makiuchi-d/gozxing"
	"github.com/makiuchi-d/gozxing/qrcode"
	"github.com/spf13/cobra"
	"github.com/tmuniversal/papercrypt/internal"
)

// qrCmd represents the data command
var qrCmd = &cobra.Command{
	Aliases:      []string{"q"},
	Args:         cobra.MaximumNArgs(1),
	SilenceUsage: true,
	Use:          "qr <input>",
	Short:        "Decode a document from a QR code.",
	Long: `Decode a document from a QR code.

This command allows you to decode data saved by PaperCrypt.
The QR code in a PaperCrypt document contains a JSON serialized object
that contains the encrypted data and the PaperCrypt metadata.`,
	Example: `papercrypt qr ./qr.png | papercrypt decode -o ./out.json -P passphrase`,
	RunE: func(_ *cobra.Command, args []string) error {
		// 1. get data from either argument or inFileName
		if len(args) != 0 {
			inFileName = args[0]
		}

		inFile, err := internal.PrintInputAndGetReader(inFileName)
		if err != nil {
			return err
		}

		img, _, err := image.Decode(inFile)
		if err != nil {
			return errors.Join(errors.New("error decoding image"), err)
		}

		if err := inFile.Close(); err != nil {
			return errors.Join(errors.New("error closing input file"), err)
		}

		bmp, err := gozxing.NewBinaryBitmapFromImage(img)
		if err != nil {
			return errors.Join(errors.New("error creating binary bitmap"), err)
		}

		qrReader := qrcode.NewQRCodeReader()
		result, err := qrReader.Decode(bmp, nil)
		if err != nil {
			return errors.Join(errors.New("error decoding QR code"), err)
		}

		// 2. Open output file
		outFile, err := internal.GetFileHandleCarefully(outFileName, overrideOutFile)
		if err != nil {
			return err
		}
		defer func(file *os.File) {
			err := internal.CloseFileIfNotStd(file)
			if err != nil {
				log.WithError(err).Error("Error closing file")
			}
		}(outFile)

		data := result.GetText()

		// 3. Deserialize
		pc := internal.PaperCrypt{}
		err = json.Unmarshal([]byte(data), &pc)
		if err != nil {
			return errors.Join(errors.New("error deserializing data"), err)
		}

		// 6. Write to file
		output, err := pc.GetText(false)
		if err != nil {
			return errors.Join(errors.New("error deserializing data"), err)
		}
		n, err := outFile.Write(output)
		if err != nil {
			return errors.Join(errors.New("error writing output"), err)
		}

		internal.PrintWrittenSize(n, outFile)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(qrCmd)
}
