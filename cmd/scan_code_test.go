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
	"encoding/json"
	"image/png"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/qr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tmuniversal/papercrypt/v3/internal"
)

// runScan writes inputData to a temp file, runs the scan command against it
// (optionally with --from-json), and returns the resulting output file
// contents along with any error returned by command execution.
func runScan(t *testing.T, inputData []byte, fromJSON bool) ([]byte, error) {
	t.Helper()

	tempDir := t.TempDir()
	inPath := filepath.Join(tempDir, "input.bin")
	outPath := filepath.Join(tempDir, "output.out")

	require.NoError(t, os.WriteFile(inPath, inputData, 0o600))

	args := []string{"scan", "-i", inPath, "-o", outPath}
	if fromJSON {
		args = append(args, "--from-json")
	}
	rootCmd.SetArgs(args)

	err := rootCmd.Execute()

	out, readErr := os.ReadFile(outPath)
	if readErr != nil {
		return nil, err
	}
	return out, err
}

// TestScanFromJSON_V3Success asserts that a PaperCrypt v3 document, provided
// as JSON via --from-json, is correctly deserialized and re-serialized to
// text form, matching what GetText produces directly.
func TestScanFromJSON_V3Success(t *testing.T) {
	pc := internal.NewPaperCrypt(
		"3.0.0",
		[]byte(decodeTestPlaintext),
		"TESTSN",
		"Test Purpose",
		"Test Comment",
		time.Now(),
		internal.PaperCryptDataFormatRaw,
	)

	jsonData, err := json.Marshal(pc)
	require.NoError(t, err)

	expectedText, err := pc.GetText(false)
	require.NoError(t, err)

	out, err := runScan(t, jsonData, true)
	require.NoError(t, err)
	assert.Equal(t, string(expectedText), string(out))
}

// TestScanFromJSON_UnsupportedVersion asserts that a document reporting an
// unsupported version (here, the removed v2 format) via the "v" field is
// rejected with an "unknown version" error, mirroring cmd/decode.go's
// version-dispatch behavior.
func TestScanFromJSON_UnsupportedVersion(t *testing.T) {
	doc := []byte(`{"v":"2.0.0"}`)

	_, err := runScan(t, doc, true)
	assert.Error(t, err)
	assert.ErrorContains(t, err, "unknown version: 2.0.0")
}

// TestScanFromJSON_LegacyFieldNameNotRecognized asserts that PaperCrypt v1's
// JSON shape, which used a capitalized "Version" field name and was
// previously recognized via a separate versionContainerV1 fallback, is no
// longer recognized at all now that the fallback struct has been removed:
// the version is parsed as empty, and the document is rejected as an
// unknown version.
func TestScanFromJSON_LegacyFieldNameNotRecognized(t *testing.T) {
	doc := []byte(`{"Version":"1.3.0"}`)

	_, err := runScan(t, doc, true)
	assert.Error(t, err)
	assert.ErrorContains(t, err, "unknown version:")
}

// TestScanCmd_QRCodeNotSupported asserts that the codematrix-based decode
// path introduced by this PR only supports Aztec codes: a valid QR code
// image, which was decodable via the old inline fallback logic, is now
// rejected with an "error decoding aztec code" error.
func TestScanCmd_QRCodeNotSupported(t *testing.T) {
	code, err := qr.Encode("papercrypt qr fallback removed", qr.M, qr.Auto)
	require.NoError(t, err)

	code, err = barcode.Scale(code, 200, 200)
	require.NoError(t, err)

	var buf bytes.Buffer
	require.NoError(t, png.Encode(&buf, code))

	_, err = runScan(t, buf.Bytes(), false)
	assert.Error(t, err)
	assert.ErrorContains(t, err, "error decoding aztec code")
}
