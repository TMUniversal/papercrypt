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
	"compress/gzip"
	"os"
	"time"

	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/caarlos0/log"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/tmuniversal/papercrypt/internal"
	"golang.org/x/term"
)

var serialNumber string
var purpose string
var comment string
var date string

var noQR bool
var lowerCasedBase16 bool

var passphrase string

// generateCmd represents the generate command
var generateCmd = &cobra.Command{
	Aliases: []string{"gen", "g"},
	Args:    cobra.NoArgs,
	Use:     "generate",
	Short:   "Generate a PaperCrypt document",
	Long: `The 'generate' command takes a JSON file as input and encrypts the data within. It then embeds the encrypted data in a 
newly created PDF file that you can print for physical storage.

Please note, to decrypt the data from the output PaperCrypt PDF, you'll need the original passphrase used during the 
encryption process. Treat this passphrase with care; loss of the passphrase could result in the permanent loss of the 
encrypted data.`,
	Example: "papercrypt generate -i <file>.json -o <file>.pdf --purpose \"My secret data\" --comment \"This is a comment\" --date \"2021-01-01 12:00:00\"",
	RunE: func(cmd *cobra.Command, args []string) error {
		// 1. Open output file
		outFile, err := internal.GetFileHandleCarefully(outFileName, overrideOutFile)
		if err != nil {
			return err
		}

		// 2. generate serial number if not provided
		if serialNumber == "" {
			var err error
			serialNumber, err = internal.GenerateSerial(6)
			if err != nil {
				return errors.Wrap(err, "error generating serial number")
			}
		}

		// 3. parse date if provided
		var timestamp time.Time
		if date == "" {
			timestamp = time.Now()
		} else {
			var err error
			timestamp, err = time.Parse("Mon, 02 Jan 2006 15:04:05.000000000 MST", date)
			if err != nil {
				// try other formats if this fails
				timestamp, err = time.Parse("2006-01-02 15:04:05", date)
				if err != nil {
					timestamp, err = time.Parse("2006-01-02", date)
					if err != nil {
						return errors.Wrap(err, "error parsing date")
					}
				}
			}
		}

		// 4. Read input file as bytes
		secretContentsFile, err := internal.PrintInputAndRead(inFileName)
		if err != nil {
			return err
		}

		// 5. Read passphrase from stdin
		var passphraseBytes []byte
		if !cmd.Flags().Lookup("passphrase").Changed {
			log.Info("Enter your encryption passphrase")
			cmd.Printf("Passphrase: ")
			passphraseBytes, err = term.ReadPassword(int(os.Stdin.Fd()))
			if err != nil {
				return errors.Wrap(err, "error reading passphrase")
			}

			log.Info("Enter your encryption passphrase again to confirm")
			cmd.Printf("Passphrase (again): ")
			passphraseAgain, err := term.ReadPassword(int(os.Stdin.Fd()))
			if err != nil {
				return errors.Wrap(err, "error reading passphrase")
			}
			if string(passphraseBytes) != string(passphraseAgain) {
				return errors.New("passphrases do not match")
			}
			passphraseAgain = nil // clear passphraseAgain
		} else {
			passphraseBytes = []byte(passphrase)
		}
		passphrase = "" // clear passphrase

		// 6. Compress secret data
		compressedData := new(bytes.Buffer)
		gzipWriter, err := gzip.NewWriterLevel(compressedData, gzip.BestCompression)
		if err != nil {
			return errors.Wrap(err, "error creating gzip writer")
		}

		_, err = gzipWriter.Write(secretContentsFile)
		if err != nil {
			return errors.Wrap(err, "error writing to gzip writer")
		}
		if err := gzipWriter.Close(); err != nil {
			return errors.Wrap(err, "error closing gzip writer")
		}

		secretContentsFile = nil // clear secretContentsFile

		// 7. Encrypt secretContentsMinimal with passphrase
		encryptedSecretContents, err := encrypt(passphraseBytes, compressedData.Bytes())
		if err != nil {
			return errors.Wrap(err, "error encrypting secret contents")
		}

		compressedData = nil  // clear compressedData
		passphraseBytes = nil // clear passphraseBytes

		// 8. Write encryptedSecretContents to outFile
		crypt := internal.NewPaperCrypt(internal.VersionInfo.GitVersion, encryptedSecretContents, serialNumber, purpose, comment, timestamp)

		var text []byte

		text, err = crypt.GetPDF(noQR, lowerCasedBase16)
		if err != nil {
			return errors.Wrap(err, "error generating PDF")
		}

		n, err := outFile.Write(text)
		if err != nil {
			return errors.Wrap(err, "error writing to file")
		}

		internal.PrintWrittenSize(n, outFile)

		if err := outFile.Close(); err != nil {
			return errors.Wrap(err, "error closing file")
		}

		return nil
	},
}

func encrypt(passphrase []byte, data []byte) (*crypto.PGPMessage, error) {
	var message = crypto.NewPlainMessage(data)

	encrypted, err := crypto.EncryptMessageWithPassword(message, passphrase)
	if err != nil {
		return nil, errors.Wrap(err, "error encrypting message")
	}

	return encrypted, nil
}

func init() {
	rootCmd.AddCommand(generateCmd)

	generateCmd.Flags().StringVarP(&serialNumber, "serial-number", "s", "", "Serial number of the sheet (optional, default: 6 random characters)")
	generateCmd.Flags().StringVarP(&purpose, "purpose", "p", "", "Purpose of the sheet (optional)")
	generateCmd.Flags().StringVarP(&comment, "comment", "c", "", "Comment on the sheet (optional)")
	generateCmd.Flags().StringVarP(&date, "date", "d", "", "Date of the sheet (optional, defaults to now)")
	generateCmd.Flags().BoolVar(&noQR, "no-qr", false, "Do not generate QR code (optional)")
	generateCmd.Flags().BoolVar(&lowerCasedBase16, "lowercase", false, "Whether to use lower case letters for hexadecimal digits")

	generateCmd.Flags().StringVarP(&passphrase, "passphrase", "P", "", "Passphrase to use for encryption. Not recommended, will be prompted for if not provided")
}
