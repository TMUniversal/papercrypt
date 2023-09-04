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
	"fmt"

	"github.com/spf13/cobra"
)

type VersionDetails struct {
	Version   string
	BuildDate string
	GitCommit string
	GitRef    string
	GoVersion string
	OsArch    string
	OsType    string
}

var VersionInfo VersionDetails

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Shows the version of the application",
	Long:  `Shows the version of the application, as well as the build date, git commit hash, git ref, Go version, os/arch and os type.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("PaperCrypt Version %s,\nbuilt on %s, from commit %s,\nfor %s/%s (Go %s)\n", VersionInfo.Version, VersionInfo.BuildDate, VersionInfo.GitCommit, VersionInfo.OsType, VersionInfo.OsArch, VersionInfo.GoVersion)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
