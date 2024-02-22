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
	"os"

	"github.com/caarlos0/log"
	"github.com/spf13/cobra"
	"github.com/tmuniversal/papercrypt/internal"
)

var (
	inFileName      string
	outFileName     string
	overrideOutFile bool
)

var verbosity int

const repo = "https://github.com/TMUniversal/papercrypt"

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:          "papercrypt",
	SilenceUsage: true,
	Short:        "PaperCrypt lets you prepare encrypted messages for printing on paper",
	Long: `PaperCrypt lets you prepare encrypted messages for printing on paper.

It is designed to let you enter data, encrypt it with a passphrase,
and then prepare a printable document that is optimized for being able to restore the data.`,
	PersistentPreRun: func(_ *cobra.Command, _ []string) {
		level := max(log.InfoLevel-log.Level(verbosity), log.DebugLevel)
		log.SetLevel(level)
		log.Debug("verbosity set to " + level.String())
	},
	PersistentPostRun: func(_ *cobra.Command, _ []string) {
		log.Info("thank you for using papercrypt!")
	},
	RunE: func(cmd *cobra.Command, _ []string) error {
		cmd.Println("PaperCrypt  Copyright (C) 2023  TMUniversal <me@tmuniversal.eu>")
		cmd.Println("This program comes with ABSOLUTELY NO WARRANTY; for details type `papercrypt show w'.")
		cmd.Println("This is free software, and you are welcome to redistribute it")
		cmd.Println("under certain conditions; type `papercrypt show c' for details.")
		cmd.Print("PaperCrypt's source code can be found at ")
		cmd.Print(internal.URL(repo))
		cmd.Println(".")
		cmd.Println()
		return cmd.Help()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&inFileName, "in", "i", "", "Input file to read from, or stdin if not provided")
	rootCmd.PersistentFlags().StringVarP(&outFileName, "out", "o", "", "Output file to write to, or stdout if not provided")
	rootCmd.PersistentFlags().BoolVarP(&overrideOutFile, "force", "f", false, "Force override of existing file")
	rootCmd.PersistentFlags().CountVarP(&verbosity, "verbose", "v", "Increase verbosity level")
}
