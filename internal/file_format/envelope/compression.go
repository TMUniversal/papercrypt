package envelope

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
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

type GzipCompressor struct{}

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
