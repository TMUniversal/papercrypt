// Package codematrix provides single-Aztec-code encoding with adaptive
// compression and CRC32 integrity verification.
package codematrix

const (
	// EncGzip indicates the payload is gzip+base64 encoded.
	EncGzip = 'G'
	// EncRaw indicates the payload is stored raw.
	EncRaw = 'R'

	headerSize = 9 // 1 flag + 8 CRC32 hex
)
