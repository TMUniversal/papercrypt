package codematrix

import (
	"fmt"

	"github.com/tmuniversal/papercrypt/v3/internal/crc24"
)

const (
	HeaderSize = 7
	Version    = byte(0x01)
	MaxSymbols = 4

	// MaxPayload is the maximum compressed chunk size per symbol.
	// Largest square DM: 144x144, 1304 data bytes. After base64 decoding:
	// inner ≤ floor(1304/4)*3 = 978 bytes. After 7-byte header: 971 bytes.
	MaxPayload = 971
)

type chunkHeader struct {
	Version  byte
	Index    byte
	Total    byte
	CRC24    uint32
	Reserved byte
}

func (h chunkHeader) Marshal() [HeaderSize]byte {
	var b [HeaderSize]byte
	b[0] = h.Version
	b[1] = h.Index
	b[2] = h.Total
	b[3] = byte(h.CRC24 >> 16)
	b[4] = byte(h.CRC24 >> 8)
	b[5] = byte(h.CRC24)
	b[6] = h.Reserved
	return b
}

func unmarshalHeader(b []byte) (chunkHeader, error) {
	if len(b) < HeaderSize {
		return chunkHeader{}, fmt.Errorf("codematrix: header too short: %d bytes", len(b))
	}
	if b[0] != Version {
		return chunkHeader{}, fmt.Errorf("codematrix: unknown version: %d", b[0])
	}
	return chunkHeader{
		Version:  b[0],
		Index:    b[1],
		Total:    b[2],
		CRC24:    uint32(b[3])<<16 | uint32(b[4])<<8 | uint32(b[5]),
		Reserved: b[6],
	}, nil
}

func crc24Checksum(data []byte) uint32 {
	return crc24.Checksum(data)
}

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
