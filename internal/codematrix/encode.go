package codematrix

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"hash/crc32"
	"image"
	"image/png"

	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/aztec"
)

// Encode encodes data into a single Aztec code image, choosing gzip+base64
// or raw encoding based on whichever produces the smaller payload.
func Encode(data []byte) (image.Image, error) {
	crc := crc32.ChecksumIEEE(data)
	crcHex := []byte(fmt.Sprintf("%08X", crc))

	rawPayload := make([]byte, 0, headerSize+len(data))
	rawPayload = append(rawPayload, EncRaw)
	rawPayload = append(rawPayload, crcHex...)
	rawPayload = append(rawPayload, data...)

	gzBytes, err := gzipBase64(data)
	if err != nil {
		return nil, err
	}
	gzPayload := make([]byte, 0, headerSize+len(gzBytes))
	gzPayload = append(gzPayload, EncGzip)
	gzPayload = append(gzPayload, crcHex...)
	gzPayload = append(gzPayload, gzBytes...)

	payload := rawPayload
	if len(gzPayload) < len(rawPayload) {
		payload = gzPayload
	}

	code, err := aztec.Encode(payload, 35, 0)
	if err != nil {
		return nil, errors.Join(errors.New("codematrix: aztec encode"), err)
	}

	qrSize := 7795
	code, err = barcode.Scale(code, qrSize, qrSize)
	if err != nil {
		return nil, errors.Join(errors.New("codematrix: scale"), err)
	}

	gray := image.NewGray(code.Bounds())
	for y := 0; y < code.Bounds().Dy(); y++ {
		for x := 0; x < code.Bounds().Dx(); x++ {
			gray.Set(x, y, code.At(x, y))
		}
	}

	return gray, nil
}

// EncodePNG encodes data into a single Aztec code and returns the
// PNG-encoded bytes.
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

func gzipBase64(data []byte) ([]byte, error) {
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
	return []byte(base64.StdEncoding.EncodeToString(buf.Bytes())), nil
}

func parseHeader(raw []byte) (flag byte, crc uint32, payload []byte, err error) {
	if len(raw) < headerSize {
		return 0, 0, nil, errors.New("codematrix: payload too short")
	}
	flag = raw[0]
	if flag != EncGzip && flag != EncRaw {
		return 0, 0, nil, fmt.Errorf("codematrix: unknown flag %c", flag)
	}
	crcBytes, err := hex.DecodeString(string(raw[1:9]))
	if err != nil {
		return 0, 0, nil, errors.Join(errors.New("codematrix: bad CRC32 hex"), err)
	}
	crc = uint32(crcBytes[0])<<24 |
		uint32(crcBytes[1])<<16 |
		uint32(crcBytes[2])<<8 |
		uint32(crcBytes[3])
	return flag, crc, raw[9:], nil
}
