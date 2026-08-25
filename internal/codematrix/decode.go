package codematrix

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"errors"
	"hash/crc32"
	"image"
	"io"

	"github.com/caarlos0/log"
	"github.com/makiuchi-d/gozxing"
	"github.com/makiuchi-d/gozxing/aztec"
)

// Decode decodes a single Aztec code image, verifies CRC32 integrity,
// and returns the original data. A CRC32 mismatch is logged as a warning.
func Decode(img image.Image) ([]byte, error) {
	bmp, err := gozxing.NewBinaryBitmapFromImage(img)
	if err != nil {
		return nil, errors.Join(errors.New("codematrix: create bitmap"), err)
	}

	reader := aztec.NewAztecReader()
	result, err := reader.Decode(bmp, nil)
	if err != nil {
		return nil, errors.Join(errors.New("codematrix: aztec decode"), err)
	}

	flag, expectedCRC, payload, err := parseHeader([]byte(result.GetText()))
	if err != nil {
		return nil, err
	}

	var data []byte
	switch flag {
	case EncGzip:
		b64, err := base64.StdEncoding.DecodeString(string(payload))
		if err != nil {
			return nil, errors.Join(errors.New("codematrix: base64 decode"), err)
		}
		gz, err := gzip.NewReader(bytes.NewReader(b64))
		if err != nil {
			return nil, errors.Join(errors.New("codematrix: gzip reader"), err)
		}
		data, err = io.ReadAll(gz)
		if err != nil {
			return nil, errors.Join(errors.New("codematrix: gzip read"), err)
		}
	case EncRaw:
		data = payload
	default:
		return nil, errors.New("codematrix: unknown flag")
	}

	actualCRC := crc32.ChecksumIEEE(data)
	if actualCRC != expectedCRC {
		log.Warnf("CRC32 mismatch: expected %08X, got %08X — data may be corrupted",
			expectedCRC, actualCRC)
	}

	return data, nil
}
