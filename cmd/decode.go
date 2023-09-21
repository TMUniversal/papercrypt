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
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/tmuniversal/papercrypt/internal"
)

var ignoreVersionMismatch bool
var ignoreChecksumMismatch bool

// decodeCmd represents the decode command
var decodeCmd = &cobra.Command{
	Aliases: []string{"dec", "d"},
	Args:    cobra.NoArgs,
	Use:     "decode",
	Short:   "Decode a PaperCrypt document",
	Long: `This command allows you to decode binary data saved by PaperCrypt. 
The data should be read from a file or stdin, you will be required to provide a passphrase.`,
	Example: `papercrypt decode -i <file>.txt -o <file>.txt`,
	Run: func(cmd *cobra.Command, args []string) {
		// 1. Open output file
		outFile := internal.GetFileHandleCarefully(cmd, outFileName, overrideOutFile)
		defer outFile.Close()

		// 2. Read inFile
		paperCryptFileContents := internal.PrintInputAndRead(cmd, inFileName)

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
					internal.Fatal(cmd, errors.Errorf("Error parsing header line: %s", headerLine))
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
				internal.Fatal(cmd, errors.Errorf("Error parsing headers: PaperCrypt Version not present in header."))
			} else {
				cmd.Println("Warning: PaperCrypt Version not present in header.")
			}
		}

		// parse git-describe version, look for major version <= 1
		// releases are tagged as vX.Y.Z
		majorVersion := strings.Split(versionLine, ".")[0]
		majorVersion = strings.TrimPrefix(majorVersion, "v")
		if !ignoreVersionMismatch && majorVersion != "1" {
			internal.Fatal(cmd, errors.Errorf("error parsing headers: unsupported PaperCrypt Version %s", versionLine))
		}

		headerCrc, ok := headers["Header CRC-32"]
		if !ok {
			if !ignoreChecksumMismatch {
				internal.Fatal(cmd, errors.Errorf("error parsing headers: Header CRC-32 not present in header"))
			} else {
				cmd.Println("Warning: Header CRC-32 not present in header")
			}
		}

		headerCrc = strings.ToLower(headerCrc)
		headerCrc = strings.ReplaceAll(headerCrc, "0x", "")
		headerCrc = strings.ReplaceAll(headerCrc, " ", "")
		headerCrc32, err := internal.ParseHexUint32(headerCrc)
		if err != nil {
			internal.Fatal(cmd, errors.Wrap(errors.Wrap(err, "invalid CRC-32 format"), "error parsing headers"))
		}

		headerWithoutCrc := bytes.ReplaceAll(paperCryptFileContentsSplit[0], []byte("\nHeader CRC-32: "+headers[internal.HeaderFieldHeaderCRC32]), []byte(""))

		if !internal.ValidateCRC32(headerWithoutCrc, headerCrc32) {
			if !ignoreChecksumMismatch {
				internal.Fatal(cmd, errors.Wrap(errors.Wrap(errors.Errorf("header CRC-32 mismatch"), "header is invalid"), "error parsing headers"))
			} else {
				cmd.Println("Warning: Header CRC-32 mismatch!")
			}
		}

		var pgpMessage *crypto.PGPMessage
		var body []byte
		body, err = internal.DeserializeBinary(&paperCryptFileContentsSplit[1])
		if err == nil {
			pgpMessage = crypto.NewPGPMessage(body)
		}

		if err != nil {
			internal.Fatal(cmd, errors.Wrap(err, "error parsing body"))
		}

		// 5. Verify Body Hashes
		body = pgpMessage.GetBinary()

		// 5.1 Verify Content Length
		bodyLength, ok := headers[internal.HeaderFieldLength]
		if !ok {
			internal.Fatal(cmd, errors.Errorf("error parsing headers: Content Length not present in header"))
		}

		if fmt.Sprint(len(body)) != bodyLength {
			internal.Fatal(cmd, errors.Errorf("content failed validation: Content Length mismatch"))
		}

		// 5.2 Verify CRC-32
		bodyCrc32, ok := headers[internal.HeaderFieldCRC32]
		if !ok {
			internal.Fatal(cmd, errors.Errorf("error parsing headers: Content CRC-32 not present in header"))
		}

		bodyCrc32Uint32, err := internal.ParseHexUint32(bodyCrc32)
		if err != nil {
			internal.Fatal(cmd, errors.Wrap(err, "error parsing headers"))
		}

		if !internal.ValidateCRC32(body, bodyCrc32Uint32) {
			if !ignoreChecksumMismatch {
				internal.Fatal(cmd, errors.Errorf("content failed validation: Content CRC-32 mismatch"))
			} else {
				cmd.Println("Warning: Content CRC-32 mismatch!")
			}
		}

		// 5.3 Verify CRC-24
		bodyCrc24, ok := headers[internal.HeaderFieldCRC24]
		if !ok {
			internal.Fatal(cmd, errors.Errorf("error parsing headers: Content CRC-24 not present in header"))
		}

		bodyCrc24Uint32, err := internal.ParseHexUint32(bodyCrc24)
		if err != nil {
			internal.Fatal(cmd, errors.Wrap(err, "error parsing headers"))
		}

		if !internal.ValidateCRC24(body, bodyCrc24Uint32) {
			if !ignoreChecksumMismatch {
				internal.Fatal(cmd, errors.Errorf("content failed validation: Content CRC-24 mismatch"))
			} else {
				cmd.Println("Warning: Content CRC-24 mismatch!")
			}
		}

		// 5.4 Verify SHA-256
		bodySha256, ok := headers[internal.HeaderFieldSHA256]
		if !ok {
			internal.Fatal(cmd, errors.Errorf("error parsing headers: Content SHA-256 not present in header"))
		}

		bodySha256Bytes, err := internal.BytesFromBase64(bodySha256)
		if err != nil {
			internal.Fatal(cmd, errors.Wrap(err, "error parsing headers"))
		}

		actualSha256 := sha256.Sum256(body)
		if !bytes.Equal(actualSha256[:], bodySha256Bytes) {
			if !ignoreChecksumMismatch {
				internal.Fatal(cmd, errors.Errorf("content failed validation: Content SHA-256 mismatch"))
			} else {
				cmd.Println("Warning: Content SHA-256 mismatch!")
			}
		}

		// 6. Construct PaperCrypt object
		headerDate, ok := headers[internal.HeaderFieldDate]
		if !ok {
			cmd.Println("Warning: Date not present in header!")
		}

		timestamp, err := time.Parse("Mon, 02 Jan 2006 15:04:05.000000000 MST", headerDate)
		if err != nil {
			internal.Fatal(cmd, errors.Wrap(err, "invalid date format"))
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
			internal.Fatal(cmd, errors.Wrap(err, "error encoding JSON"))
		}
		//cmd.Printf("Understood the following PaperCrypt document: %s\n", jsonBytes) // print the JSON

		// 8. Read passphrase from stdin
		if !cmd.Flags().Lookup("passphrase").Changed {
			cmd.Println("Enter your decryption passphrase (the passphrase you used to encrypt the data)")
			cmd.Printf("Passphrase: ")
			passphrase, err = internal.ReadTtyLine()
			if err != nil {
				internal.Fatal(cmd, errors.Wrap(err, "error reading passphrase"))
			}
		}

		// 9. Decrypt secretContents
		decryptedContents, err := crypto.DecryptMessageWithPassword(pgpMessage, []byte(passphrase))
		if err != nil {
			internal.Fatal(cmd, errors.Wrap(err, "error decrypting secret contents"))
		}

		passphrase = "" // clear passphrase

		// 10. Write decryptedContents to outFile
		n, err := outFile.Write(decryptedContents.GetBinary())
		if err != nil {
			internal.Fatal(cmd, errors.Wrap(err, "error writing to file"))
		}

		internal.PrintWrittenSize(cmd, n, outFile)
	},
}

func init() {
	rootCmd.AddCommand(decodeCmd)

	decodeCmd.Flags().BoolVar(&ignoreVersionMismatch, "ignore-version-mismatch", false, "Ignore version mismatch and continue anyway")
	decodeCmd.Flags().BoolVar(&ignoreChecksumMismatch, "ignore-header-checksum-mismatch", false, "Ignore header checksum mismatches and continue anyway")

	decodeCmd.Flags().StringVarP(&passphrase, "passphrase", "P", "", "Passphrase to use for encryption (not recommended, will be prompted for if not provided)")
}
