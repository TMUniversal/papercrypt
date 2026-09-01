/*
 * This file is part of PaperCrypt.
 *
 * PaperCrypt lets you prepare encrypted messages for printing on paper.
 * Copyright (C) 2024-2026 TMUniversal <me@tmuniversal.eu>.
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

package envelope

import (
	"bytes"
	"compress/gzip"
	"fmt"

	"github.com/tmuniversal/papercrypt/v3/internal/decompression"
)

type CompressionType uint8

const (
	CompressionRaw CompressionType = iota
	CompressionGzip
)

func (c CompressionType) String() string {
	switch c {
	case CompressionRaw:
		return "raw"
	case CompressionGzip:
		return "gzip"
	default:
		return "unknown"
	}
}

// Implementations are selected by CompressionType via NewCompressor.
type Compressor interface {
	Compress(data []byte) ([]byte, error)
	Decompress(data []byte) ([]byte, error)
	CompressionType() CompressionType
}

// CompressorOption configures a Compressor created by NewCompressor.
type CompressorOption func(*compressorConfig)

type compressorConfig struct {
	maxDecompressedSize int
}

// WithMaxDecompressedSize overrides the gzip decompressed-size cap.
// A negative value disables the cap; zero keeps the package default.
func WithMaxDecompressedSize(maxBytes int) CompressorOption {
	return func(c *compressorConfig) { c.maxDecompressedSize = maxBytes }
}

// WithNoDecompressionLimit disables the decompressed-size cap entirely.
func WithNoDecompressionLimit() CompressorOption {
	return WithMaxDecompressedSize(-1)
}

func NewCompressor(t CompressionType, opts ...CompressorOption) (Compressor, error) {
	var cfg compressorConfig
	for _, opt := range opts {
		opt(&cfg)
	}

	switch t {
	case CompressionRaw:
		return RawCompressor{}, nil
	case CompressionGzip:
		return GzipCompressor(cfg), nil
	default:
		return nil, fmt.Errorf("unsupported envelope compression type %d", t)
	}
}

type RawCompressor struct{}

func (RawCompressor) Compress(data []byte) ([]byte, error) {
	return data, nil
}

func (RawCompressor) Decompress(data []byte) ([]byte, error) {
	return data, nil
}

func (RawCompressor) CompressionType() CompressionType {
	return CompressionRaw
}

type GzipCompressor struct {
	maxDecompressedSize int
}

func (GzipCompressor) Compress(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	gz, err := gzip.NewWriterLevel(&buf, gzip.BestCompression)
	if err != nil {
		return nil, fmt.Errorf("envelope: creating gzip writer: %w", err)
	}
	if _, err := gz.Write(data); err != nil {
		return nil, fmt.Errorf("envelope: writing gzip data: %w", err)
	}
	if err := gz.Close(); err != nil {
		return nil, fmt.Errorf("envelope: closing gzip writer: %w", err)
	}
	return buf.Bytes(), nil
}

func (c GzipCompressor) Decompress(data []byte) ([]byte, error) {
	gz, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("envelope: creating gzip reader: %w", err)
	}

	out, err := decompression.ReadAll(gz, c.maxDecompressedSize)
	if err != nil {
		return nil, fmt.Errorf("envelope: reading gzip data: %w", err)
	}
	if err := gz.Close(); err != nil {
		return nil, fmt.Errorf("envelope: closing gzip reader: %w", err)
	}
	return out, nil
}

func (GzipCompressor) CompressionType() CompressionType {
	return CompressionGzip
}

// selectCompression stores content raw unless gzip makes it strictly smaller.
func selectCompression(content []byte) ([]byte, CompressionType) {
	stored, err := GzipCompressor{}.Compress(content)
	if err != nil || len(stored) >= len(content) {
		return content, CompressionRaw
	}
	return stored, CompressionGzip
}
