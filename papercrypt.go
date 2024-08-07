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

package main

import (
	_ "embed"
	"os"

	goversion "github.com/caarlos0/go-version"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
	"github.com/tmuniversal/papercrypt/v2/cmd"
	"github.com/tmuniversal/papercrypt/v2/internal"
)

// LicenseText is the license of the application as a string
//
//go:embed COPYING
var LicenseText string

// ThirdPartyLicenses is THIRD_PARTY.md as a string
//
//go:embed THIRD_PARTY.md
var ThirdPartyLicenses string

// WordList is the eff.org large word list as a string
//
//go:embed "eff.org_files_2016_07_18_eff_large_wordlist.txt"
var WordList string

//go:embed "font/Noto_Sans/NotoSans-Regular.ttf"
var pdfFontTextRegular string

//go:embed "font/Noto_Sans/NotoSans-Bold.ttf"
var pdfFontTextBold string

//go:embed "font/Noto_Sans/NotoSans-Italic.ttf"
var pdfFontTextItalic string

//go:embed "font/Inconsolata/static/Inconsolata-Medium.ttf"
var pdfFontMonoRegular string

//go:embed "font/Inconsolata/static/Inconsolata-ExtraBold.ttf"
var pdfFontMonoBold string

//go:embed "font/Inconsolata/Inconsolata-VariableFont_wdth,wght.ttf"
var pdfFontMonoItalic string

var (
	version   = ""
	commit    = ""
	treeState = ""
	date      = ""
	builtBy   = ""
)

func init() {
	// enable colored output in ci
	if os.Getenv("CI") != "" {
		lipgloss.SetColorProfile(termenv.TrueColor)
	}
}

func main() {
	cmd.LicenseText = &LicenseText
	cmd.ThirdPartyText = &ThirdPartyLicenses
	cmd.WordListFile = &WordList
	internal.VersionInfo = buildVersion(version, commit, date, builtBy, treeState)
	internal.PdfTextFontRegularBytes = []byte(pdfFontTextRegular)
	internal.PdfTextFontItalicBytes = []byte(pdfFontTextItalic)
	internal.PdfTextFontBoldBytes = []byte(pdfFontTextBold)
	internal.PdfMonoFontRegularBytes = []byte(pdfFontMonoRegular)
	internal.PdfMonoFontBoldBytes = []byte(pdfFontMonoBold)
	internal.PdfMonoFontItalicBytes = []byte(pdfFontMonoItalic)

	cmd.Execute()
}

const website = "https://github.com/TMUniversal/papercrypt"

//go:embed art.txt
var asciiArt string

func buildVersion(version, commit, date, builtBy, treeState string) goversion.Info {
	return goversion.GetVersionInfo(
		goversion.WithAppDetails("PaperCrypt", "Encrypted backups for printing on paper", website),
		goversion.WithASCIIName(asciiArt),
		func(i *goversion.Info) {
			if commit != "" {
				i.GitCommit = commit
			}
			if treeState != "" {
				i.GitTreeState = treeState
			}
			if date != "" {
				i.BuildDate = date
			}
			if version != "" {
				i.GitVersion = version
			}
			if builtBy != "" {
				i.BuiltBy = builtBy
			}
		},
	)
}
