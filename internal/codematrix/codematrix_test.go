package codematrix

import (
	"bytes"
	"testing"

	"github.com/makiuchi-d/gozxing"
)

func TestSARoundtrip(t *testing.T) {
	h := SAHeader{Index: 2, Total: 4}
	bits := h.Encode()
	got, off, err := DecodeSAHeader(bits, 0)
	if err != nil {
		t.Fatalf("DecodeSAHeader: %v", err)
	}
	if off != 16 {
		t.Fatalf("offset = %d, want 16", off)
	}
	if got.Index != 2 || got.Total != 4 {
		t.Errorf("got {%d,%d}, want {2,4}", got.Index, got.Total)
	}
}

func TestSABadMode(t *testing.T) {
	bits := gozxing.NewEmptyBitArray()
	bits.AppendBits(0x00, 16)
	_, _, err := DecodeSAHeader(bits, 0)
	if err == nil {
		t.Fatal("expected error for bad SA mode")
	}
}

func TestDataHeaderRoundtrip(t *testing.T) {
	data := []byte("hello")
	h := DataHeader{Mode: 0x04, Length: 5}
	bits := h.Encode(qrVersion, data)

	got, _, err := DecodeDataHeader(bits, 0, qrVersion)
	if err != nil {
		t.Fatalf("DecodeDataHeader: %v", err)
	}
	if got.Mode != 0x04 {
		t.Errorf("Mode = %x, want 04", got.Mode)
	}
	if got.Length != 5 {
		t.Errorf("Length = %d, want 5", got.Length)
	}
	if got.CRC32 == 0 {
		t.Error("CRC32 should not be zero")
	}
}

func TestDataHeaderBadMode(t *testing.T) {
	bits := gozxing.NewEmptyBitArray()
	bits.AppendBits(0x01, 4)
	bits.AppendBits(0, 16)
	bits.AppendBits(0, 32)
	_, _, err := DecodeDataHeader(bits, 0, qrVersion)
	if err == nil {
		t.Fatal("expected error for bad data mode")
	}
}

func TestGridDimensions(t *testing.T) {
	tests := []struct {
		n            int
		wantR, wantC int
	}{
		{0, 1, 1},
		{1, 1, 1},
		{2, 1, 2},
		{3, 2, 2},
		{4, 2, 2},
	}
	for _, tt := range tests {
		r, c := GridDimensions(tt.n)
		if r != tt.wantR || c != tt.wantC {
			t.Errorf("GridDimensions(%d) = (%d, %d), want (%d, %d)", tt.n, r, c, tt.wantR, tt.wantC)
		}
	}
}

func TestCodeLabel(t *testing.T) {
	if got := CodeLabel(0, 2); got != "1/2" {
		t.Errorf("CodeLabel(0,2) = %q, want %q", got, "1/2")
	}
	if got := CodeLabel(3, 4); got != "4/4" {
		t.Errorf("CodeLabel(3,4) = %q, want %q", got, "4/4")
	}
}

func TestRoundtripSmall(t *testing.T) {
	data := []byte("hello world")
	images, err := Encode(data)
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}
	if len(images) != 1 {
		t.Fatalf("got %d images, want 1", len(images))
	}
	got, err := Decode(images)
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}
	if !bytes.Equal(got, data) {
		t.Errorf("decoded = %q, want %q", got, data)
	}
}

func TestRoundtripLarge(t *testing.T) {
	data := make([]byte, maxChunkBytes+100)
	for i := range data {
		data[i] = byte(i % 251)
	}
	images, err := Encode(data)
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}
	if len(images) != 2 {
		t.Fatalf("got %d images, want 2", len(images))
	}
	got, err := Decode(images)
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}
	if !bytes.Equal(got, data) {
		t.Error("decoded data does not match original")
	}
}

func TestTooLarge(t *testing.T) {
	data := make([]byte, maxChunkBytes*MaxSymbols+1)
	_, err := Encode(data)
	if err == nil {
		t.Fatal("expected error for data too large")
	}
}

func TestDecodeNoImages(t *testing.T) {
	_, err := Decode(nil)
	if err == nil {
		t.Fatal("expected error for no images")
	}
}

func TestRoundtripEmpty(t *testing.T) {
	data := []byte{}
	images, err := Encode(data)
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}
	if len(images) != 1 {
		t.Fatalf("got %d images, want 1", len(images))
	}
	got, err := Decode(images)
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}
	if !bytes.Equal(got, data) {
		t.Errorf("decoded = %q, want %q", got, data)
	}
}
