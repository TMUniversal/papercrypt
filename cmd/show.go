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

	"github.com/caarlos0/log"
	"github.com/spf13/cobra"
)

var (
	LicenseText    *string
	ThirdPartyText *string
)

// urlCmd represents the url command
var showCmd = &cobra.Command{
	Aliases:      []string{"s"},
	Args:         cobra.NoArgs,
	SilenceUsage: true,
	Use:          "show",
	Short:        "Show commands: 'w', 'c', 't'",
	Long:         `Use 'show [w/c/t]' to view warranty or copyright info`,
}

var showCmdWarranty = &cobra.Command{
	Use:          "warranty",
	Aliases:      []string{"w"},
	SilenceUsage: true,
	Short:        "Show warranty info",
	Run: func(_ *cobra.Command, _ []string) {
		log.Info("This program is licensed under the terms of the GNU AGPL-3.0-or-later license.")
		log.Info("An excerpt from the license will be printed below, to view the full license, please run `papercrypt show c'.\n")
		fmt.Println(`  15. Disclaimer of Warranty.

  THERE IS NO WARRANTY FOR THE PROGRAM, TO THE EXTENT PERMITTED BY
APPLICABLE LAW.  EXCEPT WHEN OTHERWISE STATED IN WRITING THE COPYRIGHT
HOLDERS AND/OR OTHER PARTIES PROVIDE THE PROGRAM "AS IS" WITHOUT WARRANTY
OF ANY KIND, EITHER EXPRESSED OR IMPLIED, INCLUDING, BUT NOT LIMITED TO,
THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR
PURPOSE.  THE ENTIRE RISK AS TO THE QUALITY AND PERFORMANCE OF THE PROGRAM
IS WITH YOU.  SHOULD THE PROGRAM PROVE DEFECTIVE, YOU ASSUME THE COST OF
ALL NECESSARY SERVICING, REPAIR OR CORRECTION.

  16. Limitation of Liability.

  IN NO EVENT UNLESS REQUIRED BY APPLICABLE LAW OR AGREED TO IN WRITING
WILL ANY COPYRIGHT HOLDER, OR ANY OTHER PARTY WHO MODIFIES AND/OR CONVEYS
THE PROGRAM AS PERMITTED ABOVE, BE LIABLE TO YOU FOR DAMAGES, INCLUDING ANY
GENERAL, SPECIAL, INCIDENTAL OR CONSEQUENTIAL DAMAGES ARISING OUT OF THE
USE OR INABILITY TO USE THE PROGRAM (INCLUDING BUT NOT LIMITED TO LOSS OF
DATA OR DATA BEING RENDERED INACCURATE OR LOSSES SUSTAINED BY YOU OR THIRD
PARTIES OR A FAILURE OF THE PROGRAM TO OPERATE WITH ANY OTHER PROGRAMS),
EVEN IF SUCH HOLDER OR OTHER PARTY HAS BEEN ADVISED OF THE POSSIBILITY OF
SUCH DAMAGES.

  17. Interpretation of Sections 15 and 16.

  If the disclaimer of warranty and limitation of liability provided
above cannot be given local legal effect according to their terms,
reviewing courts shall apply local law that most closely approximates
an absolute waiver of all civil liability in connection with the
Program, unless a warranty or assumption of liability accompanies a
copy of the Program in return for a fee.`)
	},
}

var showCmdCopyright = &cobra.Command{
	Use:          "copyright",
	Aliases:      []string{"c", "license", "l"},
	SilenceUsage: true,
	Short:        "Show copyright info",
	Run: func(_ *cobra.Command, _ []string) {
		log.Info("This program is licensed under the terms of the GNU AGPL-3.0-or-later license.")
		fmt.Println(*LicenseText)
	},
}

var showCmdThirdParty = &cobra.Command{
	Use:          "third-party",
	Aliases:      []string{"t"},
	SilenceUsage: true,
	Short:        "Show third party licenses",
	Run: func(_ *cobra.Command, _ []string) {
		fmt.Println(*ThirdPartyText)
	},
}

func init() {
	showCmd.AddCommand(showCmdWarranty, showCmdCopyright, showCmdThirdParty)

	rootCmd.AddCommand(showCmd)
}
