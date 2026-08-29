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

// maxDecompressedSize caps GzipCompressor.Decompress output, guarding
// against decompression bombs.
const maxDecompressedSize = 1 << 30 // 1 GiB

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

	limit := c.maxDecompressedSize
	if limit == 0 {
		limit = maxDecompressedSize
	}

	in := io.Reader(gz)
	if limit >= 0 {
		in = io.LimitReader(gz, int64(limit)+1)
	}

	out, err := io.ReadAll(in)
	if err != nil {
		return nil, fmt.Errorf("envelope: reading gzip data: %w", err)
	}
	if err := gz.Close(); err != nil {
		return nil, fmt.Errorf("envelope: closing gzip reader: %w", err)
	}
	if limit >= 0 && len(out) > limit {
		return nil, fmt.Errorf(
			"envelope: decompressed content exceeds %d bytes",
			limit,
		)
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
