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
	"fmt"
	"os"

	mcobra "github.com/muesli/mango-cobra"
	"github.com/muesli/roff"
	"github.com/spf13/cobra"
)

// manCmd represents the man command.
var manCmd = &cobra.Command{
	Aliases:               []string{"man", "m"},
	Args:                  cobra.NoArgs,
	Short:                 "Generate man pages",
	SilenceUsage:          true,
	DisableFlagsInUseLine: true,
	Hidden:                true,
	RunE: func(_ *cobra.Command, _ []string) error {
		manPage, err := mcobra.NewManPage(1, rootCmd.Root())
		if err != nil {
			return err
		}

		_, err = fmt.Fprint(os.Stdout, manPage.Build(roff.NewDocument()))

		return err
	},
}

func init() {
	rootCmd.AddCommand(manCmd)
}
