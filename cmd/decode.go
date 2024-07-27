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
	"errors"
	"os"

	"github.com/caarlos0/log"
	"github.com/spf13/cobra"
	"github.com/tmuniversal/papercrypt/internal"
)

var (
	ignoreVersionMismatch  bool
	ignoreChecksumMismatch bool
)

// decodeCmd represents the decode command
var decodeCmd = &cobra.Command{
	Aliases:      []string{"dec", "d"},
	Args:         cobra.NoArgs,
	SilenceUsage: true,
	Use:          "decode",
	Short:        "Decode a PaperCrypt document",
	Long: `This command allows you to decode binary data saved by PaperCrypt. 
The data should be read from a file or stdin, you will be required to provide a passphrase.`,
	Example: `papercrypt decode -i <file>.txt -o <file>.txt`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		// 1. Open output file
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

		// 2. Read inFile
		paperCryptFileContents, err := internal.PrintInputAndRead(inFileName)
		if err != nil {
			return err
		}

		// TODO: check container version

		pc, err := internal.DeserializeV1Text(paperCryptFileContents, ignoreVersionMismatch, ignoreChecksumMismatch)
		if err != nil {
			return errors.Join(errors.New("error deserializing PaperCrypt document"), err)
		}

		// 8. Read passphrase from stdin
		var passphraseBytes []byte
		if !cmd.Flags().Lookup("passphrase").Changed {
			cmd.Println("Enter your decryption passphrase (the passphrase you used to encrypt the data)")
			passphraseBytes, err = internal.SensitivePrompt()
			if err != nil {
				return errors.Join(errors.New("error reading passphrase"), err)
			}
		} else {
			passphraseBytes = []byte(passphrase)
		}
		passphrase = "" // clear passphrase

		// 9. Decrypt secretContents
		decoded, err := pc.Decode(passphraseBytes)
		if err != nil {
			return errors.Join(errors.New("error decrypting data"), err)
		}

		// 11. Write decompressed to outFile
		n, err := outFile.Write(decoded)
		if err != nil {
			return errors.Join(errors.New("error writing to file"), err)
		}

		internal.PrintWrittenSize(n, outFile)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(decodeCmd)

	decodeCmd.Flags().BoolVar(&ignoreVersionMismatch, "ignore-version-mismatch", false, "Ignore version mismatch and continue anyway")
	decodeCmd.Flags().BoolVar(&ignoreChecksumMismatch, "ignore-header-checksum-mismatch", false, "Ignore header checksum mismatches and continue anyway")

	decodeCmd.Flags().StringVarP(&passphrase, "passphrase", "P", "", "Passphrase to use for encryption (not recommended, will be prompted for if not provided)")
}
