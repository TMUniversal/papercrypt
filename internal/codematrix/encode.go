package codematrix

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"errors"
	"fmt"
	"image"

	"github.com/makiuchi-d/gozxing"
	"github.com/makiuchi-d/gozxing/datamatrix"
	"github.com/makiuchi-d/gozxing/datamatrix/encoder"
)

func Encode(data []byte) ([]image.Image, error) {
	var buf bytes.Buffer
	gz, err := gzip.NewWriterLevel(&buf, gzip.BestCompression)
	if err != nil {
		return nil, errors.Join(errors.New("codematrix: failed to create gzip writer"), err)
	}
	if _, err := gz.Write(data); err != nil {
		return nil, errors.Join(errors.New("codematrix: failed to write gzip data"), err)
	}
	if err := gz.Close(); err != nil {
		return nil, errors.Join(errors.New("codematrix: failed to close gzip writer"), err)
	}

	compressed := buf.Bytes()
	if len(compressed) == 0 {
		compressed = []byte{0}
	}

	numSymbols := (len(compressed) + MaxPayload - 1) / MaxPayload
	if numSymbols > MaxSymbols {
		return nil, fmt.Errorf("codematrix: data too large: %d bytes compressed requires %d symbols (max %d)", len(compressed), numSymbols, MaxSymbols)
	}

	images := make([]image.Image, numSymbols)
	for i := 0; i < numSymbols; i++ {
		start := i * MaxPayload
		end := start + MaxPayload
		if end > len(compressed) {
			end = len(compressed)
		}
		chunk := compressed[start:end]

		h := chunkHeader{
			Version:  Version,
			Index:    byte(i),
			Total:    byte(numSymbols),
			CRC24:    crc24Checksum(chunk),
			Reserved: 0,
		}
		hdr := h.Marshal()

		inner := make([]byte, 0, HeaderSize+len(chunk))
		inner = append(inner, hdr[:]...)
		inner = append(inner, chunk...)

		payload := base64.StdEncoding.EncodeToString(inner)

		dmWriter := datamatrix.NewDataMatrixWriter()
		hints := map[gozxing.EncodeHintType]interface{}{
			gozxing.EncodeHintType_DATA_MATRIX_SHAPE: encoder.SymbolShapeHint_FORCE_SQUARE,
		}

		bits, err := dmWriter.Encode(payload, gozxing.BarcodeFormat_DATA_MATRIX, 0, 0, hints)
		if err != nil {
			return nil, errors.Join(fmt.Errorf("codematrix: failed to encode symbol %d/%d", i+1, numSymbols), err)
		}

		images[i] = bits
	}

	return images, nil
}
