package util

const (
	Polynomial = uint32(0x864CFB) // CRC-24 polynomial
	Initial    = uint32(0xB704CE) // Initial value
	TableSize  = uint32(256)      // Table size for faster computation

	ReversedPolynomial = uint32(0xDF3261)
	Reciprocal         = uint32(0xBE64C3)
	ReversedReciprocal = uint32(0xC3267D)
)

var crc24Table [TableSize]uint32

func generateCRCTable() {
	for i := uint32(0); i < TableSize; i++ {
		crc := uint32(i) << 16
		for j := 0; j < 8; j++ {
			if (crc & 0x800000) != 0 {
				crc = (crc << 1) ^ Polynomial
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

	crc := Initial

	for _, b := range data {
		index := byte(crc>>16) ^ b
		crc = (crc << 8) ^ crc24Table[index]
	}

	return crc & 0xFFFFFF
}

func ValidateCRC24(data []byte, checksum uint32) bool {
	return Crc24Checksum(data) == checksum
}
