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
	"fmt"
	"image"
	"io"
	"os"

	"github.com/caarlos0/log"
	"github.com/spf13/cobra"
	"github.com/tmuniversal/papercrypt/v3/internal"
	"github.com/tmuniversal/papercrypt/v3/internal/codematrix"
)

var (
	qrCmdFromJSON = false
	qrCmdToJSON   = false
)

type versionContainer struct {
	// Version should contain the semver version of PaperCrypt used to generate the document
	Version string `json:"v"`
}

// scanCmd represents the data command.
var scanCmd = &cobra.Command{
	Aliases:      []string{"q", "qr", "scan"},
	Args:         cobra.MinimumNArgs(1),
	SilenceUsage: true,
	Use:          "scan <input> [input...]",
	Short:        "Decode a document from Data Matrix code(s).",
	Long: `Decode a document from one or more Data Matrix codes.

This command reads one or more images containing Data Matrix codes and
reassembles the encoded PaperCrypt data. When multiple images are supplied,
they are treated as a single split payload (up to 4 codes).

If you have --from-json, a single JSON-encoded input is accepted instead.`,
	Example: `papercrypt scan ./code1.png ./code2.png | papercrypt decode -o ./out.json -P passphrase`,
	RunE: func(_ *cobra.Command, args []string) error {
		var data []byte
		var err error

		if qrCmdFromJSON {
			inFile, err := internal.PrintInputAndGetReader(inFileName)
			if err != nil {
				return err
			}
			defer func() {
				if err := internal.CloseFileIfNotStd(inFile); err != nil {
					log.WithError(err).Error("Error closing file")
				}
			}()

			data, err = io.ReadAll(inFile)
			if err != nil && err != io.EOF {
				return errors.Join(errors.New("error reading input file"), err)
			}
		} else {
			// Collect image files from args
			filePaths := args
			if len(filePaths) == 0 && inFileName != "" {
				filePaths = []string{inFileName}
			}

			if len(filePaths) == 0 {
				return errors.New("no input files provided")
			}
			if len(filePaths) > codematrix.MaxSymbols {
				return fmt.Errorf("too many input files: %d (max %d)", len(filePaths), codematrix.MaxSymbols)
			}

			images := make([]image.Image, len(filePaths))
			for i, fp := range filePaths {
				f, err := os.Open(fp)
				if err != nil {
					return errors.Join(fmt.Errorf("error opening file %s", fp), err)
				}
				img, _, err := image.Decode(f)
				f.Close()
				if err != nil {
					return errors.Join(fmt.Errorf("error decoding image %s", fp), err)
				}
				images[i] = img
			}

			data, err = codematrix.Decode(images)
			if err != nil {
				return errors.Join(errors.New("error decoding Data Matrix codes"), err)
			}
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

		vc := versionContainer{}
		err = json.Unmarshal(data, &vc)
		if err != nil {
			return errors.Join(errors.New("error deserializing version"), err)
		}

		paperCryptMajorVersion = internal.PaperCryptContainerVersionFromString(vc.Version)

		switch paperCryptMajorVersion {
		case internal.PaperCryptContainerVersionDevel,
			internal.PaperCryptContainerVersionMajor3:
			pc := internal.PaperCrypt{}
			err = json.Unmarshal(data, &pc)
			if err != nil {
				return errors.Join(
					errors.New("error deserializing json data as PaperCrypt"),
					err,
				)
			}

			output, err = pc.GetText(false)
			if err != nil {
				return errors.Join(errors.New("error reserializing data as PaperCrypt text"), err)
			}
		default:
			return fmt.Errorf("unknown version: %s", vc.Version)
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
