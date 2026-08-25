package codematrix

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"errors"
	"image"
	"io"

	"github.com/makiuchi-d/gozxing"
	"github.com/makiuchi-d/gozxing/aztec"
)

// Decode reads a single Aztec code image, base64-decodes the gzip payload,
// and returns the original data.
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

	b64, err := base64.StdEncoding.DecodeString(result.GetText())
	if err != nil {
		return nil, errors.Join(errors.New("codematrix: base64 decode"), err)
	}

	gz, err := gzip.NewReader(bytes.NewReader(b64))
	if err != nil {
		return nil, errors.Join(errors.New("codematrix: gzip reader"), err)
	}
	data, err := io.ReadAll(gz)
	if err != nil {
		return nil, errors.Join(errors.New("codematrix: gzip read"), err)
	}

	return data, nil
}
