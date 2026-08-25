package codematrix

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"errors"
	"image"
	"image/png"

	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/aztec"
)

// aztecSize is the Aztec code output size in pixels (165mm at 1200dpi).
const aztecSize = 7795

// Encode compresses data with gzip, base64-encodes it, and encodes it
// into a single Aztec code image.
func Encode(data []byte) (image.Image, error) {
	var buf bytes.Buffer
	gz, err := gzip.NewWriterLevel(&buf, gzip.BestCompression)
	if err != nil {
		return nil, errors.Join(errors.New("codematrix: gzip writer"), err)
	}
	if _, err := gz.Write(data); err != nil {
		return nil, errors.Join(errors.New("codematrix: gzip write"), err)
	}
	if err := gz.Close(); err != nil {
		return nil, errors.Join(errors.New("codematrix: gzip close"), err)
	}

	encoded := base64.StdEncoding.EncodeToString(buf.Bytes())

	code, err := aztec.Encode([]byte(encoded), 35, 0)
	if err != nil {
		return nil, errors.Join(errors.New("codematrix: aztec encode"), err)
	}

	code, err = barcode.Scale(code, aztecSize, aztecSize)
	if err != nil {
		return nil, errors.Join(errors.New("codematrix: scale"), err)
	}

	rgba := image.NewRGBA(code.Bounds())
	for y := 0; y < code.Bounds().Dy(); y++ {
		for x := 0; x < code.Bounds().Dx(); x++ {
			rgba.Set(x, y, code.At(x, y))
		}
	}

	return rgba, nil
}

// EncodePNG encodes data into a single Aztec code and returns the PNG-encoded bytes.
func EncodePNG(data []byte) ([]byte, error) {
	img, err := Encode(data)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, errors.Join(errors.New("codematrix: png encode"), err)
	}
	return buf.Bytes(), nil
}
