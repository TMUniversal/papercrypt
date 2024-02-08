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
	crand "crypto/rand"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"math/big"
	"math/rand"
	"os"
	"strings"

	"github.com/caarlos0/log"
	"github.com/spf13/cobra"
	"github.com/tmuniversal/papercrypt/internal"
)

const (
	passphraseSheetWordCount = 135
)

// phraseSheetCmd represents the phraseSheet command
var phraseSheetCmd = &cobra.Command{
	Aliases:      []string{"ps", "p"},
	Args:         cobra.MaximumNArgs(1),
	SilenceUsage: true,
	Use:          "phraseSheet [base64 seed]",
	Short:        "Generate a passphrase sheet.",
	Example:      "papercrypt phraseSheet -o phraseSheet.pdf",
	RunE: func(cmd *cobra.Command, args []string) error {
		// 1. Open output file
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

		if len(wordList) == 0 {
			generateWordList()
		}

		// 2. Generate seed (if not provided)
		var seed int64
		if len(args) == 0 {
			random, err := crand.Int(crand.Reader, big.NewInt(1<<63-1))
			if err != nil {
				return errors.Join(errors.New("error generating random seed"), err)
			}
			seed = random.Int64()
		} else {
			seedBytes, err := base64.StdEncoding.DecodeString(strings.TrimSpace(args[0]))
			if err != nil {
				return errors.Join(errors.New("error decoding seed"), err)
			}
			seed = int64(binary.BigEndian.Uint64(seedBytes))
			if err != nil {
				return errors.Join(errors.New("error converting seed to int64"), err)
			}
		}

		// 3. Get words
		words, err := GenerateFromSeed(seed, passphraseSheetWordCount)
		if err != nil {
			return errors.Join(errors.New("error generating words"), err)
		}

		// 4. Generate PDF
		data, err := internal.GeneratePassphraseSheetPDF(seed, words)
		if err != nil {
			return errors.Join(errors.New("error generating PDF"), err)
		}

		// 5. Write PDF
		n, err := outFile.Write(data)
		if err != nil {
			return errors.Join(errors.New("error writing PDF"), err)
		}

		internal.PrintWrittenSize(n, outFile)
		return nil
	},
}

func GenerateFromSeed(seed int64, amount int) ([]string, error) {
	if amount < 1 {
		return nil, errors.New("amount must be greater than 0")
	}
	// 2. Generate random numbers
	gen := rand.New(rand.NewSource(seed))

	words := make([]string, amount)
	for i := 0; i < amount; i++ {
		random := gen.Intn(len(wordList)) // Intn returns [0, n) (excludes n)
		w := wordList[random]

		if internal.SliceHasString(words, w) {
			// if the word is already in the slice, try again
			log.WithField("word", w).WithField("index", i).Warn("Duplicate word appeared, trying again...")
			i--
			continue
		}

		words[i] = w
	}
	return words, nil
}

func init() {
	rootCmd.AddCommand(phraseSheetCmd)
}
