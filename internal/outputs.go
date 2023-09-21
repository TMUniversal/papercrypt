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

package internal

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func Fatal(cmd *cobra.Command, err error) {
	cmd.Println(err)
	os.Exit(1)
}

func PrintInputSource(cmd *cobra.Command, inFileName string) {
	if inFileName == "" || inFileName == "-" {
		cmd.Println("Reading from stdin")
	} else {
		cmd.Printf("Reading from %s\n", inFileName)
	}
}

func PrintWrittenSize(cmd *cobra.Command, size int, file *os.File) {
	if size == 0 {
		cmd.Println("No data written.")
	} else {
		cmd.Printf("Wrote %s to %s\n", SprintBinarySize(size), file.Name())
	}
}

func SprintBinarySize64(size int64) string {
	if size < 1024 {
		return fmt.Sprintf("%d B", size)
	}
	if size < 1024*1024 {
		return fmt.Sprintf("%.2f KiB", float64(size)/1024)
	}
	if size < 1024*1024*1024 {
		return fmt.Sprintf("%.2f MiB", float64(size)/(1024*1024))
	}
	if size < 1024*1024*1024*1024 {
		return fmt.Sprintf("%.2f GiB", float64(size)/(1024*1024*1024))
	}
	return fmt.Sprintf("%.2f TiB", float64(size)/(1024*1024*1024*1024))
}

func SprintBinarySize(size int) string {
	return SprintBinarySize64(int64(size))
}
