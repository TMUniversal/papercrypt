package envelope

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
)

// CompressionType identifies the compression applied to an envelope's content.
type CompressionType uint8

const (
	// CompressionRaw marks content stored without compression.
	CompressionRaw CompressionType = iota
	// CompressionGzip marks content compressed with gzip.
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

// compress returns content, optionally gzip-compressed with BestCompression
// when that makes it strictly smaller than the original.
func compress(content []byte) ([]byte, CompressionType) {
	var buf bytes.Buffer
	gz, err := gzip.NewWriterLevel(&buf, gzip.BestCompression)
	if err != nil {
		return content, CompressionRaw
	}
	if _, err := gz.Write(content); err != nil {
		return content, CompressionRaw
	}
	if err := gz.Close(); err != nil {
		return content, CompressionRaw
	}

	if buf.Len() >= len(content) {
		return content, CompressionRaw
	}
	return buf.Bytes(), CompressionGzip
}

// decompress expands a gzip-compressed payload.
func decompress(data []byte) ([]byte, error) {
	gz, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("envelope: creating gzip reader: %w", err)
	}
	defer gz.Close()

	out, err := io.ReadAll(gz)
	if err != nil {
		return nil, fmt.Errorf("envelope: reading gzip data: %w", err)
	}
	return out, nil
}
