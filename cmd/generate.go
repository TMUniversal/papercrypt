package cmd

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/pkg/errors"
	"os"
	"papercrypt/util"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var inFileName string
var outFileName string
var overrideOutFile bool

var serialNumber string
var purpose string
var comment string
var date string
var outputPdf bool
var noQR bool

// generateCmd represents the generate command
var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
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

		// 1. Read passphrase from stdin
		reader := bufio.NewReader(os.Stdin)
		fmt.Println("Enter your encryption passphrase (i.e. the key phrase from `papercrypt generateKey`): ")
		passphrase, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("Error reading passphrase: %s\n", err)
			os.Exit(1)
		}

		fmt.Println("Enter your encryption passphrase again: ")
		passphraseAgain, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("Error reading passphrase: %s\n", err)
			os.Exit(1)
		}
		if passphrase != passphraseAgain {
			fmt.Printf("Passphrases do not match\n")
			os.Exit(1)
		}
		passphraseAgain = "" // clear passphraseAgain

		passphrase = strings.ReplaceAll(passphrase, "\r", "")
		passphrase = strings.ReplaceAll(passphrase, "\n", "")

		// 2. Read inFile as JSON, minimize
		secretContentsFile, err := os.OpenFile(inFileName, os.O_RDONLY, 0)
		if err != nil {
			fmt.Printf("Error opening file: %s\n", err)
			os.Exit(1)
		}

		jsonDecoder := json.NewDecoder(secretContentsFile)
		var secretContents map[string]interface{}
		err = jsonDecoder.Decode(&secretContents)
		if err != nil {
			fmt.Printf("Error decoding JSON: %s\n", err)
			os.Exit(1)
		}

		var secretContentsMinimal bytes.Buffer
		err = json.NewEncoder(&secretContentsMinimal).Encode(secretContents)
		if err != nil {
			fmt.Printf("Error encoding JSON: %s\n", err)
			os.Exit(1)
		}

		// 3. Encrypt secretContentsMinimal with passphrase
		encryptedSecretContents, err := encrypt([]byte(passphrase), secretContentsMinimal.Bytes())

		// 4. Write encryptedSecretContents to outFile
		var outFile *os.File

		if outFileName == "" || outFileName == "-" {
			outFile = os.Stdout
		} else {

			outFile, err = os.OpenFile(outFileName, os.O_CREATE|os.O_WRONLY, 0600)
			if err != nil {
				fmt.Printf("Error opening file: %s\n", err)
				os.Exit(1)
			}
			defer outFile.Close()
		}

		if serialNumber == "" {
			serialNumber, err = util.GenerateSerial(6)
			if err != nil {
				fmt.Printf("Error generating serial number: %s\n", err)
				os.Exit(1)
			}
		}

		if date == "" {
			date = time.Now().Format("Mon, 02 Jan 2006 15:04:05.000000000")
		}

		timestamp, err := time.Parse("Mon, 02 Jan 2006 15:04:05.000000000", date)
		if err != nil {
			timestamp, err = time.Parse("2006-01-02 15:04:05", date)
			if err != nil {
				timestamp, err = time.Parse("2006-01-02", date)
				if err != nil {
					fmt.Printf("Error parsing date: %s\n", err)
					os.Exit(1)
				}
			}
		}

		crypt := util.NewPaperCrypt(VersionInfo.Version, encryptedSecretContents, serialNumber, purpose, comment, timestamp)

		var text []byte

		if outputPdf {
			text, err = crypt.GetBinary(noQR)
		} else {
			text, err = crypt.GetText(true)
		}

		if err != nil {
			fmt.Printf("Error creating file contents: %s\n", err)
			os.Exit(1)
		}

		n, err := outFile.Write(text)
		if err != nil {
			fmt.Printf("Error writing file: %s\n", err)
			os.Exit(1)
		}

		fmt.Printf("Wrote %d bytes to %s\n", n, outFile.Name())

		fmt.Println("Done!")
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

	generateCmd.Flags().StringVarP(&inFileName, "in-file", "i", "", "Input JSON file (Required)")
	generateCmd.MarkFlagRequired("in-file")

	generateCmd.Flags().StringVarP(&outFileName, "out-file", "o", "", "Output PDF file (Required)")
	generateCmd.MarkFlagRequired("out-file")

	generateCmd.Flags().BoolVarP(&overrideOutFile, "force", "f", false, "Override output file if it exists (defaults to false)")
	generateCmd.Flags().StringVarP(&serialNumber, "serial-number", "s", "", "Serial number of the sheet (optional, default: 6 random characters)")
	generateCmd.Flags().StringVarP(&purpose, "purpose", "p", "", "Purpose of the sheet (optional)")
	generateCmd.Flags().StringVarP(&comment, "comment", "c", "", "Comment on the sheet (optional)")
	generateCmd.Flags().StringVarP(&date, "date", "d", "", "Date of the sheet (optional, defaults to now)")
	generateCmd.Flags().BoolVar(&noQR, "no-qr", false, "Do not generate QR code (optional)")
	generateCmd.Flags().BoolVar(&outputPdf, "pdf", false, "Whether to output a PDF (optional, defaults to false, currently not implemented)")
}
