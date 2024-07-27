/*
 * This file is part of PaperCrypt.
 *
 * PaperCrypt lets you prepare encrypted messages for printing on paper.
 * Copyright (C) 2024 TMUniversal <me@tmuniversal.eu>.
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

package internal

import (
	"errors"
	"math/rand"

	"github.com/caarlos0/log"
)

func GenerateFromSeed(seed int64, amount int, wordList *[]string) ([]string, error) {
	if amount < 1 {
		return nil, errors.New("amount must be greater than 0")
	}
	// 2. Generate random numbers
	gen := rand.New(rand.NewSource(seed))

	words := make([]string, amount)
	for i := 0; i < amount; i++ {
		random := gen.Intn(len(*wordList)) // Intn returns [0, n) (excludes n)
		w := (*wordList)[random]

		if SliceHasString(words, w) {
			// if the word is already in the slice, try again
			log.WithField("word", w).WithField("index", i).Warn("Duplicate word appeared, trying again...")
			i--
			continue
		}

		words[i] = w
	}
	return words, nil
}
