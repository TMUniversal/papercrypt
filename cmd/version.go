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

	"github.com/spf13/cobra"
	"github.com/tmuniversal/papercrypt/v2/internal"
)

var versionCmd = &cobra.Command{
	Aliases:      []string{"v"},
	SilenceUsage: true,
	Args:         cobra.NoArgs,
	Use:          "version",
	Short:        "Shows the version and build information of the application",
	PersistentPostRun: func(_ *cobra.Command, _ []string) {
		// Present to override the default behavior of the root command
	},
	Run: func(_ *cobra.Command, _ []string) {
		fmt.Println(internal.VersionInfo.String())
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
