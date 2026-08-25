/*
 * This file is part of PaperCrypt.
 *
 * PaperCrypt lets you prepare encrypted messages for printing on paper.
 * Copyright (C) 2026 TMUniversal <me@tmuniversal.eu>.
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

// Package cmd implements CLI commands and basic functionality around executing them
package cmd

import (
	"errors"
	"os"

	"github.com/caarlos0/log"
	"github.com/spf13/cobra"
	"github.com/tmuniversal/papercrypt/v2/internal"
)

var (
	ignoreVersionMismatchJSON  bool
	ignoreChecksumMismatchJSON bool
)

var decodeJSONCmd = &cobra.Command{
	Aliases:      []string{"decj", "dj"},
	Args:         cobra.NoArgs,
	SilenceUsage: true,
	Use:          "decode-json",
	Short:        "Decode a PaperCrypt document from JSON",
	Long: `This command allows you to decode JSON data saved by PaperCrypt (e.g. from a QR code scan).
The JSON data should be read from a file or stdin, you will be required to provide a passphrase.
This is useful when scanning QR codes with external software that outputs JSON.`,
	Example: `papercrypt decode-json -i <file>.json -o <file>.txt`,
	RunE: func(cmd *cobra.Command, _ []string) error {
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

		paperCryptFileContents, err := internal.PrintInputAndRead(inFileName)
		if err != nil {
			return err
		}

		var pc internal.PaperCrypt
		err = pc.UnmarshalJSON(paperCryptFileContents)
		if err != nil {
			return errors.Join(errors.New("error deserializing JSON data as PaperCrypt"), err)
		}

		paperCryptMajorVersion := internal.PaperCryptContainerVersionFromString(pc.Version)

		if paperCryptMajorVersion == internal.PaperCryptContainerVersionUnknown {
			return errors.New("unknown version")
		}

		var passphraseBytes []byte
		if !cmd.Flags().Lookup("passphrase").Changed {
			cmd.Println(
				"Enter your decryption passphrase (the passphrase you used to encrypt the data)",
			)
			passphraseBytes, err = internal.SensitivePrompt()
			if err != nil {
				return errors.Join(errors.New("error reading passphrase"), err)
			}
		} else {
			passphraseBytes = []byte(passphrase)
		}
		passphrase = ""

		decoded, err := pc.Decode(passphraseBytes)
		if err != nil {
			return errors.Join(errors.New("error decrypting data"), err)
		}

		n, err := outFile.Write(decoded)
		if err != nil {
			return errors.Join(errors.New("error writing to file"), err)
		}

		internal.PrintWrittenSizeToDebug(n, outFile)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(decodeJSONCmd)

	decodeJSONCmd.Flags().
		BoolVar(&ignoreVersionMismatchJSON, "ignore-version-mismatch", false, "Ignore version mismatch and continue anyway")
	decodeJSONCmd.Flags().
		BoolVar(&ignoreChecksumMismatchJSON, "ignore-header-checksum-mismatch", false, "Ignore header checksum mismatches and continue anyway")

	decodeJSONCmd.Flags().
		StringVarP(&passphrase, "passphrase", "P", "", "Passphrase to use for encryption (not recommended, will be prompted for if not provided)")
}
