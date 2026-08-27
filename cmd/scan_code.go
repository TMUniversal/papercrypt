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
	"compress/gzip"
	"errors"
	"image"
	"io"
	"os"
	"strings"

	"github.com/caarlos0/log"
	"github.com/spf13/cobra"
	"github.com/tmuniversal/papercrypt/v3/internal"
	"github.com/tmuniversal/papercrypt/v3/internal/codematrix"
	"github.com/tmuniversal/papercrypt/v3/internal/file_format"
	"github.com/tmuniversal/papercrypt/v3/internal/file_format/envelope"
	"github.com/tmuniversal/papercrypt/v3/internal/terminal"
)

var (
	qrCmdFromBinary = false
	qrCmdToBinary   = false
)

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
The resulting data can be read by this command, by supplying the --from-binary flag.
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

		var envelopeStr string

		if qrCmdFromBinary {
			data, err := io.ReadAll(inFile)
			if err != nil {
				return errors.Join(errors.New("error reading input file"), err)
			}
			envelopeStr = strings.TrimSpace(string(data))
		} else {
			img, _, err := image.Decode(inFile)
			if err != nil {
				return errors.Join(errors.New("error decoding image"), err)
			}

			envelopeStr, err = codematrix.Decode(img)
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

		// 3. Write raw envelope string (passthrough mode)
		if qrCmdToBinary {
			n, err := outFile.WriteString(envelopeStr)
			if err != nil {
				return errors.Join(errors.New("error writing output"), err)
			}
			terminal.PrintWrittenSizeToDebug(n, outFile)
			return nil
		}

		// 4. Deserialize to text format
		pc, err := deserializePaperCrypt(envelopeStr)
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

// deserializePaperCrypt unwraps an envelope string and returns a PaperCrypt.
func deserializePaperCrypt(data string) (*file_format.PaperCrypt, error) {
	// Try envelope-wrapped binary (format: PCE1 + base45(CRC32) + base45(content))
	if strings.HasPrefix(data, envelope.Magic) {
		content, err := envelope.Unwrap(data, envelope.Base45Encoder{})
		if err != nil {
			return nil, errors.Join(errors.New("error unwrapping envelope"), err)
		}

		gz, err := gzip.NewReader(bytes.NewReader(content))
		if err != nil {
			return nil, errors.Join(errors.New("error creating gzip reader"), err)
		}
		binary, err := io.ReadAll(gz)
		if err != nil {
			return nil, errors.Join(errors.New("error reading gzip data"), err)
		}

		pc, err := file_format.UnmarshalBinary(binary)
		if err != nil {
			return nil, errors.Join(errors.New("error deserializing binary container"), err)
		}
		return pc, nil
	}

	return nil, errors.New("unsupported format: expected PCE1 envelope")
}

func init() {
	rootCmd.AddCommand(scanCmd)

	scanCmd.Flags().
		BoolVarP(&qrCmdFromBinary, "from-binary", "B", false, "Read input as envelope string instead of an image")
	scanCmd.Flags().
		BoolVarP(&qrCmdToBinary, "to-binary", "b", false, "Write envelope string output instead of plaintext")
}
