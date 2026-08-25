package codematrix

import (
	"hash/crc32"

	"github.com/makiuchi-d/gozxing"
)

type DataHeader struct {
	Mode   byte
	Length int
	CRC32  uint32
}

func (h DataHeader) Encode(version int, data []byte) *gozxing.BitArray {
	bits := gozxing.NewEmptyBitArray()
	bits.AppendBits(int(h.Mode), 4)

	countBits := 16
	if version < 10 {
		countBits = 8
	}
	bits.AppendBits(h.Length, countBits)

	crc := crc32.ChecksumIEEE(data)
	h.CRC32 = crc
	bits.AppendBits(int(crc>>24), 8)
	bits.AppendBits(int(crc>>16)&0xFF, 8)
	bits.AppendBits(int(crc>>8)&0xFF, 8)
	bits.AppendBits(int(crc)&0xFF, 8)

	return bits
}

func DecodeDataHeader(bits *gozxing.BitArray, offset int, version int) (DataHeader, int, error) {
	if offset+4 > bits.GetSize() {
		return DataHeader{}, 0, errDataHeaderShort
	}

	mode := byte(0)
	for i := 0; i < 4; i++ {
		if bits.Get(offset + i) {
			mode |= 1 << uint(3-i)
		}
	}
	offset += 4

	if mode != 0x04 {
		return DataHeader{}, 0, errInvalidByteMode
	}

	countBits := 16
	if version < 10 {
		countBits = 8
	}
	if offset+countBits > bits.GetSize() {
		return DataHeader{}, 0, errDataHeaderShort
	}

	length := 0
	for i := 0; i < countBits; i++ {
		if bits.Get(offset + i) {
			length |= 1 << uint(countBits-1-i)
		}
	}
	offset += countBits

	if offset+32 > bits.GetSize() {
		return DataHeader{}, 0, errDataHeaderShort
	}

	var crc uint32
	for i := 0; i < 4; i++ {
		crc <<= 8
		for j := 0; j < 8; j++ {
			if bits.Get(offset + i*8 + j) {
				crc |= 1 << uint(7-j)
			}
		}
	}
	offset += 32

	return DataHeader{Mode: mode, Length: length, CRC32: crc}, offset, nil
}
