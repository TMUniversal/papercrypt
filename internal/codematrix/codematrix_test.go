package codematrix

import (
	"bytes"
	"fmt"
	"hash/crc32"
	"image"
	"testing"
)

func TestRoundtripGzip(t *testing.T) {
	data := bytes.Repeat([]byte("hello world "), 100)
	data = append([]byte(`{"v":"3","d":"`), data...)
	data = append(data, []byte(`"}`)...)
	img, err := Encode(data)
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}
	got, err := Decode(img)
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}
	if !bytes.Equal(got, data) {
		t.Errorf("roundtrip mismatch")
	}
}

func TestRoundtripRaw(t *testing.T) {
	data := make([]byte, 500)
	for i := range data {
		data[i] = byte(i % 251)
	}
	img, err := Encode(data)
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}
	got, err := Decode(img)
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}
	if !bytes.Equal(got, data) {
		t.Errorf("roundtrip mismatch")
	}
}

func TestRoundtripEmpty(t *testing.T) {
	img, err := Encode([]byte{})
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}
	got, err := Decode(img)
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected empty, got %d bytes", len(got))
	}
}

func TestHeaderParse(t *testing.T) {
	raw := []byte("G" + "DEADBEEF" + "payload")
	flag, crc, payload, err := parseHeader(raw)
	if err != nil {
		t.Fatalf("parseHeader: %v", err)
	}
	if flag != 'G' {
		t.Errorf("flag = %c, want G", flag)
	}
	if crc != 0xDEADBEEF {
		t.Errorf("crc = %08X, want DEADBEEF", crc)
	}
	if string(payload) != "payload" {
		t.Errorf("payload = %q, want payload", payload)
	}
}

func TestHeaderTooShort(t *testing.T) {
	_, _, _, err := parseHeader([]byte("G"))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestHeaderBadFlag(t *testing.T) {
	_, _, _, err := parseHeader([]byte("X" + "00000000" + "data"))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestEncodePNG(t *testing.T) {
	data := []byte("hello")
	pngBytes, err := EncodePNG(data)
	if err != nil {
		t.Fatalf("EncodePNG: %v", err)
	}
	if len(pngBytes) == 0 {
		t.Fatal("empty PNG output")
	}
	img, _, err := image.Decode(bytes.NewReader(pngBytes))
	if err != nil {
		t.Fatalf("decode PNG: %v", err)
	}
	got, err := Decode(img)
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}
	if !bytes.Equal(got, data) {
		t.Errorf("roundtrip mismatch")
	}
}

func TestFlagSelection(t *testing.T) {
	compressible := bytes.Repeat([]byte("x"), 1000)
	compressible = append(
		[]byte(`{"data":"`), compressible...,
	)
	compressible = append(compressible, []byte(`"}`)...)
	rawPayload := make([]byte, 0, headerSize+len(compressible))
	rawPayload = append(rawPayload, EncRaw)
	crcHex := fmt.Sprintf("%08X", crc32Checksum(compressible))
	rawPayload = append(rawPayload, []byte(crcHex)...)
	rawPayload = append(rawPayload, compressible...)

	gzBytes, _ := gzipBase64(compressible)
	gzPayload := make([]byte, 0, headerSize+len(gzBytes))
	gzPayload = append(gzPayload, EncGzip)
	gzPayload = append(gzPayload, []byte(crcHex)...)
	gzPayload = append(gzPayload, gzBytes...)

	if len(gzPayload) >= len(rawPayload) {
		t.Errorf("gzip+base64 (%d) should be smaller than raw (%d)",
			len(gzPayload), len(rawPayload))
	}
}

func crc32Checksum(data []byte) uint32 {
	return crc32.ChecksumIEEE(data)
}
