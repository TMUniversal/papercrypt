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
	"bytes"
	"compress/gzip"
	"errors"
	"os"
	"time"

	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/caarlos0/log"
	"github.com/spf13/cobra"
	"github.com/tmuniversal/papercrypt/v2/internal"
)

var (
	serialNumber string
	purpose      string
	comment      string
	date         string
)

var (
	noQR             bool
	lowerCasedBase16 bool
	rawData          bool
)

var passphrase string

// generateCmd represents the generate command.
var generateCmd = &cobra.Command{
	Aliases:      []string{"gen", "g"},
	Args:         cobra.NoArgs,
	SilenceUsage: true,
	Use:          "generate",
	Short:        "Generate a PaperCrypt document",
	Long: `The 'generate' command takes a JSON file as input and encrypts the data within. It then embeds the encrypted data in a 
newly created PDF file that you can print for physical storage.

Please note, to decrypt the data from the output PaperCrypt PDF, you'll need the original passphrase used during the 
encryption process. Treat this passphrase with care; loss of the passphrase could result in the permanent loss of the 
encrypted data.`,
	Example: "papercrypt generate -i <file>.json -o <file>.pdf --purpose \"My secret data\" --comment \"This is a comment\" --date \"2021-01-01 12:00:00\"",
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

		// 2. generate serial number if not provided
		if serialNumber == "" {
			var err error
			serialNumber, err = internal.GenerateSerial(6)
			if err != nil {
				return errors.Join(errors.New("error generating serial number"), err)
			}
		}

		// 3. parse date if provided
		var timestamp time.Time
		if date == "" {
			timestamp = time.Now()
		} else {
			var err error
			timestamp, err = time.Parse(internal.TimeStampFormatLong, date)
			if err != nil {
				// try other formats if this fails
				timestamp, err = time.Parse(internal.TimeStampFormatShort, date)
				if err != nil {
					timestamp, err = time.Parse(internal.TimeStampFormatDate, date)
					if err != nil {
						return errors.Join(errors.New("error parsing date"), err)
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
			passphraseBytes, err = internal.SensitivePrompt()
			if err != nil {
				return errors.Join(errors.New("error reading passphrase"), err)
			}

			log.Info("Enter your passphrase again to confirm")
			passphraseAgain, err := internal.SensitivePrompt()
			if err != nil {
				return errors.Join(errors.New("error reading passphrase"), err)
			}
			if string(passphraseBytes) != string(passphraseAgain) {
				return errors.New("passphrases do not match")
			}
		} else {
			passphraseBytes = []byte(passphrase)
		}

		// 6. Compress secret data
		compressedData := new(bytes.Buffer)
		gzipWriter, err := gzip.NewWriterLevel(compressedData, gzip.BestCompression)
		if err != nil {
			return errors.Join(errors.New("error creating gzip writer"), err)
		}

		_, err = gzipWriter.Write(secretContentsFile)
		if err != nil {
			return errors.Join(errors.New("error writing to gzip writer"), err)
		}
		if err := gzipWriter.Close(); err != nil {
			return errors.Join(errors.New("error closing gzip writer"), err)
		}

		var data []byte

		// 7. Encrypt with passphrase
		if !rawData {
			encryptedSecretContents, err := encrypt(passphraseBytes, compressedData.Bytes())
			if err != nil {
				return errors.Join(errors.New("error encrypting secret contents"), err)
			}

			compressedData.Reset()
			gzipWriter.Reset(compressedData)
			_, err = gzipWriter.Write(encryptedSecretContents.GetBinary())
			if err != nil {
				return errors.Join(errors.New("error writing to gzip writer"), err)
			}
			if err := gzipWriter.Close(); err != nil {
				return errors.Join(errors.New("error closing gzip writer"), err)
			}
		}

		// Take the unencrypted, compressed data (if rawData is true) or the encrypted, re-compressed data
		data = compressedData.Bytes()

		// 8. Write encryptedSecretContents to outFile
		format := internal.PaperCryptDataFormatPGP
		if rawData {
			format = internal.PaperCryptDataFormatRaw
		}
		crypt := internal.NewPaperCrypt(
			internal.VersionInfo.GitVersion,
			data,
			serialNumber,
			purpose,
			comment,
			timestamp,
			format,
		)

		var text []byte

		text, err = crypt.GetPDF(noQR, lowerCasedBase16)
		if err != nil {
			return errors.Join(errors.New("error generating PDF"), err)
		}

		n, err := outFile.Write(text)
		if err != nil {
			return errors.Join(errors.New("error writing to file"), err)
		}

		internal.PrintWrittenSizeToDebug(n, outFile)
		return nil
	},
}

func encrypt(passphrase []byte, data []byte) (*crypto.PGPMessage, error) {
	message := crypto.NewPlainMessage(data)

	encrypted, err := crypto.EncryptMessageWithPassword(message, passphrase)
	if err != nil {
		return nil, errors.Join(errors.New("error encrypting message"), err)
	}

	return encrypted, nil
}

func init() {
	rootCmd.AddCommand(generateCmd)

	generateCmd.Flags().
		StringVarP(&serialNumber, "serial-number", "s", "", "Serial number of the sheet (optional, default: 6 random characters)")
	generateCmd.Flags().StringVarP(&purpose, "purpose", "p", "", "Purpose of the sheet (optional)")
	generateCmd.Flags().StringVarP(&comment, "comment", "c", "", "Comment on the sheet (optional)")
	generateCmd.Flags().
		StringVarP(&date, "date", "d", "", "Date of the sheet (optional, defaults to now)")
	generateCmd.Flags().BoolVar(&noQR, "no-qr", false, "Do not generate 2D code (optional)")
	generateCmd.Flags().
		BoolVar(&lowerCasedBase16, "lowercase", false, "Whether to use lower case letters for hexadecimal digits")
	generateCmd.Flags().BoolVar(&rawData, "raw", false, "Do not encrypt the data, just compress it")

	generateCmd.Flags().
		StringVarP(&passphrase, "passphrase", "P", "", "Passphrase to use for encryption. Not recommended, will be prompted for if not provided")
}
