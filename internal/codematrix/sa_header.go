package codematrix

import "github.com/makiuchi-d/gozxing"

type SAHeader struct {
	Index byte
	Total byte
}

func (h SAHeader) Encode() *gozxing.BitArray {
	bits := gozxing.NewEmptyBitArray()
	bits.AppendBits(0x03, 4)
	bits.AppendBits((int(h.Index)<<4)|int(h.Total-1), 8)
	bits.AppendBits(0x00, 8)
	return bits
}

func DecodeSAHeader(bits *gozxing.BitArray, offset int) (SAHeader, int, error) {
	if offset+16 > bits.GetSize() {
		return SAHeader{}, 0, errSAHeaderTooShort
	}

	mode := 0
	for i := 0; i < 4; i++ {
		if bits.Get(offset + i) {
			mode |= 1 << uint(3-i)
		}
	}
	if mode != 0x03 {
		return SAHeader{}, 0, errInvalidSAMode
	}

	seqByte := byte(0)
	for i := 0; i < 8; i++ {
		if bits.Get(offset + 4 + i) {
			seqByte |= 1 << uint(7-i)
		}
	}

	_ = bits.Get(offset + 12)
	_ = bits.Get(offset + 13)
	_ = bits.Get(offset + 14)
	_ = bits.Get(offset + 15)

	return SAHeader{
		Index: seqByte >> 4,
		Total: (seqByte & 0x0F) + 1,
	}, offset + 16, nil
}
