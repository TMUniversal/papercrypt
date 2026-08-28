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

// Compressor compresses and decompresses content for the envelope.
// Implementations are selected by CompressionType via NewCompressor.
type Compressor interface {
	// Compress compresses the given data.
	Compress(data []byte) ([]byte, error)
	// Decompress decompresses the given data.
	Decompress(data []byte) ([]byte, error)
	// CompressionType identifies this compressor, stored in the header.
	CompressionType() CompressionType
}

// NewCompressor returns the Compressor registered for the given compression
// type, or an error if the type is not supported.
func NewCompressor(t CompressionType) (Compressor, error) {
	switch t {
	case CompressionRaw:
		return RawCompressor{}, nil
	case CompressionGzip:
		return GzipCompressor{}, nil
	default:
		return nil, fmt.Errorf("unsupported envelope compression type %d", t)
	}
}

// RawCompressor implements Compressor without transforming the data.
type RawCompressor struct{}

// Compress returns the data unchanged.
func (RawCompressor) Compress(data []byte) ([]byte, error) {
	return data, nil
}

// Decompress returns the data unchanged.
func (RawCompressor) Decompress(data []byte) ([]byte, error) {
	return data, nil
}

// CompressionType returns CompressionRaw.
func (RawCompressor) CompressionType() CompressionType {
	return CompressionRaw
}

// GzipCompressor implements Compressor using gzip.
type GzipCompressor struct{}

// Compress compresses data using gzip at BestCompression level.
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

// Decompress expands a gzip-compressed payload.
func (GzipCompressor) Decompress(data []byte) ([]byte, error) {
	gz, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("envelope: creating gzip reader: %w", err)
	}

	out, err := io.ReadAll(gz)
	if err != nil {
		return nil, fmt.Errorf("envelope: reading gzip data: %w", err)
	}
	if err := gz.Close(); err != nil {
		return nil, fmt.Errorf("envelope: closing gzip reader: %w", err)
	}
	return out, nil
}

// CompressionType returns CompressionGzip.
func (GzipCompressor) CompressionType() CompressionType {
	return CompressionGzip
}

// selectCompression returns the content to store, gzip-compressing it when
// that makes it strictly smaller than the original, together with the
// compression type recorded in the envelope header.
func selectCompression(content []byte) ([]byte, CompressionType) {
	stored, err := GzipCompressor{}.Compress(content)
	if err != nil || len(stored) >= len(content) {
		return content, CompressionRaw
	}
	return stored, CompressionGzip
}
