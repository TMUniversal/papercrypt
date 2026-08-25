package codematrix

import (
	"fmt"

	"github.com/makiuchi-d/gozxing"
	"github.com/makiuchi-d/gozxing/common/reedsolomon"
	"github.com/makiuchi-d/gozxing/qrcode/decoder"
	"github.com/makiuchi-d/gozxing/qrcode/encoder"
)

func encodeQRMatrix(dataBits *gozxing.BitArray) (*gozxing.BitMatrix, error) {
	interleaved, err := interleave(dataBits)
	if err != nil {
		return nil, err
	}

	version, _ := decoder.Version_GetVersionForNumber(qrVersion)
	dim := version.GetDimensionForVersion()

	bestMask := -1
	bestPenalty := int(^uint(0) >> 1)

	for mask := 0; mask < 8; mask++ {
		bm := encoder.NewByteMatrix(dim, dim)
		if err := encoder.MatrixUtil_buildMatrix(interleaved, ecLevel, version, mask, bm); err != nil {
			continue
		}
		p := calculatePenalty(bm, dim)
		if p < bestPenalty {
			bestPenalty = p
			bestMask = mask
		}
	}

	if bestMask == -1 {
		return nil, errNoValidMask
	}

	bm := encoder.NewByteMatrix(dim, dim)
	if err := encoder.MatrixUtil_buildMatrix(interleaved, ecLevel, version, bestMask, bm); err != nil {
		return nil, fmt.Errorf("codematrix: build matrix: %w", err)
	}

	return byteMatrixToBitMatrix(bm), nil
}

func byteMatrixToBitMatrix(bm *encoder.ByteMatrix) *gozxing.BitMatrix {
	w := bm.GetWidth()
	h := bm.GetHeight()
	out, _ := gozxing.NewBitMatrix(w, h)

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			if bm.Get(x, y) == 1 {
				out.Set(x, y)
			}
		}
	}

	return out
}

func interleave(dataBits *gozxing.BitArray) (*gozxing.BitArray, error) {
	version, _ := decoder.Version_GetVersionForNumber(qrVersion)
	ecBlocks := version.GetECBlocksForLevel(ecLevel)
	numDataBytes := totalCodewords - ecBlocks.GetTotalECCodewords()
	numRSBlocks := ecBlocks.GetNumBlocks()

	if dataBits.GetSizeInBytes() != numDataBytes {
		return nil, errDataBitsMismatch
	}

	type blockPair struct {
		data []byte
		ec   []byte
	}

	blocks := make([]blockPair, numRSBlocks)
	offset := 0
	maxData := 0
	maxEC := 0

	for i := 0; i < numRSBlocks; i++ {
		dataLen, ecLen := blockSizes(numDataBytes, numRSBlocks, i)
		data := make([]byte, dataLen)
		dataBits.ToBytes(8*offset, data, 0, dataLen)

		ec, err := rsEncode(data, ecLen)
		if err != nil {
			return nil, err
		}

		blocks[i] = blockPair{data: data, ec: ec}
		offset += dataLen
		if dataLen > maxData {
			maxData = dataLen
		}
		if len(ec) > maxEC {
			maxEC = len(ec)
		}
	}

	if offset != numDataBytes {
		return nil, errDataBitsMismatch
	}

	result := gozxing.NewEmptyBitArray()

	for i := 0; i < maxData; i++ {
		for _, b := range blocks {
			if i < len(b.data) {
				result.AppendBits(int(b.data[i]), 8)
			}
		}
	}
	for i := 0; i < maxEC; i++ {
		for _, b := range blocks {
			if i < len(b.ec) {
				result.AppendBits(int(b.ec[i]), 8)
			}
		}
	}

	if totalCodewords != result.GetSizeInBytes() {
		return nil, errInterleaveMismatch
	}

	return result, nil
}

func blockSizes(numDataBytes, numRSBlocks, blockID int) (int, int) {
	group2 := totalCodewords % numRSBlocks
	group1 := numRSBlocks - group2
	dataPerBlock := numDataBytes / numRSBlocks
	totalPerBlock := totalCodewords / numRSBlocks

	if blockID < group1 {
		return dataPerBlock, totalPerBlock - dataPerBlock
	}
	return dataPerBlock + 1, totalPerBlock + 1 - (dataPerBlock + 1)
}

func rsEncode(data []byte, numEC int) ([]byte, error) {
	n := len(data)
	buf := make([]int, n+numEC)
	for i := 0; i < n; i++ {
		buf[i] = int(data[i]) & 0xFF
	}
	if err := reedsolomon.NewReedSolomonEncoder(reedsolomon.GenericGF_QR_CODE_FIELD_256).Encode(buf, numEC); err != nil {
		return nil, fmt.Errorf("codematrix: RS encode: %w", err)
	}
	ec := make([]byte, numEC)
	for i := 0; i < numEC; i++ {
		ec[i] = byte(buf[n+i])
	}
	return ec, nil
}

func calculatePenalty(bm *encoder.ByteMatrix, dim int) int {
	p := 0

	for y := 0; y < dim; y++ {
		cnt := 0
		var last int8 = -1
		for x := 0; x < dim; x++ {
			c := bm.Get(x, y)
			if x > 0 && c == last {
				cnt++
				if cnt == 5 {
					p += 3
				} else if cnt > 5 {
					p++
				}
			} else {
				cnt = 1
			}
			last = c
		}
	}

	for x := 0; x < dim; x++ {
		cnt := 0
		var last int8 = -1
		for y := 0; y < dim; y++ {
			c := bm.Get(x, y)
			if y > 0 && c == last {
				cnt++
				if cnt == 5 {
					p += 3
				} else if cnt > 5 {
					p++
				}
			} else {
				cnt = 1
			}
			last = c
		}
	}

	for y := 0; y < dim-1; y++ {
		for x := 0; x < dim-1; x++ {
			c := bm.Get(x, y)
			if c == bm.Get(x+1, y) && c == bm.Get(x, y+1) && c == bm.Get(x+1, y+1) {
				p += 3
			}
		}
	}

	return p
}
