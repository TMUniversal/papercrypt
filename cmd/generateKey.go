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
	"math/big"
	"os"
	"strings"

	"github.com/pkg/errors"
	"github.com/tmuniversal/papercrypt/internal"

	"github.com/spf13/cobra"
)

var words int

var WordListFile *string
var wordList = make([]string, 0)

var generateKeyCmd = &cobra.Command{
	Aliases: []string{"key", "gen", "k"},
	Use:     "generateKey",
	Short:   "Generates a mnemonic key phrase",
	Long: `This command generates a mnemonic key phrase base on the eff.org large word list,
which can be found here: https://www.eff.org/files/2016/07/18/eff_large_wordlist.txt.`,
	Run: func(cmd *cobra.Command, args []string) {
		out, err := internal.GetFileHandleCarefully(outFileName, overrideOutFile)
		if err != nil {
			cmd.Println("Error opening output file:", err)
			os.Exit(1)
		}
		defer out.Close()

		cmd.Println("Generating key phrase...")
		keyPhrase, err := generateMnemonic(words)
		if err != nil {
			cmd.Println("An error occurred:", err)
			os.Exit(1)
		}
		cmd.Println("Key phrase generated.")

		n, err := out.WriteString(strings.Join(keyPhrase, " "))
		if err != nil {
			cmd.Println("Error writing file:", err)
			os.Exit(1)
		}

		cmd.Printf("Wrote %s bytes to %s\n", internal.SprintBinarySize(n), out.Name())
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

	generateKeyCmd.Flags().IntVarP(&words, "words", "w", 24, "Number of words to include in the key phrase (defaults to 24)")
}
