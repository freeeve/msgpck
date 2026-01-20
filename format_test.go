package msgpck

import (
	"testing"
)

func TestStringFormats(t *testing.T) {
	tests := []struct {
		len int
	}{
		{0},
		{31},   // max fixstr
		{32},   // str8
		{255},  // max str8
		{256},  // str16
		{1000}, // str16
	}

	for _, tc := range tests {
		s := string(make([]byte, tc.len))
		e := NewEncoder(tc.len + 10)
		e.EncodeStringBytes([]byte(s))
		if len(e.Bytes()) == 0 {
			t.Errorf("EncodeStringBytes len=%d failed", tc.len)
		}
	}
}

func TestBinaryFormats(t *testing.T) {
	tests := []struct {
		len int
	}{
		{0},
		{255},   // max bin8
		{256},   // bin16
		{65535}, // max bin16
	}

	for _, tc := range tests {
		b := make([]byte, tc.len)
		e := NewEncoder(tc.len + 10)
		e.EncodeBinary(b)
		if len(e.Bytes()) == 0 {
			t.Errorf("EncodeBinary len=%d failed", tc.len)
		}
	}
}

func TestMapFormats(t *testing.T) {
	tests := []struct {
		len int
	}{
		{0},
		{15},    // max fixmap
		{16},    // map16
		{65535}, // max map16
	}

	for _, tc := range tests {
		e := NewEncoder(tc.len*4 + 10)
		e.EncodeMapHeader(tc.len)
		if len(e.Bytes()) == 0 {
			t.Errorf("EncodeMapHeader len=%d failed", tc.len)
		}
	}
}

func TestArrayFormats(t *testing.T) {
	tests := []struct {
		len int
	}{
		{0},
		{15},    // max fixarray
		{16},    // array16
		{65535}, // max array16
	}

	for _, tc := range tests {
		e := NewEncoder(tc.len + 10)
		e.EncodeArrayHeader(tc.len)
		if len(e.Bytes()) == 0 {
			t.Errorf("EncodeArrayHeader len=%d failed", tc.len)
		}
	}
}

func TestExtFormats(t *testing.T) {
	tests := []struct {
		name string
		data []byte
	}{
		{"fixext1", []byte{formatFixExt1, 1, 0xff}},
		{"fixext2", []byte{formatFixExt2, 1, 0xff, 0xff}},
		{"fixext4", []byte{formatFixExt4, 1, 0xff, 0xff, 0xff, 0xff}},
		{"fixext8", []byte{formatFixExt8, 1, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}},
		{"fixext16", append([]byte{formatFixExt16, 1}, make([]byte, 16)...)},
		{"ext8", []byte{formatExt8, 3, 1, 0xff, 0xff, 0xff}},
		{"ext16", append([]byte{formatExt16, 0, 3, 1}, make([]byte, 3)...)},
		{"ext32", append([]byte{formatExt32, 0, 0, 0, 3, 1}, make([]byte, 3)...)},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			d := NewDecoder(tc.data)
			v, err := d.Decode()
			if err != nil {
				t.Errorf("%s decode failed: %v", tc.name, err)
			}
			if v.Type != TypeExt {
				t.Errorf("%s: expected ext type, got %v", tc.name, v.Type)
			}
		})
	}
}

func TestTypeString(t *testing.T) {
	types := []Type{
		TypeNil,
		TypeBool,
		TypeInt,
		TypeUint,
		TypeFloat32,
		TypeFloat64,
		TypeString,
		TypeBinary,
		TypeArray,
		TypeMap,
		TypeExt,
	}

	for _, typ := range types {
		s := typ.String()
		if s == "" {
			t.Errorf("String() returned empty for type %v", typ)
		}
	}
}
