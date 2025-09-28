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
	"io"
	"os"

	"github.com/caarlos0/log"
	"github.com/makiuchi-d/gozxing"
	"github.com/makiuchi-d/gozxing/aztec"
	"github.com/makiuchi-d/gozxing/qrcode"
	"github.com/spf13/cobra"
	pcv1 "github.com/tmuniversal/papercrypt/internal"
	"github.com/tmuniversal/papercrypt/v2/internal"
)

var (
	qrCmdFromJSON = false
	qrCmdToJSON   = false
)

type versionContainerV1 struct {
	// Version should contain the semver version of PaperCrypt used to generate the document
	Version string `json:"Version"`
}

type versionContainer struct {
	// Version should contain the semver version of PaperCrypt used to generate the document
	Version string `json:"v"`
}

// scanCmd represents the data command.
var scanCmd = &cobra.Command{
	Aliases:      []string{"q", "qr", "scan"},
	Args:         cobra.MaximumNArgs(1),
	SilenceUsage: true,
	Use:          "scan <input>",
	Short:        "Decode a document from a 2D code (aztec or qr).",
	Long: `Decode a document from a 2D code (aztec or qr).

This command allows you to decode data saved by PaperCrypt.
The Aztec/QR code in a PaperCrypt document contains a JSON serialized object
that contains the encrypted data and the PaperCrypt metadata.

If you have trouble scanning the QR code with this command,
you may also try a QR code scanner app on your phone or tablet,
such as "Scandit" (https://apps.apple.com/de/app/scandit-barcode-scanner/id453880584
or https://play.google.com/store/apps/details?id=com.scandit.demoapp).
The resulting JSON data can be read by this command, by supplying the --json flag.
`,
	Example: `papercrypt scan ./code.png | papercrypt decode -o ./out.json -P passphrase`,
	RunE: func(_ *cobra.Command, args []string) error {
		// 1. get data from either argument or inFileName
		if len(args) != 0 {
			inFileName = args[0]
		}

		inFile, err := internal.PrintInputAndGetReader(inFileName)
		if err != nil {
			return err
		}

		var data []byte

		if qrCmdFromJSON {
			data, err = io.ReadAll(inFile)
			if err != nil && err != io.EOF {
				return errors.Join(errors.New("error reading input file"), err)
			}
		} else {
			img, _, err := image.Decode(inFile)
			if err != nil {
				return errors.Join(errors.New("error decoding image"), err)
			}

			bmp, err := gozxing.NewBinaryBitmapFromImage(img)
			if err != nil {
				return errors.Join(errors.New("error creating binary bitmap"), err)
			}

			// attempt to decode as aztec first
			aztecReader := aztec.NewAztecReader()
			result, err := aztecReader.Decode(bmp, nil)
			if err != nil {
				log.Debugf("error decoding aztec: %s", err)
				// if that fails, try qrcode
				qrReader := qrcode.NewQRCodeReader()
				result, err = qrReader.Decode(bmp, nil)
				if err != nil {
					return errors.Join(errors.New("error decoding QR code"), err)
				}
				log.Debug("decoded as QR code")
			}

			data = []byte(result.GetText())
		}

		if err := internal.CloseFileIfNotStd(inFile); err != nil {
			return errors.Join(errors.New("error closing input file"), err)
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

		if qrCmdToJSON {
			n, err := outFile.Write(data)
			if err != nil {
				return errors.Join(errors.New("error writing output"), err)
			}

			internal.PrintWrittenSizeToDebug(n, outFile)
			return nil
		}

		// 3. Deserialize
		var output []byte
		var paperCryptMajorVersion internal.PaperCryptContainerVersion

		// decode version information or find .Data.Data (string)
		vc := versionContainerV1{}
		err = json.Unmarshal(data, &vc)
		if err != nil {
			return errors.Join(errors.New("error deserializing version"), err)
		}

		paperCryptMajorVersion = internal.PaperCryptContainerVersionFromString(vc.Version)

		if paperCryptMajorVersion == internal.PaperCryptContainerVersionUnknown {
			vc := versionContainer{}
			err = json.Unmarshal(data, &vc)
			if err != nil {
				return errors.Join(errors.New("error deserializing version"), err)
			}

			paperCryptMajorVersion = internal.PaperCryptContainerVersionFromString(vc.Version)
		}

		switch paperCryptMajorVersion {
		case internal.PaperCryptContainerVersionMajor1:
			pc := pcv1.PaperCrypt{} // Use the v1 package for PaperCrypt v1, as we do not need to have the serialization code here
			err = json.Unmarshal(data, &pc)
			if err != nil {
				return errors.Join(
					errors.New("error deserializing json data as PaperCrypt v1"),
					err,
				)
			}

			output, err = pc.GetText(false)
			if err != nil {
				return errors.Join(errors.New("error reserializing data as PaperCrypt text"), err)
			}
		case internal.PaperCryptContainerVersionDevel,
			internal.PaperCryptContainerVersionMajor2:
			pc := internal.PaperCrypt{}
			err = json.Unmarshal(data, &pc)
			if err != nil {
				return errors.Join(
					errors.New("error deserializing json data as PaperCrypt v2"),
					err,
				)
			}

			output, err = pc.GetText(false)
			if err != nil {
				return errors.Join(errors.New("error reserializing data as PaperCrypt text"), err)
			}
		default:
			return errors.New("unknown version")
		}

		// 6. Write to file
		n, err := outFile.Write(output)
		if err != nil {
			return errors.Join(errors.New("error writing output"), err)
		}

		internal.PrintWrittenSizeToDebug(n, outFile)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(scanCmd)

	scanCmd.Flags().
		BoolVarP(&qrCmdFromJSON, "from-json", "j", false, "Read input from JSON instead of an image")
	scanCmd.Flags().
		BoolVarP(&qrCmdToJSON, "to-json", "J", false, "Write JSON output instead of plaintext, this cannot be used in the decode command (yet).")
}
