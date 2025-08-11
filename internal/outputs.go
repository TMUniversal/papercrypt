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

package internal

import (
	"fmt"
	"os"

	"github.com/caarlos0/log"
)

// PrintWrittenSize logs the amount of data written in human-readable notation.
// A warning is issues when the size is 0.
func PrintWrittenSize(size int, file *os.File) {
	if size == 0 {
		log.Warn(Warning(fmt.Sprintf("No data written to %s", file.Name())))
	} else {
		log.WithField("size", size).WithField("path", file.Name()).Debug(fmt.Sprintf("%s written to %s.", sprintBinarySize(size), file.Name()))
	}
}

func sprintBinarySize64(size int64) string {
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

func sprintBinarySize(size int) string {
	return sprintBinarySize64(int64(size))
}
