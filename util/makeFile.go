package util

import (
	"crypto/sha256"
	"fmt"
	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/pkg/errors"
	"hash/crc32"
	"time"
)

type PaperCrypt struct {
	// Version is the version of the paper crypt
	Version string

	// Data is the encrypted data
	Data *crypto.PGPMessage

	// SerialNumber is the serial number of the paper crypt
	SerialNumber string

	// Purpose is the purpose of the paper crypt
	Purpose string

	// Comment is the comment on the paper crypt
	Comment string

	// CreatedAt is the creation date of the paper crypt
	CreatedAt time.Time
}

// NewPaperCrypt creates a new paper crypt
func NewPaperCrypt(version string, data *crypto.PGPMessage, serialNumber string, purpose string, comment string, createdAt time.Time) *PaperCrypt {
	return &PaperCrypt{
		Version:      version,
		Data:         data,
		SerialNumber: serialNumber,
		Purpose:      purpose,
		Comment:      comment,
		CreatedAt:    createdAt,
	}
}

// SerializeBinary returns the encrypted binary data,
// formatted for restoration
// lines will hold 20 bytes of data, prefaces by the line number, followed by the CBC-24 of the line,
// bytes are printed as two base16 (hex) digits, separated by a space.
// Example:
//
//	1: 00 01 02 03 04 05 06 07 08 09 0A 0B 0C 0D 0E 0F 10 11 12 13 <CBC-24 of this line>
//	2: ...
//
// 10: ...
// ...
// n-1: ... <CBC-24 of this line>
// n: <CBC-24 of the block>
func SerializeBinary(data *[]byte) string {
	return string(*data)
}

// GetBinary returns the binary representation of the paper crypt
// TODO(2023-08-12): make this return pdf data, instead of acsii text
func (p *PaperCrypt) GetBinary(noQR bool) ([]byte, error) {
	return p.GetText(false)
}

// GetText returns the text representation of the paper crypt
func (p *PaperCrypt) GetText(armor bool) ([]byte, error) {
	var data []byte

	if armor {
		//stringData, err := p.Data.GetArmoredWithCustomHeaders(fmt.Sprintf("PaperCrypt/%s (https://github.com/TMUniversal/PaperCrypt), https://openpgp.org/", p.Version), constants.ArmorHeaderVersion)
		stringData, err := p.Data.GetArmored()
		if err != nil {
			return nil, errors.Errorf("error getting armored data: %s", err)
		}

		data = []byte(stringData)
	} else {
		data = p.Data.GetBinary()
	}

	crc32tableIEEE := crc32.MakeTable(crc32.IEEE)
	dataCRC32 := crc32.Checksum(data, crc32tableIEEE)
	dataSHA256 := sha256.Sum256(data)

	header := fmt.Sprintf(
		`PaperCrypt/%s
Content Serial: %s (Base32)
Purpose: %s
Comment: %s
Date: %s
Content CRC-32: %x
Content SHA-256: %x
Content Length: %d bytes`,
		p.Version,
		p.SerialNumber,
		p.Purpose,
		p.Comment,
		// format time with nanosecond precision
		// Sat, 12 Aug 2023 17:33:20.123456789
		p.CreatedAt.Format("Mon, 02 Jan 2006 15:04:05.000000000"),
		dataCRC32,
		dataSHA256,
		len(data))

	headerCRC32 := crc32.Checksum([]byte(header), crc32tableIEEE)

	serializedData := string(data)
	if armor {
		serializedData = SerializeBinary(&data)
	}

	return []byte(
		fmt.Sprintf(`%s
Header CRC-32: %x


%s
`,
			header,
			headerCRC32,
			serializedData)), nil
}
