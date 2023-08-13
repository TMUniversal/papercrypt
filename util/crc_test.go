package util

import (
	"testing"
)

func TestChecksumValidation(t *testing.T) {
	data := []byte{0x12, 0x34, 0x56, 0x78, 0x9A, 0xBC, 0xDE, 0xF0}
	generateCRCTable()
	checksum := Crc24Checksum(data)

	valid := ValidateCRC24(data, checksum)
	if !valid {
		t.Errorf("Expected checksum validation to be true, but got false.")
	}
}

func TestChecksumInvalidation(t *testing.T) {
	data := []byte{0x12, 0x34, 0x56, 0x78, 0x9A, 0xBC, 0xDE, 0xF0}
	generateCRCTable()
	checksum := Crc24Checksum(data)

	// Modify the data to invalidate the checksum
	data[0] = 0xAB

	valid := ValidateCRC24(data, checksum)
	if valid {
		t.Errorf("Expected checksum validation to be false, but got true.")
	}
}

func TestBoth(t *testing.T) {
	data := []byte{0x2d, 0x2d, 0x2d, 0x2d, 0x2d, 0x42, 0x45, 0x47, 0x49, 0x4e, 0x20, 0x50, 0x47, 0x50, 0x20, 0x4d, 0x45, 0x53, 0x53, 0x41, 0x47, 0x45}
	generateCRCTable()
	checksum := uint32(0xc55238)

	valid := ValidateCRC24(data, checksum)
	if !valid {
		t.Errorf("Expected checksum validation to be true, but got false.")
	}
}
