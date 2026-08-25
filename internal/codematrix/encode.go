package codematrix

import (
	"fmt"
	"hash/crc32"
	"image"
	"image/color"

	"github.com/makiuchi-d/gozxing"
	"github.com/makiuchi-d/gozxing/multi/qrcode"
)

func Encode(data []byte) ([]image.Image, error) {
	chunks := splitData(data)
	if len(chunks) > MaxSymbols {
		return nil, fmt.Errorf("codematrix: data too large: %d chunks (max %d)", len(chunks), MaxSymbols)
	}

	images := make([]image.Image, len(chunks))
	for i, chunk := range chunks {
		payload, err := buildPayload(chunk, i, len(chunks))
		if err != nil {
			return nil, fmt.Errorf("codematrix: build payload %d/%d: %w", i+1, len(chunks), err)
		}

		bm, err := encodeQRMatrix(payload)
		if err != nil {
			return nil, fmt.Errorf("codematrix: encode QR %d/%d: %w", i+1, len(chunks), err)
		}

		images[i] = bitMatrixToImage(bm)
	}

	return images, nil
}

func Decode(images []image.Image) ([]byte, error) {
	if len(images) == 0 {
		return nil, errNoImages
	}
	if len(images) > MaxSymbols {
		return nil, errTooManyImages
	}

	reader := qrcode.NewQRCodeMultiReader()

	var allResults []*gozxing.Result
	for i, img := range images {
		bm, err := gozxing.NewBinaryBitmapFromImage(img)
		if err != nil {
			return nil, fmt.Errorf("codematrix: create bitmap %d: %w", i+1, err)
		}
		results, err := reader.DecodeMultipleWithoutHint(bm)
		if err != nil {
			return nil, fmt.Errorf("codematrix: decode QR %d: %w", i+1, err)
		}
		allResults = append(allResults, results...)
	}

	if len(allResults) == 0 {
		return nil, fmt.Errorf("codematrix: no QR codes decoded")
	}

	var raw []byte
	if len(allResults) == 1 {
		raw = []byte(allResults[0].GetText())
	} else {
		for _, r := range allResults {
			raw = append(raw, []byte(r.GetText())...)
		}
	}

	return decodePayload(raw)
}

func decodePayload(raw []byte) ([]byte, error) {
	bits := gozxing.NewEmptyBitArray()
	for _, b := range raw {
		bits.AppendBits(int(b), 8)
	}

	offset := 0

	_, newOffset, err := DecodeSAHeader(bits, offset)
	if err != nil {
		return nil, fmt.Errorf("codematrix: decode SA header: %w", err)
	}
	offset = newOffset

	dh, newOffset, err := DecodeDataHeader(bits, offset, qrVersion)
	if err != nil {
		return nil, fmt.Errorf("codematrix: decode data header: %w", err)
	}
	offset = newOffset

	remaining := bits.GetSize() - offset
	needed := dh.Length * 8
	if remaining < needed {
		return nil, fmt.Errorf("codematrix: not enough data bits: need %d, have %d", needed, remaining)
	}

	data := make([]byte, dh.Length)
	bits.ToBytes(offset, data, 0, dh.Length)

	if crc32.ChecksumIEEE(data) != dh.CRC32 {
		return nil, fmt.Errorf("codematrix: CRC32 mismatch")
	}

	return data, nil
}

func bitMatrixToImage(bm *gozxing.BitMatrix) image.Image {
	w := bm.GetWidth()
	h := bm.GetHeight()
	gray := image.NewGray(image.Rect(0, 0, w, h))

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			if bm.Get(x, y) {
				gray.SetGray(x, y, color.Gray{Y: 0})
			} else {
				gray.SetGray(x, y, color.Gray{Y: 255})
			}
		}
	}

	return gray
}
