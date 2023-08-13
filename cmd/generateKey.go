package cmd

import (
	"crypto/rand"
	_ "embed"
	"fmt"
	"math/big"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var genKeyOutFile string
var genKeyOutFileOverride bool
var words int

//go:embed "eff.org_files_2016_07_18_eff_large_wordlist.txt"
var wordListFile string
var wordList = make([]string, 0)

var generateKeyCmd = &cobra.Command{
	Use:   "generateKey",
	Short: "Generates a mnemonic key phrase",
	Long:  `This command generates a mnemonic key phrase base on the eff.org large word list.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Generating key phrase...")
		keyPhrase, err := generateMnemonic(words)
		if err != nil {
			fmt.Println("An error occurred:", err)
			os.Exit(1)
		}
		fmt.Println("Key phrase generated.")

		var out *os.File

		if genKeyOutFile == "" || genKeyOutFile == "-" {
			out = os.Stdout
		} else {
			if _, err := os.Stat(genKeyOutFile); err == nil {
				if !genKeyOutFileOverride {
					fmt.Printf("File %s already exists. Use -f to override.\n", genKeyOutFile)
					os.Exit(1)
				}
			}

			out, err = os.OpenFile(genKeyOutFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
			if err != nil {
				fmt.Println("An error occurred:", err)
				os.Exit(1)
			}
			defer out.Close()
		}

		n, err := out.WriteString(strings.Join(keyPhrase, " "))
		if err != nil {
			fmt.Println("Error writing file:", err)
			os.Exit(1)
		}

		if out != os.Stdout {
			fmt.Printf("Wrote %d bytes to %s\n", n, out.Name())
		}
	},
}

func generateMnemonic(amount int) ([]string, error) {
	if len(wordList) == 0 {
		wordListArray := strings.Split(wordListFile, "\n")

		for i, word := range wordListArray {
			wordListArray[i] = strings.TrimSpace(strings.Split(word, "\t")[1])
		}

		for _, word := range wordListArray {
			if strings.TrimSpace(word) == "" {
				continue
			}

			wordList = append(wordList, word)
		}
	}

	// choose `amount` random words from wordListArray

	words := make([]string, amount)
	for i := 0; i < amount; i++ {
		randInt, err := rand.Int(rand.Reader, big.NewInt(int64(len(wordList))))
		if err != nil {
			return nil, err
		}

		words[i] = wordList[randInt.Int64()]
	}

	return words, nil
}

func init() {
	rootCmd.AddCommand(generateKeyCmd)

	generateKeyCmd.Flags().IntVarP(&words, "words", "w", 24, "Number of words to include in the key phrase (defaults to 24)")
	generateKeyCmd.Flags().StringVarP(&genKeyOutFile, "out", "o", "", "File to write the key phrase to (defaults to stdout)")
	generateKeyCmd.Flags().BoolVarP(&genKeyOutFileOverride, "force", "f", false, "Override the output file if it exists")
}
