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
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"

	"github.com/caarlos0/log"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/tmuniversal/papercrypt/internal"
)

var words int

var WordListFile *string
var wordList = make([]string, 0)

const wordListUrl = "https://www.eff.org/files/2016/07/18/eff_large_wordlist.txt"

var wordListUrlFormatted = internal.URL(wordListUrl)

var generateKeyCmd = &cobra.Command{
	Aliases: []string{"key", "gen", "k"},
	Args:    cobra.NoArgs,
	Use:     "generateKey",
	Short:   "Generates a mnemonic key phrase",
	Long: fmt.Sprintf(`This command generates a mnemonic key phrase base on the eff.org large word list,
which can be found here: %s.`, wordListUrlFormatted),
	RunE: func(cmd *cobra.Command, args []string) error {
		out, err := internal.GetFileHandleCarefully(outFileName, overrideOutFile)
		if err != nil {
			return err
		}
		defer out.Close()

		log.Info("Generating key phrase...")
		keyPhrase, err := generateMnemonic(words)
		if err != nil {
			return errors.Wrap(err, "error generating key phrase")
		}
		log.Info("Key phrase generated.")

		n, err := out.WriteString(strings.Join(keyPhrase, " "))
		if err != nil {
			return errors.Wrap(err, "error writing key phrase")
		}

		internal.PrintWrittenSize(n, out)

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
		return nil, errors.Wrap(err, "Error generating random seed")
	}

	return GenerateFromSeed(randInt.Int64(), amount)
}

func init() {
	rootCmd.AddCommand(generateKeyCmd)

	generateKeyCmd.Flags().IntVarP(&words, "words", "w", 24, "Number of words to include in the key phrase")
}
