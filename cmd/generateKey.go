package cmd

import (
	"crypto/rand"
	_ "embed"
	"fmt"
	"github.com/spf13/cobra"
	"math/big"
	"os"
	"strings"
)

var words int

//go:embed "eff.org_files_2016_07_18_eff_large_wordlist.txt"
var wordListFile string
var wordList []string = make([]string, 0)

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
		fmt.Println("Key phrase generated: ", strings.Join(keyPhrase, " "))
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

	generateKeyCmd.Flags().IntVarP(&words, "words", "w", 24, "Number of words to include in the key phrase (Default is 12)")
}
