package codematrix

import (
	"github.com/makiuchi-d/gozxing"
	"github.com/makiuchi-d/gozxing/qrcode/decoder"
)

func buildPayload(data []byte, index, total int) (*gozxing.BitArray, error) {
	bits := gozxing.NewEmptyBitArray()

	sa := SAHeader{Index: byte(index), Total: byte(total)}
	saBits := sa.Encode()
	bits.AppendBitArray(saBits)

	dh := DataHeader{Mode: 0x04, Length: len(data)}
	dhBits := dh.Encode(qrVersion, data)
	bits.AppendBitArray(dhBits)

	for _, b := range data {
		bits.AppendBits(int(b), 8)
	}

	version, _ := decoder.Version_GetVersionForNumber(qrVersion)
	ecBlocks := version.GetECBlocksForLevel(ecLevel)
	numDataBytes := totalCodewords - ecBlocks.GetTotalECCodewords()
	capacity := numDataBytes * 8

	if bits.GetSize() > capacity {
		return nil, errPayloadTooLarge
	}

	terminateBits(bits, numDataBytes)

	return bits, nil
}

func terminateBits(bits *gozxing.BitArray, numDataBytes int) {
	capacity := numDataBytes * 8

	for i := 0; i < 4 && bits.GetSize() < capacity; i++ {
		bits.AppendBit(false)
	}

	remaining := bits.GetSize() & 0x07
	if remaining > 0 {
		for i := remaining; i < 8; i++ {
			bits.AppendBit(false)
		}
	}

	padBytes := numDataBytes - bits.GetSizeInBytes()
	for i := 0; i < padBytes; i++ {
		v := 0xEC
		if (i & 0x1) == 0 {
			v = 0x11
		}
		bits.AppendBits(v, 8)
	}
}
