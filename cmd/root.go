package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var inFileName string
var outFileName string
var overrideOutFile bool

var LicenseText *string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "papercrypt",
	Short: "PaperCrypt lets you prepare encrypted messages for printing on paper",
	Long: `PaperCrypt lets you prepare encrypted messages for printing on paper.

It is designed to let you enter any JSON data, encrypt it with a passphrase,
and then prepare a printable document that is optimized for being able to restore the data.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Println("PaperCrypt  Copyright (C) 2023  TMUniversal <me@tmuniversal.eu>")
		cmd.Println("This program comes with ABSOLUTELY NO WARRANTY; for details type `papercrypt show w'.")
		cmd.Println("This is free software, and you are welcome to redistribute it")
		cmd.Println("under certain conditions; type `papercrypt show c' for details.")
		cmd.Println("PaperCrypt's source code can be found at https://github.com/TMUniversal/PaperCrypt\n")
		cmd.Help()
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
}
