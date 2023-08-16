package cmd

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"papercrypt/util"
	"strings"
	"time"

	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/spf13/cobra"
)

var ignoreVersionMismatch bool
var ignoreChecksumMismatch bool

// decodeCmd represents the decode command
var decodeCmd = &cobra.Command{
	Aliases: []string{"dec", "d"},
	Use:     "decode",
	Short:   "Decode a PaperCrypt document",
	Long: `This command allows you to decode binary data saved by PaperCrypt. 
The data should be read from a file or stdin, you will be required to provide a passphrase.`,
	Example: `papercrypt decode -i <file>.txt -o <file>.txt`,
	Run: func(cmd *cobra.Command, args []string) {
		// 0. check if outfile exists
		if _, err := os.Stat(outFileName); err == nil {
			if overrideOutFile {
				fmt.Printf("Overriding existing file \"%s\"!\n", outFileName)
			} else {
				fmt.Printf("File %s already exists, use --force to override\n", outFileName)
				os.Exit(1)
			}
		}

		if inFileName == "" || inFileName == "-" {
			fmt.Printf("Reading from stdin\n")
		} else {
			fmt.Printf("Reading from %s\n", inFileName)
		}

		// 3. Read inFile
		paperCryptFileContents, err := os.ReadFile(inFileName)
		if err != nil && err != io.EOF {
			fmt.Printf("Error opening file: %s\n", err)
			os.Exit(1)
		}

		// 3.1 Normalize line endings
		paperCryptFileContents = bytes.ReplaceAll(paperCryptFileContents, []byte("\r\n"), []byte("\n"))
		paperCryptFileContents = bytes.ReplaceAll(paperCryptFileContents, []byte("\r"), []byte("\n"))

		// 3.2 Split into header and body
		paperCryptFileContentsSplit := bytes.SplitN(paperCryptFileContents, []byte("\n\n\n"), 2)

		// 4. Read Headers if present
		var headers map[string]string
		if len(paperCryptFileContentsSplit) == 2 {
			headers = make(map[string]string)

			headerLines := bytes.Split(paperCryptFileContentsSplit[0], []byte("\n"))
			for _, headerLine := range headerLines {
				headerLineSplit := bytes.SplitN(headerLine, []byte(": "), 2)
				if len(headerLineSplit) != 2 {
					fmt.Printf("Error parsing header line: %s\n", headerLine)
					os.Exit(1)
				}

				key := string(headerLineSplit[0])
				key = strings.TrimPrefix(key, "# ")

				headers[key] = string(headerLineSplit[1])
			}
		}

		// 5. Run Header Validation

		versionLine, ok := headers["PaperCrypt Version"]
		if !ok {
			if !ignoreVersionMismatch {
				fmt.Println("Error parsing headers: PaperCrypt Version not present in header.")
				os.Exit(1)
			} else {
				fmt.Println("Warning: PaperCrypt Version not present in header.")
			}
		}

		// parse git-describe version, look for major version <= 1
		// releases are tagged as vX.Y.Z
		majorVersion := strings.Split(versionLine, ".")[0]
		majorVersion = strings.TrimPrefix(majorVersion, "v")
		if !ignoreVersionMismatch && majorVersion != "1" {
			fmt.Printf("Error parsing headers: unsupported PaperCrypt Version %s\n", versionLine)
			os.Exit(1)
		}

		headerCrc, ok := headers["Header CRC-32"]
		if !ok {
			if !ignoreChecksumMismatch {
				fmt.Println("Error parsing headers: Header CRC-32 not present in header")
				os.Exit(1)
			} else {
				fmt.Println("Warning: Header CRC-32 not present in header")
			}
		}

		headerCrc = strings.ToLower(headerCrc)
		headerCrc = strings.ReplaceAll(headerCrc, "0x", "")
		headerCrc = strings.ReplaceAll(headerCrc, " ", "")
		headerCrc32, err := util.ParseHexUint32(headerCrc)

		headerWithoutCrc := bytes.ReplaceAll(paperCryptFileContentsSplit[0], []byte("\nHeader CRC-32: "+headers["Header CRC-32"]), []byte(""))

		if !util.ValidateCRC32(headerWithoutCrc, headerCrc32) {
			if !ignoreChecksumMismatch {
				fmt.Printf("Error parsing headers: header is invalid: Header CRC-32 mismatch\n")
				os.Exit(1)
			} else {
				fmt.Printf("Warning: Header CRC-32 mismatch\n")
			}
		}

		// 6. Read Body
		serializationType, ok := headers["Serialization Type"]
		if !ok {
			fmt.Printf("Warning: Serialization Type not present in header, assuming 'papercrypt/base16+crc'\n")
			serializationType = "papercrypt/base16+crc"
		}

		var pgpMessage *crypto.PGPMessage
		if serializationType == "papercrypt/base16+crc" {
			fmt.Println("Decoding body as papercrypt/base16+crc")
			var body []byte
			body, err = util.DeserializeBinary(&paperCryptFileContentsSplit[1])
			if err == nil {
				pgpMessage = crypto.NewPGPMessage(body)
			}
		} else if serializationType == "openpgp/armor" {
			pgpMessage, err = crypto.NewPGPMessageFromArmored(string(paperCryptFileContentsSplit[1]))
		} else {
			fmt.Printf("Error: unknown serialization type %s\n", serializationType)
			os.Exit(1)
		}

		if err != nil {
			fmt.Printf("Error parsing body: %s\n", err)
			os.Exit(1)
		}

		// 7. Verify Body Hashes
		body := pgpMessage.GetBinary()

		// 7.0 Verify Content Length
		bodyLength, ok := headers["Content Length"]
		if !ok {
			fmt.Printf("Error parsing headers: Content Length not present in header\n")
			os.Exit(1)
		}

		if fmt.Sprint(len(body)) != bodyLength {
			fmt.Printf("Validation failure: Content Length mismatch!\n")
			os.Exit(1)
		}

		// 7.1 Verify CRC-32
		bodyCrc32, ok := headers["Content CRC-32"]
		if !ok {
			fmt.Printf("Error parsing headers: Content CRC-32 not present in header\n")
			os.Exit(1)

		}

		bodyCrc32Uint32, err := util.ParseHexUint32(bodyCrc32)
		if err != nil {
			fmt.Printf("Error parsing headers: %s\n", err)
			os.Exit(1)
		}

		if !util.ValidateCRC32(body, bodyCrc32Uint32) {
			if !ignoreChecksumMismatch {
				fmt.Printf("Validation failure: Content CRC-32 mismatch!\n")
				os.Exit(1)
			} else {
				fmt.Printf("Warning: Content CRC-32 mismatch\n")
			}
		}

		// 7.2 Verify CRC-24
		bodyCrc24, ok := headers["Content CRC-24"]
		if !ok {
			fmt.Printf("Error parsing headers: Content CRC-24 not present in header\n")
			os.Exit(1)

		}

		bodyCrc24Uint32, err := util.ParseHexUint32(bodyCrc24)
		if err != nil {
			fmt.Printf("Error parsing headers: %s\n", err)
			os.Exit(1)
		}

		if !util.ValidateCRC24(body, bodyCrc24Uint32) {
			if !ignoreChecksumMismatch {
				fmt.Printf("Validation failure: Content CRC-24 mismatch\n")
				os.Exit(1)
			} else {
				fmt.Printf("Warning: Content CRC-24 mismatch\n")
			}
		}

		// 7.3 Verify SHA-256
		bodySha256, ok := headers["Content SHA-256"]
		if !ok {
			fmt.Printf("Error parsing headers: Content SHA-256 not present in header\n")
			os.Exit(1)
		}

		bodySha256Bytes, err := util.BytesFromBase64(bodySha256)
		if err != nil {
			fmt.Printf("Error parsing headers: %s\n", err)
			os.Exit(1)
		}

		actualSha256 := sha256.Sum256(body)
		if !bytes.Equal(actualSha256[:], bodySha256Bytes) {
			if !ignoreChecksumMismatch {
				fmt.Printf("Error parsing headers: Content SHA-256 mismatch\n")
				os.Exit(1)
			} else {
				fmt.Printf("Warning: Content SHA-256 mismatch\n")
			}
		}

		// 8. Construct PaperCrypt object
		headerDate, ok := headers["Date"]
		if !ok {
			fmt.Printf("Warning: Date not present in header\n")
		}

		timestamp, err := time.Parse("Mon, 02 Jan 2006 15:04:05.000000000 MST", headerDate)
		if err != nil {
			fmt.Printf("Error parsing headers: invalid date format: %s\n", err)
			os.Exit(1)
		}

		paperCrypt := util.NewPaperCrypt(
			versionLine,
			pgpMessage,
			headers["Content Serial"],
			headers["Purpose"],
			headers["Comment"],
			timestamp,
		)

		// 9. Print PaperCrypt object
		fmt.Printf("Understood the following PaperCrypt document:\n")

		jsonBytes, err := json.MarshalIndent(paperCrypt, "", "  ")
		if err != nil {
			fmt.Printf("Error encoding JSON: %s\n", err)
			os.Exit(1)
		}
		fmt.Printf("%s\n", jsonBytes)

		// 10. Read passphrase from stdin
		if !cmd.Flags().Lookup("passphrase").Changed {
			reader := bufio.NewReader(os.Stdin)
			fmt.Println("Enter your decryption passphrase (the passphrase you used to encrypt the data): ")
			passphrase, err = reader.ReadString('\n')
			if err != nil && err != io.EOF {
				fmt.Printf("Error reading passphrase: %s\n", err)
				os.Exit(1)
			}

			passphrase = strings.ReplaceAll(passphrase, "\r", "")
			passphrase = strings.ReplaceAll(passphrase, "\n", "")
		}

		// 11. Decrypt secretContents
		decryptedContents, err := crypto.DecryptMessageWithPassword(pgpMessage, []byte(passphrase))
		if err != nil {
			fmt.Printf("Error decrypting secret contents: %s\n", err)
			os.Exit(1)
		}

		passphrase = "" // clear passphrase

		// 12. Write decryptedContents to outFile
		var outFile *os.File

		if outFileName == "" || outFileName == "-" {
			outFile = os.Stdout
		} else {
			outFile, err = os.OpenFile(outFileName, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
			if err != nil {
				fmt.Printf("Error opening file: %s\n", err)
				os.Exit(1)
			}
			defer outFile.Close()
		}

		n, err := outFile.Write(decryptedContents.GetBinary())
		if err != nil {
			fmt.Printf("Error writing to file: %s\n", err)
			os.Exit(1)
		}

		fmt.Printf("Wrote %d bytes to %s\n", n, outFileName)
	},
}

func init() {
	rootCmd.AddCommand(decodeCmd)

	decodeCmd.Flags().StringVarP(&inFileName, "in", "i", "", "Input file to read from, or stdin if not provided")
	decodeCmd.MarkFlagRequired("in")
	decodeCmd.Flags().StringVarP(&outFileName, "out", "o", "", "Output file to write to, or stdout if not provided")
	decodeCmd.MarkFlagRequired("out")

	decodeCmd.Flags().BoolVarP(&overrideOutFile, "force", "f", false, "Force override of existing file")
	decodeCmd.Flags().BoolVar(&ignoreVersionMismatch, "ignore-version-mismatch", false, "Ignore version mismatch and continue anyway")
	decodeCmd.Flags().BoolVar(&ignoreChecksumMismatch, "ignore-header-checksum-mismatch", false, "Ignore header checksum mismatches and continue anyway")

	decodeCmd.Flags().StringVarP(&passphrase, "passphrase", "P", "", "Passphrase to use for encryption (not recommended, will be prompted for if not provided)")
}
