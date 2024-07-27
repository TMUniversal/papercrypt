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
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"os"
	"strings"

	"github.com/caarlos0/log"
	"github.com/spf13/cobra"
	"github.com/tmuniversal/papercrypt/internal"
)

var words int

var (
	WordListFile *string
	wordList     = make([]string, 0)
)

const wordListURL = "https://www.eff.org/files/2016/07/18/eff_large_wordlist.txt"

var wordListURLFormatted = internal.URL(wordListURL)

var generateKeyCmd = &cobra.Command{
	Aliases:      []string{"key", "gen", "k"},
	Args:         cobra.NoArgs,
	SilenceUsage: true,
	Use:          "generate-key",
	Short:        "Generates a mnemonic key phrase",
	Long: fmt.Sprintf(`This command generates a mnemonic key phrase base on the eff.org large word list,
which can be found here: %s.`, wordListURLFormatted),
	RunE: func(_ *cobra.Command, _ []string) error {
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

		log.Info("Generating key phrase...")
		keyPhrase, err := generateMnemonic(words)
		if err != nil {
			return errors.Join(errors.New("error generating key phrase"), err)
		}
		log.Info("Key phrase generated.")

		wordString := strings.Join(keyPhrase, " ")
		if outFile == os.Stdout {
			wordString = internal.Bold(wordString)
		}

		n, err := outFile.WriteString(wordString)
		if err != nil {
			return errors.Join(errors.New("error writing key phrase"), err)
		}

		if outFile == os.Stdout {
			fmt.Fprintln(outFile)
		}

		internal.PrintWrittenSize(n, outFile)
		return nil
	},
}

func generateWordList() {
	wordListArray := strings.Split(*WordListFile, "\n")

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

func generateMnemonic(amount int) ([]string, error) {
	if len(wordList) == 0 {
		generateWordList()
	}

	// choose `amount` random words from wordListArray
	randInt, err := rand.Int(rand.Reader, big.NewInt(int64(len(wordList))))
	if err != nil {
		return nil, errors.Join(errors.New("error generating random seed"), err)
	}

	return internal.GenerateFromSeed(randInt.Int64(), amount, &wordList)
}

func init() {
	rootCmd.AddCommand(generateKeyCmd)

	generateKeyCmd.Flags().IntVarP(&words, "words", "w", 24, "Number of words to include in the key phrase")
}
