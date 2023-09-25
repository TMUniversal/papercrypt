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
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/caarlos0/log"
	"github.com/spf13/cobra"
	"github.com/tmuniversal/papercrypt/internal"
)

var (
	ignoreVersionMismatch  bool
	ignoreChecksumMismatch bool
)

var (
	errorParsingHeader     = errors.New("error parsing header")
	errorParsingBody       = errors.New("error parsing body")
	errorValidationFailure = errors.New("validation failure")
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
	RunE: func(cmd *cobra.Command, args []string) error {
		// 1. Open output file
		outFile, err := internal.GetFileHandleCarefully(outFileName, overrideOutFile)
		if err != nil {
			return err
		}

		// 2. Read inFile
		paperCryptFileContents, err := internal.PrintInputAndRead(inFileName)
		if err != nil {
			return err
		}

		// 2.1 Normalize line endings
		paperCryptFileContents = bytes.ReplaceAll(paperCryptFileContents, []byte("\r\n"), []byte("\n"))
		paperCryptFileContents = bytes.ReplaceAll(paperCryptFileContents, []byte("\r"), []byte("\n"))

		// 2.2 Split into header and body
		paperCryptFileContentsSplit := bytes.SplitN(paperCryptFileContents, []byte("\n\n\n"), 2)

		// 3. Read Headers if present
		var headers map[string]string
		if len(paperCryptFileContentsSplit) == 2 {
			headers = make(map[string]string)

			headerLines := bytes.Split(paperCryptFileContentsSplit[0], []byte("\n"))
			for _, headerLine := range headerLines {
				headerLineSplit := bytes.SplitN(headerLine, []byte(": "), 2)
				if len(headerLineSplit) != 2 {
					return errors.Join(errorParsingHeader, fmt.Errorf("error parsing header line: %s", headerLine))
				}

				key := string(headerLineSplit[0])
				key = strings.TrimPrefix(key, "# ")

				headers[key] = string(headerLineSplit[1])
			}
		}

		// 4. Run Header Validation
		versionLine, ok := headers[internal.HeaderFieldVersion]
		if !ok {
			if !ignoreVersionMismatch {
				return errors.Join(errorParsingHeader, newFieldNotPresentError(internal.HeaderFieldVersion))
			}

			log.Warn("PaperCrypt Version not present in header.")
		}

		// parse git-describe version, look for major version <= 1
		// releases are tagged as vX.Y.Z
		majorVersion := strings.Split(versionLine, ".")[0]
		majorVersion = strings.TrimPrefix(majorVersion, "v")
		if !ignoreVersionMismatch && majorVersion != "1" && majorVersion != "devel" {
			return errors.Join(errorParsingHeader, fmt.Errorf("unsupported PaperCrypt version '%s'", versionLine))
		}

		headerCrc, ok := headers[internal.HeaderFieldHeaderCRC32]
		if !ok {
			if !ignoreChecksumMismatch {
				return errors.Join(errorParsingHeader, newFieldNotPresentError(internal.HeaderFieldHeaderCRC32))
			}

			log.Warn("Header CRC-32 not present in header")
		}

		headerCrc = strings.ToLower(headerCrc)
		headerCrc = strings.ReplaceAll(headerCrc, "0x", "")
		headerCrc = strings.ReplaceAll(headerCrc, " ", "")
		headerCrc32, err := internal.ParseHexUint32(headerCrc)
		if err != nil {
			return errors.Join(errorParsingHeader, errors.New("invalid CRC-32 format"), err)
		}

		headerWithoutCrc := bytes.ReplaceAll(paperCryptFileContentsSplit[0], []byte("\n"+internal.HeaderFieldHeaderCRC32+": "+headers[internal.HeaderFieldHeaderCRC32]), []byte(""))

		if !internal.ValidateCRC32(headerWithoutCrc, headerCrc32) {
			if !ignoreChecksumMismatch {
				return errors.Join(errorParsingHeader, errorValidationFailure, errors.New("header CRC-32 mismatch"))
			}

			log.Warn("Header CRC-32 mismatch!")
		}

		var pgpMessage *crypto.PGPMessage
		var body []byte
		body, err = internal.DeserializeBinary(&paperCryptFileContentsSplit[1])
		if err == nil {
			pgpMessage = crypto.NewPGPMessage(body)
		}

		if err != nil {
			return errors.Join(errorParsingBody, err)
		}

		// 5. Verify Body Hashes
		body = pgpMessage.GetBinary()

		// 5.1 Verify Content Length
		bodyLength, ok := headers[internal.HeaderFieldContentLength]
		if !ok {
			return errors.Join(errorParsingBody, newFieldNotPresentError(internal.HeaderFieldContentLength))
		}

		if fmt.Sprint(len(body)) != bodyLength {
			return errors.Join(errorValidationFailure, fmt.Errorf("`%s` mismatch: expected %s, got %d", internal.HeaderFieldContentLength, bodyLength, len(body)))
		}

		// 5.2 Verify CRC-32
		bodyCrc32, ok := headers[internal.HeaderFieldCRC32]
		if !ok {
			return errors.Join(errorValidationFailure, newFieldNotPresentError(internal.HeaderFieldCRC32))
		}

		bodyCrc32Uint32, err := internal.ParseHexUint32(bodyCrc32)
		if err != nil {
			return errors.Join(errorParsingBody, err)
		}

		if !internal.ValidateCRC32(body, bodyCrc32Uint32) {
			if !ignoreChecksumMismatch {
				return errors.Join(errorValidationFailure, fmt.Errorf("`%s` mismatch", internal.HeaderFieldCRC32))
			}

			log.Warn("Content CRC-32 mismatch!")
		}

		// 5.3 Verify CRC-24
		bodyCrc24, ok := headers[internal.HeaderFieldCRC24]
		if !ok {
			return errors.Join(errorParsingBody, newFieldNotPresentError(internal.HeaderFieldCRC24))
		}

		bodyCrc24Uint32, err := internal.ParseHexUint32(bodyCrc24)
		if err != nil {
			return errors.Join(errorParsingBody, err)
		}

		if !internal.ValidateCRC24(body, bodyCrc24Uint32) {
			if !ignoreChecksumMismatch {
				return errors.Join(errorValidationFailure, fmt.Errorf("`%s` mismatch", internal.HeaderFieldCRC24))
			}

			log.Warn("Content CRC-24 mismatch!")
		}

		// 5.4 Verify SHA-256
		bodySha256, ok := headers[internal.HeaderFieldSHA256]
		if !ok {
			return errors.Join(errorParsingBody, newFieldNotPresentError(internal.HeaderFieldSHA256))
		}

		bodySha256Bytes, err := internal.BytesFromBase64(bodySha256)
		if err != nil {
			return errors.Join(errorParsingBody, err)
		}

		actualSha256 := sha256.Sum256(body)
		if !bytes.Equal(actualSha256[:], bodySha256Bytes) {
			if !ignoreChecksumMismatch {
				return errors.Join(errorValidationFailure, fmt.Errorf("`%s` mismatch", internal.HeaderFieldSHA256))
			}

			log.Warn("Content SHA-256 mismatch!")
		}

		// 6. Construct PaperCrypt object
		headerDate, ok := headers[internal.HeaderFieldDate]
		if !ok {
			log.Warn("Date not present in header!")
		}

		timestamp, err := time.Parse("Mon, 02 Jan 2006 15:04:05.000000000 MST", headerDate)
		if err != nil {
			return errors.Join(errors.New("invalid date format"), err)
		}

		// we don't need to pass the checksums, as they are already verified
		// and will just be recalculated
		paperCrypt := internal.NewPaperCrypt(
			versionLine,
			pgpMessage,
			headers[internal.HeaderFieldSerial],
			headers[internal.HeaderFieldPurpose],
			headers[internal.HeaderFieldComment],
			timestamp,
		)

		// 7. Serialize PaperCrypt object
		_, err = json.MarshalIndent(paperCrypt, "", "  ")
		if err != nil {
			return errors.Join(errors.New("error encoding JSON"), err)
		}
		log.WithField("json", paperCrypt).Debug("Serialized PaperCrypt document")

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
		decryptedContents, err := crypto.DecryptMessageWithPassword(pgpMessage, passphraseBytes)
		if err != nil {
			return errors.Join(errors.New("error decrypting secret contents"), err)
		}

		// 10. Decompress content
		gzipReader, err := gzip.NewReader(bytes.NewReader(decryptedContents.GetBinary()))
		if err != nil {
			return errors.Join(errors.New("error creating gzip reader"), err)
		}

		decompressed := new(bytes.Buffer)
		if _, err := decompressed.ReadFrom(gzipReader); err != nil {
			return errors.Join(errors.New("error reading from gzip reader"), err)
		}
		if err := gzipReader.Close(); err != nil {
			return errors.Join(errors.New("error closing gzip reader"), err)
		}
		decryptedContents = nil // clear decryptedContents

		// 11. Write decompressed to outFile
		n, err := outFile.Write(decompressed.Bytes())
		if err != nil {
			return errors.Join(errors.New("error writing to file"), err)
		}

		internal.PrintWrittenSize(n, outFile)
		return internal.CloseFileIfNotStd(outFile)
	},
}

func newFieldNotPresentError(field string) error {
	return fmt.Errorf("`%s` not present in header", field)
}

func init() {
	rootCmd.AddCommand(decodeCmd)

	decodeCmd.Flags().BoolVar(&ignoreVersionMismatch, "ignore-version-mismatch", false, "Ignore version mismatch and continue anyway")
	decodeCmd.Flags().BoolVar(&ignoreChecksumMismatch, "ignore-header-checksum-mismatch", false, "Ignore header checksum mismatches and continue anyway")

	decodeCmd.Flags().StringVarP(&passphrase, "passphrase", "P", "", "Passphrase to use for encryption (not recommended, will be prompted for if not provided)")
}
