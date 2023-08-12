package cmd

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/pkg/errors"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var inFile string
var outFile string

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

		fmt.Printf("Encrypting with passphrase [%s]\n", passphrase)

		// 2. Read inFile as JSON, minimize
		secretContentsFile, err := os.OpenFile(inFile, os.O_RDONLY, 0)
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
		encryptedSecretContentsFile, err := os.OpenFile(outFile, os.O_CREATE|os.O_WRONLY, 0600)
		if err != nil {
			fmt.Printf("Error opening file: %s\n", err)
			os.Exit(1)
		}

		n, err := encryptedSecretContentsFile.Write(encryptedSecretContents.GetBinary())
		if err != nil {
			fmt.Printf("Error writing file: %s\n", err)
			os.Exit(1)
		}

		fmt.Printf("Wrote %d bytes to %s\n", n, outFile)
		armored, err := encryptedSecretContents.GetArmored()
		if err != nil {
			fmt.Printf("Error getting armored contents: %s\n", err)
			os.Exit(1)
		}
		fmt.Printf("Encrypted Contents:\n%s\n", armored)

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

	generateCmd.Flags().StringVarP(&inFile, "in-file", "i", "", "Input JSON file (Required)")
	generateCmd.MarkFlagRequired("in-file")

	generateCmd.Flags().StringVarP(&outFile, "out-file", "o", "", "Output PDF file (Required)")
	generateCmd.MarkFlagRequired("out-file")
}
