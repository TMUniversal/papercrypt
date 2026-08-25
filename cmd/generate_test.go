/*
 * This file is part of PaperCrypt.
 *
 * PaperCrypt lets you prepare encrypted messages for printing on paper.
 * Copyright (C) 2023-2026 TMUniversal <me@tmuniversal.eu>.
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
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// resetGenerateVars resets the package-level flag-bound variables used by the
// generate command to their zero values. cobra/pflag does not reset a flag's
// bound variable to its default when a flag is simply omitted from a later
// Execute() call on the same command instance, so tests must reset this
// state themselves to avoid leaking values between test cases.
func resetGenerateVars() {
	serialNumber = ""
	purpose = ""
	comment = ""
	date = ""
	noQR = false
	lowerCasedBase16 = false
	rawData = false
	passphrase = ""
}

// runGenerate resets generate-command state, writes inputData to a temp
// input file, runs the generate command against it (always supplying the
// "example" passphrase plus a purpose/comment), and returns the path to the
// produced output file along with any error returned by command execution.
func runGenerate(t *testing.T, inputData []byte, extraArgs ...string) (string, error) {
	t.Helper()
	resetGenerateVars()

	tempDir := t.TempDir()
	inPath := filepath.Join(tempDir, "input.json")
	outPath := filepath.Join(tempDir, "output.pdf")

	require.NoError(t, os.WriteFile(inPath, inputData, 0o600))

	args := append([]string{
		"generate",
		"-i", inPath,
		"-o", outPath,
		"-P", "example",
		"--purpose", "Test Purpose",
		"--comment", "Test Comment",
	}, extraArgs...)
	rootCmd.SetArgs(args)

	err := rootCmd.Execute()
	return outPath, err
}

// TestGenerateFlags_NoCodeAndNoQRAlias asserts that "--no-code" is the
// visible, primary flag, that "--no-qr" is retained only as a hidden alias,
// and that both are bound to the same underlying noQR variable.
func TestGenerateFlags_NoCodeAndNoQRAlias(t *testing.T) {
	noCodeFlag := generateCmd.Flags().Lookup("no-code")
	require.NotNil(t, noCodeFlag)
	assert.False(t, noCodeFlag.Hidden, "--no-code should be a visible flag")

	noQRFlag := generateCmd.Flags().Lookup("no-qr")
	require.NotNil(t, noQRFlag)
	assert.True(t, noQRFlag.Hidden, "--no-qr should be hidden")

	noQR = false
	require.NoError(t, generateCmd.Flags().Set("no-qr", "true"))
	assert.True(t, noQR, "--no-qr should set the shared noQR variable")

	noQR = false
	require.NoError(t, generateCmd.Flags().Set("no-code", "true"))
	assert.True(t, noQR, "--no-code should set the shared noQR variable")

	// reset shared state so later tests are unaffected by this direct flag mutation
	noQR = false
}

// TestGenerateCommand_PGPMode asserts that the default (PGP) mode of the
// generate command, which now encrypts before gzip-compressing the
// ciphertext, succeeds end-to-end and produces a non-empty PDF document.
func TestGenerateCommand_PGPMode(t *testing.T) {
	outPath, err := runGenerate(t, []byte(decodeTestPlaintext))
	require.NoError(t, err)

	pdfBytes, err := os.ReadFile(outPath)
	require.NoError(t, err)
	assert.NotEmpty(t, pdfBytes)
	assert.True(t, bytes.HasPrefix(pdfBytes, []byte("%PDF")), "output should be a PDF document")
}

// TestGenerateCommand_RawMode asserts that "--raw" mode, which now stores
// the input completely uncompressed and unencrypted, still succeeds
// end-to-end and produces a valid PDF document.
func TestGenerateCommand_RawMode(t *testing.T) {
	outPath, err := runGenerate(t, []byte(decodeTestPlaintext), "--raw")
	require.NoError(t, err)

	pdfBytes, err := os.ReadFile(outPath)
	require.NoError(t, err)
	assert.NotEmpty(t, pdfBytes)
	assert.True(t, bytes.HasPrefix(pdfBytes, []byte("%PDF")), "output should be a PDF document")
}

// TestGenerateCommand_NoCodeMode asserts that the new primary "--no-code"
// flag name is correctly wired through to command execution (skipping 2D
// code generation), and that the command still succeeds and produces a
// valid PDF document.
func TestGenerateCommand_NoCodeMode(t *testing.T) {
	outPath, err := runGenerate(t, []byte(decodeTestPlaintext), "--no-code")
	require.NoError(t, err)

	pdfBytes, err := os.ReadFile(outPath)
	require.NoError(t, err)
	assert.NotEmpty(t, pdfBytes)
	assert.True(t, bytes.HasPrefix(pdfBytes, []byte("%PDF")), "output should be a PDF document")
}

// TestGenerateCommand_RawModeWithNoCode asserts that "--raw" and "--no-code"
// can be combined successfully, exercising both changed code paths together.
func TestGenerateCommand_RawModeWithNoCode(t *testing.T) {
	outPath, err := runGenerate(t, []byte(decodeTestPlaintext), "--raw", "--no-code")
	require.NoError(t, err)

	pdfBytes, err := os.ReadFile(outPath)
	require.NoError(t, err)
	assert.NotEmpty(t, pdfBytes)
	assert.True(t, bytes.HasPrefix(pdfBytes, []byte("%PDF")), "output should be a PDF document")
}
