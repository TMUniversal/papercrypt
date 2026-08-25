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
	"compress/gzip"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tmuniversal/papercrypt/v3/internal"
)

// decodeTestPlaintext is the plaintext payload embedded in the documents built below.
const decodeTestPlaintext = `{"message":"hello, papercrypt!"}`

// buildV3PGPDocument builds a fully valid PaperCrypt v3 text document whose
// data was encrypted (with passphrase "example") and gzip compressed, mirroring
// the (new) order of operations used by the generate command: encrypt, then compress.
func buildV3PGPDocument(t *testing.T) []byte {
	t.Helper()

	ciphertext, err := encrypt([]byte("example"), []byte(decodeTestPlaintext))
	require.NoError(t, err)

	var buf bytes.Buffer
	gz, err := gzip.NewWriterLevel(&buf, gzip.BestCompression)
	require.NoError(t, err)
	_, err = gz.Write(ciphertext.Bytes())
	require.NoError(t, err)
	require.NoError(t, gz.Close())

	pc := internal.NewPaperCrypt(
		"3.0.0",
		buf.Bytes(),
		"TESTSN",
		"Test Purpose",
		"Test Comment",
		time.Now(),
		internal.PaperCryptDataFormatPGP,
	)

	text, err := pc.GetText(false)
	require.NoError(t, err)
	return text
}

// buildV3RawDocument builds a valid PaperCrypt v3 text document holding raw
// (unencrypted, uncompressed) data, as produced by `generate --raw`.
func buildV3RawDocument(t *testing.T) []byte {
	t.Helper()

	pc := internal.NewPaperCrypt(
		"3.0.0",
		[]byte(decodeTestPlaintext),
		"TESTSN",
		"Test Purpose",
		"Test Comment",
		time.Now(),
		internal.PaperCryptDataFormatRaw,
	)

	text, err := pc.GetText(false)
	require.NoError(t, err)
	return text
}

// runDecode writes docContent to a temp file, runs the decode command against
// it (always supplying the "example" passphrase), and returns the resulting
// output file contents along with any error returned by command execution.
func runDecode(t *testing.T, docContent []byte, extraArgs ...string) ([]byte, error) {
	t.Helper()

	tempDir := t.TempDir()
	inPath := filepath.Join(tempDir, "input.txt")
	outPath := filepath.Join(tempDir, "output.json")

	require.NoError(t, os.WriteFile(inPath, docContent, 0o600))

	args := append([]string{"decode", "-i", inPath, "-o", outPath, "-P", "example"}, extraArgs...)
	rootCmd.SetArgs(args)

	err := rootCmd.Execute()

	out, readErr := os.ReadFile(outPath)
	if readErr != nil {
		return nil, err
	}
	return out, err
}

func TestDecodeV3PGP(t *testing.T) {
	doc := buildV3PGPDocument(t)

	out, err := runDecode(t, doc)
	assert.NoError(t, err)
	assert.Equal(t, decodeTestPlaintext, string(out))
}

func TestDecodeV3Raw(t *testing.T) {
	doc := buildV3RawDocument(t)

	out, err := runDecode(t, doc)
	assert.NoError(t, err)
	assert.Equal(t, decodeTestPlaintext, string(out))
}

// TestDecodeUnsupportedVersionV1 asserts that PaperCrypt v1 documents, which
// were supported prior to this change, are now rejected as an unknown version.
func TestDecodeUnsupportedVersionV1(t *testing.T) {
	doc := []byte("# PaperCrypt Version: 1.3.0\n# Content Serial: ABCDEF\n\n\n 1: AA\n")

	_, err := runDecode(t, doc)
	assert.Error(t, err)
	assert.ErrorContains(t, err, "unknown version")
}

// TestDecodeUnsupportedVersionV2 asserts that PaperCrypt v2 documents, which
// were supported prior to this change (via DeserializeV2Text), are now
// rejected as an unknown version, since only devel/v3 documents are handled.
func TestDecodeUnsupportedVersionV2(t *testing.T) {
	doc := []byte("# PaperCrypt Version: 2.0.0\n# Content Serial: ABCDEF\n\n\n 1: AA\n")

	_, err := runDecode(t, doc)
	assert.Error(t, err)
	assert.ErrorContains(t, err, "unknown version")
}

// TestDecodeDevelVersionAccepted asserts that a "devel" build version string
// is still accepted, exercising the other branch of the version switch.
func TestDecodeDevelVersionAccepted(t *testing.T) {
	pc := internal.NewPaperCrypt(
		"devel",
		[]byte(decodeTestPlaintext),
		"TESTSN",
		"",
		"",
		time.Now(),
		internal.PaperCryptDataFormatRaw,
	)

	text, err := pc.GetText(false)
	require.NoError(t, err)

	out, err := runDecode(t, text)
	assert.NoError(t, err)
	assert.Equal(t, decodeTestPlaintext, string(out))
}

// TestDecodeUnknownVersionHeaderMissing verifies that a document with no
// version header at all is rejected before any further processing.
func TestDecodeUnknownVersionHeaderMissing(t *testing.T) {
	doc := []byte("# Content Serial: ABCDEF\n\n\n 1: AA\n")

	_, err := runDecode(t, doc)
	assert.Error(t, err)
	assert.ErrorContains(t, err, "unknown version")
}
