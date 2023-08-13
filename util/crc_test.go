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
