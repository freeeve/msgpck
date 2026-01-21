package msgpck

import (
	"bytes"
	"testing"
)

func TestEncoderBufferGrowth(t *testing.T) {
	e := NewEncoder(1) // Start with tiny buffer

	// Write something larger than initial buffer
	bigString := string(make([]byte, 100))
	e.EncodeString(bigString)

	if len(e.Bytes()) == 0 {
		t.Error("encoder should have grown buffer")
	}
}

func TestEncoderStringFormats(t *testing.T) {
	e := NewEncoder(256)

	// fixstr (0-31 bytes)
	e.Reset()
	e.EncodeString("hi")
	if e.Bytes()[0]&0xe0 != 0xa0 {
		t.Error("short string should use fixstr")
	}

	// str8 (32-255 bytes)
	e.Reset()
	e.EncodeString(string(make([]byte, 50)))
	if e.Bytes()[0] != formatStr8 {
		t.Error("medium string should use str8")
	}

	// str16 (256-65535 bytes)
	e.Reset()
	e.EncodeString(string(make([]byte, 300)))
	if e.Bytes()[0] != formatStr16 {
		t.Error("large string should use str16")
	}
}

func TestEncoderBinaryFormats(t *testing.T) {
	e := NewEncoder(256)

	// bin8 (0-255 bytes)
	e.Reset()
	e.EncodeBinary(make([]byte, 50))
	if e.Bytes()[0] != formatBin8 {
		t.Error("small binary should use bin8")
	}

	// bin16 (256-65535 bytes)
	e.Reset()
	e.EncodeBinary(make([]byte, 300))
	if e.Bytes()[0] != formatBin16 {
		t.Error("medium binary should use bin16")
	}
}

func TestEncoderMapHeader(t *testing.T) {
	e := NewEncoder(32)

	// fixmap (0-15)
	e.Reset()
	e.EncodeMapHeader(5)
	if e.Bytes()[0]&0xf0 != 0x80 {
		t.Error("small map should use fixmap")
	}

	// map16 (16-65535)
	e.Reset()
	e.EncodeMapHeader(20)
	if e.Bytes()[0] != formatMap16 {
		t.Error("medium map should use map16")
	}
}

func TestEncoderArrayHeader(t *testing.T) {
	e := NewEncoder(32)

	// fixarray (0-15)
	e.Reset()
	e.EncodeArrayHeader(5)
	if e.Bytes()[0]&0xf0 != 0x90 {
		t.Error("small array should use fixarray")
	}

	// array16 (16-65535)
	e.Reset()
	e.EncodeArrayHeader(20)
	if e.Bytes()[0] != formatArray16 {
		t.Error("medium array should use array16")
	}
}

// TestEncoderIntFormats tests integer encoding formats
func TestEncoderIntFormats(t *testing.T) {
	tests := []struct {
		val    int64
		expect byte
	}{
		{0, 0x00},
		{127, 0x7f},
		{-1, 0xff},
		{-32, 0xe0},
		{-33, formatInt8},
		{128, formatUint8},
		{256, formatUint16},
		{65536, formatUint32},
		{-129, formatInt16},
		{-32769, formatInt32},
	}
	for _, tc := range tests {
		e := NewEncoder(16)
		e.EncodeInt(tc.val)
		if e.Bytes()[0] != tc.expect {
			t.Errorf("EncodeInt(%d): got format 0x%x, want 0x%x", tc.val, e.Bytes()[0], tc.expect)
		}
	}
}

// TestEncoderUintFormats tests unsigned integer encoding formats
func TestEncoderUintFormats(t *testing.T) {
	tests := []struct {
		val    uint64
		expect byte
	}{
		{0, 0x00},
		{127, 0x7f},
		{128, formatUint8},
		{256, formatUint16},
		{65536, formatUint32},
		{1 << 32, formatUint64},
	}
	for _, tc := range tests {
		e := NewEncoder(16)
		e.EncodeUint(tc.val)
		if e.Bytes()[0] != tc.expect {
			t.Errorf("EncodeUint(%d): got format 0x%x, want 0x%x", tc.val, e.Bytes()[0], tc.expect)
		}
	}
}

// TestEncoderStringFormatsTable tests string encoding formats with table-driven tests
func TestEncoderStringFormatsTable(t *testing.T) {
	tests := []struct {
		len    int
		expect byte
	}{
		{0, 0xa0},
		{31, 0xbf},
		{32, formatStr8},
		{255, formatStr8},
		{256, formatStr16},
		{65535, formatStr16},
	}
	for _, tc := range tests {
		e := NewEncoder(tc.len + 10)
		e.EncodeString(string(make([]byte, tc.len)))
		if e.Bytes()[0] != tc.expect {
			t.Errorf("EncodeString(len=%d): got format 0x%x, want 0x%x", tc.len, e.Bytes()[0], tc.expect)
		}
	}
}

// TestEncoderBinaryFormatsTable tests binary encoding formats with table-driven tests
func TestEncoderBinaryFormatsTable(t *testing.T) {
	tests := []struct {
		len    int
		expect byte
	}{
		{0, formatBin8},
		{255, formatBin8},
		{256, formatBin16},
		{65535, formatBin16},
	}
	for _, tc := range tests {
		e := NewEncoder(tc.len + 10)
		e.EncodeBinary(make([]byte, tc.len))
		if e.Bytes()[0] != tc.expect {
			t.Errorf("EncodeBinary(len=%d): got format 0x%x, want 0x%x", tc.len, e.Bytes()[0], tc.expect)
		}
	}
}

// TestEncoderArrayFormatsTable tests array header encoding formats
func TestEncoderArrayFormatsTable(t *testing.T) {
	tests := []struct {
		len    int
		expect byte
	}{
		{0, 0x90},
		{15, 0x9f},
		{16, formatArray16},
		{65535, formatArray16},
	}
	for _, tc := range tests {
		e := NewEncoder(16)
		e.EncodeArrayHeader(tc.len)
		if e.Bytes()[0] != tc.expect {
			t.Errorf("EncodeArrayHeader(%d): got format 0x%x, want 0x%x", tc.len, e.Bytes()[0], tc.expect)
		}
	}
}

// TestEncoderMapFormatsTable tests map header encoding formats
func TestEncoderMapFormatsTable(t *testing.T) {
	tests := []struct {
		len    int
		expect byte
	}{
		{0, 0x80},
		{15, 0x8f},
		{16, formatMap16},
		{65535, formatMap16},
	}
	for _, tc := range tests {
		e := NewEncoder(16)
		e.EncodeMapHeader(tc.len)
		if e.Bytes()[0] != tc.expect {
			t.Errorf("EncodeMapHeader(%d): got format 0x%x, want 0x%x", tc.len, e.Bytes()[0], tc.expect)
		}
	}
}

// TestEncoderFloatFormats tests float encoding formats
func TestEncoderFloatFormats(t *testing.T) {
	t.Run("float32", func(t *testing.T) {
		e := NewEncoder(16)
		e.EncodeFloat32(3.14)
		if e.Bytes()[0] != formatFloat32 {
			t.Error("EncodeFloat32 format wrong")
		}
	})

	t.Run("float64", func(t *testing.T) {
		e := NewEncoder(16)
		e.EncodeFloat64(3.14)
		if e.Bytes()[0] != formatFloat64 {
			t.Error("EncodeFloat64 format wrong")
		}
	})
}

// TestEncodeValue tests encoding Value types
func TestEncodeValue(t *testing.T) {
	tests := []struct {
		name string
		v    Value
	}{
		{"nil", Value{Type: TypeNil}},
		{"bool true", Value{Type: TypeBool, Bool: true}},
		{"bool false", Value{Type: TypeBool, Bool: false}},
		{"int", Value{Type: TypeInt, Int: -42}},
		{"uint", Value{Type: TypeUint, Uint: 42}},
		{"float32", Value{Type: TypeFloat32, Float32: 3.14}},
		{"float64", Value{Type: TypeFloat64, Float64: 3.14159}},
		{"string", Value{Type: TypeString, Bytes: []byte("hello")}},
		{"binary", Value{Type: TypeBinary, Bytes: []byte{1, 2, 3}}},
		{"array", Value{Type: TypeArray, Array: []Value{{Type: TypeInt, Int: 1}}}},
		{"map", Value{Type: TypeMap, Map: []KV{{Key: []byte("k"), Value: Value{Type: TypeInt, Int: 1}}}}},
		{"ext", Value{Type: TypeExt, Ext: Ext{Type: 1, Data: []byte{1, 2, 3, 4}}}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			e := NewEncoder(64)
			e.EncodeValue(&tc.v)
			if len(e.Bytes()) == 0 {
				t.Error("EncodeValue produced no output")
			}
		})
	}
}

// TestNewEncoderBuffer tests NewEncoderBuffer
func TestNewEncoderBuffer(t *testing.T) {
	buf := make([]byte, 0, 100)
	e := NewEncoderBuffer(buf)
	e.EncodeString("test")
	if e.Len() == 0 {
		t.Error("NewEncoderBuffer failed")
	}
}

// TestExtension tests extension type encoding/decoding
func TestExtension(t *testing.T) {
	e := NewEncoder(32)
	e.EncodeExt(42, []byte{1, 2, 3, 4})

	d := NewDecoder(e.Bytes())
	v, err := d.Decode()
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	if v.Type != TypeExt {
		t.Fatalf("expected TypeExt, got %v", v.Type)
	}
	if v.Ext.Type != 42 {
		t.Errorf("ext type = %d, want 42", v.Ext.Type)
	}
	if !bytes.Equal(v.Ext.Data, []byte{1, 2, 3, 4}) {
		t.Errorf("ext data = %v, want [1 2 3 4]", v.Ext.Data)
	}
}
