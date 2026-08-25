package codematrix

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"errors"
	"fmt"
	"image"
	"io"
	"sort"

	"github.com/makiuchi-d/gozxing"
	"github.com/makiuchi-d/gozxing/datamatrix"
)

func Decode(images []image.Image) ([]byte, error) {
	if len(images) == 0 {
		return nil, errors.New("codematrix: no images provided")
	}
	if len(images) > MaxSymbols {
		return nil, fmt.Errorf("codematrix: too many images: %d (max %d)", len(images), MaxSymbols)
	}

	type decodedChunk struct {
		header  chunkHeader
		payload []byte
	}

	chunks := make([]decodedChunk, len(images))
	dmReader := datamatrix.NewDataMatrixReader()

	for i, img := range images {
		bmp, err := gozxing.NewBinaryBitmapFromImage(img)
		if err != nil {
			return nil, errors.Join(fmt.Errorf("codematrix: failed to create bitmap for image %d", i+1), err)
		}

		result, err := dmReader.Decode(bmp, nil)
		if err != nil {
			return nil, errors.Join(fmt.Errorf("codematrix: failed to decode Data Matrix from image %d", i+1), err)
		}

		inner, err := base64.StdEncoding.DecodeString(result.GetText())
		if err != nil {
			return nil, errors.Join(fmt.Errorf("codematrix: failed to base64-decode image %d", i+1), err)
		}

		if len(inner) < HeaderSize {
			return nil, fmt.Errorf("codematrix: image %d payload too short: %d bytes", i+1, len(inner))
		}

		h, err := unmarshalHeader(inner[:HeaderSize])
		if err != nil {
			return nil, fmt.Errorf("codematrix: image %d: %w", i+1, err)
		}

		payload := inner[HeaderSize:]
		if crc24Checksum(payload) != h.CRC24 {
			return nil, fmt.Errorf("codematrix: CRC24 mismatch in symbol %d", i+1)
		}

		chunks[i] = decodedChunk{
			header:  h,
			payload: payload,
		}
	}

	sort.Slice(chunks, func(i, j int) bool {
		return chunks[i].header.Index < chunks[j].header.Index
	})

	total := chunks[0].header.Total
	if int(total) != len(chunks) {
		return nil, fmt.Errorf("codematrix: expected %d symbols, got %d", total, len(chunks))
	}

	for i, c := range chunks {
		if int(c.header.Index) != i {
			return nil, fmt.Errorf("codematrix: missing or duplicate symbol at index %d", i)
		}
		if c.header.Total != total {
			return nil, fmt.Errorf("codematrix: inconsistent total count at symbol %d", i+1)
		}
	}

	var compressed []byte
	for _, c := range chunks {
		compressed = append(compressed, c.payload...)
	}

	gz, err := gzip.NewReader(bytes.NewReader(compressed))
	if err != nil {
		return nil, errors.Join(errors.New("codematrix: failed to create gzip reader"), err)
	}
	defer gz.Close()

	data, err := io.ReadAll(gz)
	if err != nil {
		return nil, errors.Join(errors.New("codematrix: failed to decompress gzip data"), err)
	}

	return data, nil
}
