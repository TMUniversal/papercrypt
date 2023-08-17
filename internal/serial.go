package internal

import (
	"bytes"
	"crypto/rand"
	"encoding/base32"
	"math"
	"math/big"

	"github.com/pkg/errors"
)

// GenerateSerial generates a random serial number of length `length`
func GenerateSerial(length uint8) (string, error) {
	// generate `length` random bytes,
	// encode them as base64,
	// and return the first `length` characters

	numbers := make([]*big.Int, length)

	for i := uint8(0); i < length; i++ {
		randInt, err := rand.Int(rand.Reader, big.NewInt(math.MaxInt64))
		if err != nil {
			return "", errors.Errorf("error generating random bytes: %s", err)
		}

		numbers[i] = randInt
	}

	buf := new(bytes.Buffer)
	encoder := base32.NewEncoder(base32.StdEncoding, buf)
	for _, number := range numbers {
		_, err := encoder.Write(number.Bytes())
		if err != nil {
			return "", errors.Errorf("error encoding bytes: %s", err)
		}
	}
	err := encoder.Close()
	if err != nil {
		return "", errors.Errorf("error closing base64 encoder: %s", err)
	}

	return buf.String()[:length], nil
}

// DecodeSerial decodes a serial number
func DecodeSerial(serial string) ([]byte, error) {
	decoder := base32.NewDecoder(base32.StdEncoding, bytes.NewBufferString(serial))
	var decoded []byte
	_, err := decoder.Read(decoded)
	if err != nil {
		return nil, errors.Errorf("error decoding serial: %s", err)
	}

	return decoded, nil
}
