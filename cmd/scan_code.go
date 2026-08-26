/*
 * This file is part of PaperCrypt.
 *
 * PaperCrypt lets you prepare encrypted messages for printing on paper.
 * Copyright (C) 2023-2026 TMUniversal <me@tmuniversal.eu>.
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
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"io"
	"os"

	"github.com/caarlos0/log"
	"github.com/spf13/cobra"
	"github.com/tmuniversal/papercrypt/v3/internal"
	"github.com/tmuniversal/papercrypt/v3/internal/codematrix"
	"github.com/tmuniversal/papercrypt/v3/internal/file_format"
	"github.com/tmuniversal/papercrypt/v3/internal/file_format/envelope"
	"github.com/tmuniversal/papercrypt/v3/internal/terminal"
)

var (
	qrCmdFromJSON   = false
	qrCmdToJSON     = false
	qrCmdFromBinary = false
	qrCmdToBinary   = false
)

type versionContainer struct {
	// Version should contain the semver version of PaperCrypt used to generate the document
	Version string `json:"v"`
}

// isBinaryContainer checks if data starts with the binary container magic.
func isBinaryContainer(data []byte) bool {
	return len(data) >= 4 && bytes.Equal(data[0:4], file_format.BinaryMagic[:])
}

// scanCmd represents the data command.
var scanCmd = &cobra.Command{
	Aliases:      []string{"q", "qr", "scan"},
	Args:         cobra.MaximumNArgs(1),
	SilenceUsage: true,
	Use:          "scan <input>",
	Short:        "Decode a document from a 2D code (QR).",
	Long: `Decode a document from a 2D code (QR).

This command allows you to decode data saved by PaperCrypt.
The 2D code in a PaperCrypt document contains a serialized object
that contains the encrypted data and the PaperCrypt metadata.

If you have trouble scanning the QR code with this command,
you may also try a QR code scanner app on your phone or tablet,
such as "Scandit" (https://apps.apple.com/de/app/scandit-barcode-scanner/id453880584
or https://play.google.com/store/apps/details?id=com.scandit.demoapp).
The resulting data can be read by this command, by supplying the --json or --binary flag.
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

		if qrCmdFromJSON || qrCmdFromBinary {
			data, err = io.ReadAll(inFile)
			if err != nil && err != io.EOF {
				return errors.Join(errors.New("error reading input file"), err)
			}
		} else {
			img, _, err := image.Decode(inFile)
			if err != nil {
				return errors.Join(errors.New("error decoding image"), err)
			}

			data, err = codematrix.Decode(img)
			if err != nil {
				return errors.Join(errors.New("error decoding 2D code"), err)
			}
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

		// 3. Write raw data (passthrough modes)
		if qrCmdToJSON || qrCmdToBinary {
			var out []byte

			if qrCmdToBinary {
				pc, err := deserializePaperCrypt(data)
				if err != nil {
					return err
				}

				bin, err := file_format.MarshalBinary(pc)
				if err != nil {
					return errors.Join(errors.New("error marshalling to binary"), err)
				}

				out = envelope.Wrap(bin)
			} else {
				out = data
			}

			n, err := outFile.Write(out)
			if err != nil {
				return errors.Join(errors.New("error writing output"), err)
			}

			terminal.PrintWrittenSizeToDebug(n, outFile)
			return nil
		}

		// 4. Deserialize to text format
		pc, err := deserializePaperCrypt(data)
		if err != nil {
			return err
		}

		output, err := pc.GetText(false)
		if err != nil {
			return errors.Join(errors.New("error reserializing data as PaperCrypt text"), err)
		}

		// 5. Write to file
		n, err := outFile.Write(output)
		if err != nil {
			return errors.Join(errors.New("error writing output"), err)
		}

		terminal.PrintWrittenSizeToDebug(n, outFile)
		return nil
	},
}

// deserializePaperCrypt auto-detects binary or JSON format and returns a PaperCrypt.
func deserializePaperCrypt(data []byte) (*file_format.PaperCrypt, error) {
	if isBinaryContainer(data) {
		bin, err := file_format.UnmarshalBinary(data)
		if err != nil {
			return nil, errors.Join(errors.New("error deserializing binary container"), err)
		}
		return bin, nil
	}

	// Try envelope-wrapped binary
	if len(data) >= envelope.HeaderSize && bytes.Equal(data[0:4], envelope.Magic[:]) {
		payload, err := envelope.Unwrap(data)
		if err != nil {
			return nil, errors.Join(errors.New("error unwrapping envelope"), err)
		}

		bin, err := file_format.UnmarshalBinary(payload)
		if err != nil {
			return nil, errors.Join(errors.New("error deserializing binary container"), err)
		}
		return bin, nil
	}

	// Fall back to JSON
	vc := versionContainer{}
	err := json.Unmarshal(data, &vc)
	if err != nil {
		return nil, errors.Join(errors.New("error deserializing version"), err)
	}

	paperCryptMajorVersion := file_format.PaperCryptContainerVersionFromString(vc.Version)

	switch paperCryptMajorVersion {
	case file_format.PaperCryptContainerVersionDevel,
		file_format.PaperCryptContainerVersionMajor3:
		pc := file_format.PaperCrypt{}
		err = json.Unmarshal(data, &pc)
		if err != nil {
			return nil, errors.Join(
				errors.New("error deserializing JSON data as PaperCrypt"),
				err,
			)
		}
		return &pc, nil
	default:
		return nil, fmt.Errorf("unknown version: %s", vc.Version)
	}
}

func init() {
	rootCmd.AddCommand(scanCmd)

	scanCmd.Flags().
		BoolVarP(&qrCmdFromJSON, "from-json", "j", false, "Read input from JSON instead of an image")
	scanCmd.Flags().
		BoolVarP(&qrCmdToJSON, "to-json", "J", false, "Write JSON output instead of plaintext")
	scanCmd.Flags().
		BoolVarP(&qrCmdFromBinary, "from-binary", "B", false, "Read input as binary container instead of an image")
	scanCmd.Flags().
		BoolVarP(&qrCmdToBinary, "to-binary", "b", false, "Write binary container output instead of plaintext")
}
