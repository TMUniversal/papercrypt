package cmd

import (
	"bufio"
	"io"
	"os"
	"strings"
	"time"

	"github.com/tmuniversal/papercrypt/internal"

	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var serialNumber string
var purpose string
var comment string
var date string

var outputPdf bool
var noQR bool
var lowerCasedBase16 bool
var asciiArmor bool

var passphrase string

// generateCmd represents the generate command
var generateCmd = &cobra.Command{
	Aliases: []string{"gen", "g"},
	Use:     "generate",
	Short:   "Generate a PaperCrypt document",
	Long: `The 'generate' command takes a JSON file as input and encrypts the data within. It then embeds the encrypted data in a 
newly created PDF file that you can print for physical storage.

Please note, to decrypt the data from the output PaperCrypt PDF, you'll need the original passphrase used during the 
encryption process. Treat this passphrase with care; loss of the passphrase could result in the permanent loss of the 
encrypted data.`,
	Example: "papercrypt generate -i <file>.json -o <file>.pdf --purpose \"My secret data\" --comment \"This is a comment\" --date \"2021-01-01 12:00:00\"",
	Run: func(cmd *cobra.Command, args []string) {
		// 1. Open output file
		outFile, err := internal.GetFileHandleCarefully(outFileName, overrideOutFile)
		if err != nil {
			cmd.Println("Error opening output file:", err)
			os.Exit(1)
		}
		defer outFile.Close()

		// 2. generate serial number if not provided
		if serialNumber == "" {
			var err error
			serialNumber, err = internal.GenerateSerial(6)
			if err != nil {
				cmd.Printf("Error generating serial number: %s\n", err)
				os.Exit(1)
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
						cmd.Printf("Error parsing date: %s\n", err)
						os.Exit(1)
					}
				}
			}
		}

		// 4. Read input file as bytes
		secretContentsFile, err := os.ReadFile(inFileName)
		if err != nil {
			cmd.Printf("Error reading file: %s\n", err)
			os.Exit(1)
		}

		// 5. Read passphrase from stdin
		if !cmd.Flags().Lookup("passphrase").Changed {
			reader := bufio.NewReader(os.Stdin)
			cmd.Println("Enter your encryption passphrase (i.e. the key phrase from `papercrypt generateKey`): ")
			passphrase, err = reader.ReadString('\n')
			if err != nil && err != io.EOF {
				cmd.Printf("Error reading passphrase: %s\n", err)
				os.Exit(1)
			}

			cmd.Println("Enter your encryption passphrase again: ")
			passphraseAgain, err := reader.ReadString('\n')
			if err != nil && err != io.EOF {
				cmd.Printf("Error reading passphrase: %s\n", err)
				os.Exit(1)
			}
			if passphrase != passphraseAgain {
				cmd.Printf("Passphrases do not match! Aborting.\n")
				os.Exit(1)
			}
			passphraseAgain = "" // clear passphraseAgain

			passphrase = strings.ReplaceAll(passphrase, "\r", "")
			passphrase = strings.ReplaceAll(passphrase, "\n", "")
		}

		// 6. Encrypt secretContentsMinimal with passphrase
		encryptedSecretContents, err := encrypt([]byte(passphrase), secretContentsFile)
		if err != nil {
			cmd.Printf("Error encrypting secret contents: %s\n", err)
			os.Exit(1)
		}

		passphrase = "" // clear passphrase

		// 7. Write encryptedSecretContents to outFile
		crypt := internal.NewPaperCrypt(VersionInfo.Version, encryptedSecretContents, serialNumber, purpose, comment, timestamp)

		var text []byte

		if outputPdf {
			text, err = crypt.GetPDF(asciiArmor, noQR, lowerCasedBase16)
		} else {
			text, err = crypt.GetText(asciiArmor, lowerCasedBase16)
		}

		if err != nil {
			cmd.Printf("Error creating file contents: %s\n", err)
			os.Exit(1)
		}

		n, err := outFile.Write(text)
		if err != nil {
			cmd.Printf("Error writing file: %s\n", err)
			os.Exit(1)
		}

		cmd.Printf("Wrote %d bytes to %s\n", n, outFile.Name())

		cmd.Println("Done!")
	},
}

func encrypt(passphrase []byte, data []byte) (*crypto.PGPMessage, error) {
	var message = crypto.NewPlainMessage(data)

	encrypted, err := crypto.EncryptMessageWithPassword(message, passphrase)
	if err != nil {
		return nil, errors.Errorf("error encrypting message: %s", err)
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
	generateCmd.Flags().BoolVar(&outputPdf, "pdf", true, "Whether to output a PDF (optional, defaults to true)")
	generateCmd.Flags().BoolVar(&lowerCasedBase16, "lowercase", false, "Whether to use lower case letters for hexadecimal digits (optional, defaults to false)")
	generateCmd.Flags().BoolVar(&asciiArmor, "armor", false, "Whether to use ASCII armor instead of hex+crc serialization (optional, defaults to hex)")

	generateCmd.Flags().StringVarP(&passphrase, "passphrase", "P", "", "Passphrase to use for encryption (not recommended, will be prompted for if not provided)")
}
