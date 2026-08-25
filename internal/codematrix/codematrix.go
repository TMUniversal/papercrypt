package codematrix

import (
	"fmt"

	"github.com/makiuchi-d/gozxing/qrcode/decoder"
)

const (
	MaxSymbols   = 16
	quietModules = 4

	qrVersion        = 40
	totalCodewords   = 3706
	numDataCodewords = 1276
	maxChunkBytes    = 1267

	ecLevel = decoder.ErrorCorrectionLevel_H
)

func GridDimensions(n int) (rows, cols int) {
	switch {
	case n <= 1:
		return 1, 1
	case n == 2:
		return 1, 2
	default:
		return 2, 2
	}
}

func CodeLabel(index, total int) string {
	return fmt.Sprintf("%d/%d", index+1, total)
}

func splitData(data []byte) [][]byte {
	n := len(data) / maxChunkBytes
	if len(data)%maxChunkBytes > 0 {
		n++
	}
	if n == 0 {
		n = 1
	}

	chunks := make([][]byte, n)
	for i := 0; i < n; i++ {
		start := i * maxChunkBytes
		end := start + maxChunkBytes
		if end > len(data) {
			end = len(data)
		}
		chunks[i] = data[start:end]
	}
	return chunks
}
