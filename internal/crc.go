package internal

import (
	"hash/crc32"
)

const (
	CRC24Polynomial = uint32(0x864CFB) // CRC-24 polynomial
	CRC24Initial    = uint32(0xB704CE) // Initial value
	CRC24TableSize  = uint32(256)      // Table size for faster computation
)

var crc24Table [CRC24TableSize]uint32

func generateCRCTable() {
	for i := uint32(0); i < CRC24TableSize; i++ {
		crc := uint32(i) << 16
		for j := 0; j < 8; j++ {
			if (crc & 0x800000) != 0 {
				crc = (crc << 1) ^ CRC24Polynomial
			} else {
				crc <<= 1
			}
		}
		crc24Table[i] = crc & 0xFFFFFF
	}
}

func Crc24Checksum(data []byte) uint32 {
	if crc24Table[0] == 0 {
		generateCRCTable()
	}

	crc := CRC24Initial

	for _, b := range data {
		index := byte(crc>>16) ^ b
		crc = (crc << 8) ^ crc24Table[index]
	}

	return crc & 0xFFFFFF
}

func ValidateCRC24(data []byte, checksum uint32) bool {
	return Crc24Checksum(data) == checksum
}

func ValidateCRC32(data []byte, checksum uint32) bool {
	return crc32.ChecksumIEEE(data) == checksum
}
