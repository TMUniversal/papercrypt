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
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
	"github.com/tmuniversal/papercrypt/internal"
)

// manCmd represents the man command
var manCmd = &cobra.Command{
	Aliases: []string{"man", "m"},
	Use:     "manual",
	Short:   "Generate man page",
	Long: `Generate man pages for PaperCrypt commands.

Generated pages will be placed in the directory specified by --out, defaulting to './man'.`,
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		// 1. outFileName is the directory here, if it is unset, default to man
		if outFileName == "" {
			outFileName = "man"
		}

		// 2. Generate files
		err := GenerateManPage(outFileName)
		if err != nil {
			cmd.Print(errors.Wrap(err, "Error generating man page"))
			os.Exit(1)
		}
	},
}

func GenerateManPage(dir string) error {
	header := &doc.GenManHeader{
		Source: "PaperCrypt " + internal.VersionInfo.Version,
	}

	return doc.GenManTree(rootCmd, header, dir)
}

func init() {
	rootCmd.AddCommand(manCmd)
}
