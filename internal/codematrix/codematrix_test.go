package codematrix

import (
	"bytes"
	"testing"
)

func TestHeaderMarshalUnmarshal(t *testing.T) {
	h := chunkHeader{
		Version:  Version,
		Index:    2,
		Total:    4,
		Reserved: 0,
	}
	b := h.Marshal()

	got, err := unmarshalHeader(b[:])
	if err != nil {
		t.Fatalf("unmarshalHeader: %v", err)
	}

	if got.Version != h.Version {
		t.Errorf("Version = %d, want %d", got.Version, h.Version)
	}
	if got.Index != h.Index {
		t.Errorf("Index = %d, want %d", got.Index, h.Index)
	}
	if got.Total != h.Total {
		t.Errorf("Total = %d, want %d", got.Total, h.Total)
	}
}

func TestHeaderBadVersion(t *testing.T) {
	b := [HeaderSize]byte{0xFF, 0, 0, 0}
	_, err := unmarshalHeader(b[:])
	if err == nil {
		t.Fatal("expected error for bad version")
	}
}

func TestHeaderTooShort(t *testing.T) {
	_, err := unmarshalHeader([]byte{0x01, 0x02})
	if err == nil {
		t.Fatal("expected error for short header")
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
		t.Errorf("decoded data = %q, want %q", got, data)
	}
}

func TestRoundtripLarge(t *testing.T) {
	// Generate data that exceeds MaxPayload to trigger multi-symbol split
	data := make([]byte, MaxPayload+100)
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

func TestRoundtripFourSymbols(t *testing.T) {
	data := make([]byte, MaxPayload*3+50)
	for i := range data {
		data[i] = byte(i % 251)
	}

	images, err := Encode(data)
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}

	if len(images) != 4 {
		t.Fatalf("got %d images, want 4", len(images))
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
	data := make([]byte, MaxPayload*MaxSymbols+1)
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
		t.Errorf("decoded data = %q, want %q", got, data)
	}
}
