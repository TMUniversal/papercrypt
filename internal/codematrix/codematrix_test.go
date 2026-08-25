package codematrix

import (
	"bytes"
	"image"
	"testing"
)

func TestRoundtrip(t *testing.T) {
	data := append(
		[]byte(`{"v":"3.0.0-dev","f":"PGP","sn":"test","d":"`),
		bytes.Repeat([]byte("hello world "), 100)...,
	)
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

func TestRoundtripRawBytes(t *testing.T) {
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
