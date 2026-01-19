package msgpck

import (
	"bytes"
	"math"
	"reflect"
	"testing"
)

// TestRoundTripPrimitives tests encoding and decoding of primitive types
func TestRoundTripPrimitives(t *testing.T) {
	// Note: DecodeAny normalizes all integers to int64 (unless > MaxInt64)
	// and float32 to float64 for consistent types
	tests := []struct {
		name  string
		value any
	}{
		{"nil", nil},
		{"true", true},
		{"false", false},
		{"zero", int64(0)},
		{"positive fixint", int64(42)},
		{"max fixint", int64(127)},
		{"negative fixint", int64(-1)},
		{"min fixint", int64(-32)},
		{"uint8", int64(200)},
		{"uint16", int64(1000)},
		{"uint32", int64(100000)},
		{"uint64", int64(1 << 40)},
		{"int8", int64(-100)},
		{"int16", int64(-1000)},
		{"int32", int64(-100000)},
		{"int64", int64(-1 << 40)},
		{"float32", float64(3.140000104904175)}, // float32 promoted to float64
		{"float64", float64(3.14159265359)},
		{"empty string", ""},
		{"short string", "hello"},
		{"fixstr max", string(make([]byte, 31))},
		{"str8", string(make([]byte, 100))},
		{"str16", string(make([]byte, 300))},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoded, err := Marshal(tt.value)
			if err != nil {
				t.Fatalf("Marshal failed: %v", err)
			}

			decoded, err := Unmarshal(encoded)
			if err != nil {
				t.Fatalf("Unmarshal failed: %v", err)
			}

			if !reflect.DeepEqual(decoded, tt.value) {
				t.Errorf("got %v (%T), want %v (%T)", decoded, decoded, tt.value, tt.value)
			}
		})
	}
}

// TestRoundTripContainers tests arrays and maps
func TestRoundTripContainers(t *testing.T) {
	// Note: all integers decode to int64
	tests := []struct {
		name  string
		value any
	}{
		{"empty array", []any{}},
		{"int array", []any{int64(1), int64(2), int64(3)}},
		{"mixed array", []any{int64(1), "hello", true, nil}},
		{"empty map", map[string]any{}},
		{"string map", map[string]any{"a": int64(1), "b": int64(2)}},
		{"nested map", map[string]any{
			"inner": map[string]any{"x": int64(10)},
		}},
		{"nested array", []any{[]any{int64(1), int64(2)}, []any{int64(3), int64(4)}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoded, err := Marshal(tt.value)
			if err != nil {
				t.Fatalf("Marshal failed: %v", err)
			}

			decoded, err := Unmarshal(encoded)
			if err != nil {
				t.Fatalf("Unmarshal failed: %v", err)
			}

			if !reflect.DeepEqual(decoded, tt.value) {
				t.Errorf("got %v, want %v", decoded, tt.value)
			}
		})
	}
}

// TestZeroCopyDecode tests that string data points into source buffer
func TestZeroCopyDecode(t *testing.T) {
	data := []byte{0xa5, 'h', 'e', 'l', 'l', 'o'} // fixstr "hello"

	d := NewDecoder(data)
	v, err := d.Decode()
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	if v.Type != TypeString {
		t.Fatalf("expected TypeString, got %v", v.Type)
	}

	// Check that Bytes points into the original data
	if &v.Bytes[0] != &data[1] {
		t.Error("string bytes not pointing into source buffer (not zero-copy)")
	}
}

// TestDecodeStruct tests struct unmarshaling
func TestDecodeStruct(t *testing.T) {
	type Person struct {
		Name string `msgpack:"name"`
		Age  int    `msgpack:"age"`
	}

	// Encode
	original := Person{Name: "Alice", Age: 30}
	encoded, err := Marshal(original)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	// Decode
	var decoded Person
	err = UnmarshalStruct(encoded, &decoded)
	if err != nil {
		t.Fatalf("UnmarshalStruct failed: %v", err)
	}

	if decoded != original {
		t.Errorf("got %+v, want %+v", decoded, original)
	}
}

// TestOmitEmpty tests omitempty tag
func TestOmitEmpty(t *testing.T) {
	type Data struct {
		Name  string `msgpack:"name"`
		Value int    `msgpack:"value,omitempty"`
	}

	// Value is zero, should be omitted
	original := Data{Name: "test", Value: 0}
	encoded, err := Marshal(original)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	// Decode as map to check fields
	decoded, err := Unmarshal(encoded)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	m := decoded.(map[string]any)
	if _, ok := m["value"]; ok {
		t.Error("value field should be omitted")
	}
	if m["name"] != "test" {
		t.Errorf("name = %v, want test", m["name"])
	}
}

// TestSecurityLimits tests that security limits are enforced
func TestSecurityLimits(t *testing.T) {
	t.Run("string too long", func(t *testing.T) {
		cfg := DefaultConfig().WithMaxStringLen(10)
		// str8 with length 100
		data := []byte{FormatStr8, 100}
		data = append(data, make([]byte, 100)...)

		d := NewDecoderWithConfig(data, cfg)
		_, err := d.Decode()
		if err != ErrStringTooLong {
			t.Errorf("expected ErrStringTooLong, got %v", err)
		}
	})

	t.Run("array too long", func(t *testing.T) {
		cfg := DefaultConfig().WithMaxArrayLen(5)
		// array16 with length 1000
		data := []byte{FormatArray16, 0x03, 0xe8} // 1000

		d := NewDecoderWithConfig(data, cfg)
		_, err := d.Decode()
		if err != ErrArrayTooLong {
			t.Errorf("expected ErrArrayTooLong, got %v", err)
		}
	})

	t.Run("map too long", func(t *testing.T) {
		cfg := DefaultConfig().WithMaxMapLen(5)
		// map16 with length 1000
		data := []byte{FormatMap16, 0x03, 0xe8}

		d := NewDecoderWithConfig(data, cfg)
		_, err := d.Decode()
		if err != ErrMapTooLong {
			t.Errorf("expected ErrMapTooLong, got %v", err)
		}
	})

	t.Run("max depth exceeded", func(t *testing.T) {
		cfg := DefaultConfig().WithMaxDepth(3)
		// Deeply nested arrays: [[[[1]]]]
		data := []byte{
			0x91, // fixarray 1
			0x91, // fixarray 1
			0x91, // fixarray 1
			0x91, // fixarray 1 - this exceeds depth 3
			0x01, // fixint 1
		}

		d := NewDecoderWithConfig(data, cfg)
		_, err := d.Decode()
		if err != ErrMaxDepthExceeded {
			t.Errorf("expected ErrMaxDepthExceeded, got %v", err)
		}
	})
}

// TestUnexpectedEOF tests handling of truncated data
func TestUnexpectedEOF(t *testing.T) {
	tests := []struct {
		name string
		data []byte
	}{
		{"empty", []byte{}},
		{"truncated uint16", []byte{FormatUint16, 0x01}},
		{"truncated string", []byte{0xa5, 'h', 'e'}}, // fixstr 5, only 2 chars
		{"truncated array", []byte{0x92, 0x01}},      // fixarray 2, only 1 element
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := NewDecoder(tt.data)
			_, err := d.Decode()
			if err != ErrUnexpectedEOF {
				t.Errorf("expected ErrUnexpectedEOF, got %v", err)
			}
		})
	}
}

// TestBinaryData tests encoding/decoding of binary data
func TestBinaryData(t *testing.T) {
	original := []byte{0x00, 0x01, 0x02, 0xff, 0xfe}
	encoded, err := Marshal(original)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	decoded, err := Unmarshal(encoded)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if !bytes.Equal(decoded.([]byte), original) {
		t.Errorf("got %v, want %v", decoded, original)
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

// TestFloatSpecialValues tests special float values
func TestFloatSpecialValues(t *testing.T) {
	tests := []struct {
		name  string
		value float64
	}{
		{"positive infinity", math.Inf(1)},
		{"negative infinity", math.Inf(-1)},
		{"NaN", math.NaN()},
		{"max float64", math.MaxFloat64},
		{"smallest positive", math.SmallestNonzeroFloat64},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoded, err := Marshal(tt.value)
			if err != nil {
				t.Fatalf("Marshal failed: %v", err)
			}

			decoded, err := Unmarshal(encoded)
			if err != nil {
				t.Fatalf("Unmarshal failed: %v", err)
			}

			got := decoded.(float64)
			if math.IsNaN(tt.value) {
				if !math.IsNaN(got) {
					t.Errorf("expected NaN, got %v", got)
				}
			} else if got != tt.value {
				t.Errorf("got %v, want %v", got, tt.value)
			}
		})
	}
}

// TestNestedStruct tests nested struct decoding
func TestNestedStruct(t *testing.T) {
	type Address struct {
		City string `msgpack:"city"`
	}
	type Person struct {
		Name    string  `msgpack:"name"`
		Address Address `msgpack:"address"`
	}

	original := Person{
		Name:    "Alice",
		Address: Address{City: "NYC"},
	}

	encoded, err := Marshal(original)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var decoded Person
	err = UnmarshalStruct(encoded, &decoded)
	if err != nil {
		t.Fatalf("UnmarshalStruct failed: %v", err)
	}

	if decoded != original {
		t.Errorf("got %+v, want %+v", decoded, original)
	}
}

// TestSliceInStruct tests slice field in struct
func TestSliceInStruct(t *testing.T) {
	type Data struct {
		Values []int `msgpack:"values"`
	}

	original := Data{Values: []int{1, 2, 3}}

	encoded, err := Marshal(original)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var decoded Data
	err = UnmarshalStruct(encoded, &decoded)
	if err != nil {
		t.Fatalf("UnmarshalStruct failed: %v", err)
	}

	if !reflect.DeepEqual(decoded, original) {
		t.Errorf("got %+v, want %+v", decoded, original)
	}
}

// TestMapInStruct tests map field in struct
func TestMapInStruct(t *testing.T) {
	type Data struct {
		Meta map[string]string `msgpack:"meta"`
	}

	original := Data{Meta: map[string]string{"key": "value"}}

	encoded, err := Marshal(original)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var decoded Data
	err = UnmarshalStruct(encoded, &decoded)
	if err != nil {
		t.Fatalf("UnmarshalStruct failed: %v", err)
	}

	if !reflect.DeepEqual(decoded, original) {
		t.Errorf("got %+v, want %+v", decoded, original)
	}
}

// BenchmarkDecodeMap benchmarks map decoding (the hot path)
func BenchmarkDecodeMap(b *testing.B) {
	data := map[string]any{
		"name":  "Alice",
		"age":   30,
		"email": "alice@example.com",
		"active": true,
	}
	encoded, _ := Marshal(data)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		d := NewDecoder(encoded)
		_, _ = d.Decode()
	}
}

// BenchmarkDecodeMapAny benchmarks decoding to map[string]any
func BenchmarkDecodeMapAny(b *testing.B) {
	data := map[string]any{
		"name":  "Alice",
		"age":   30,
		"email": "alice@example.com",
		"active": true,
	}
	encoded, _ := Marshal(data)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		d := NewDecoder(encoded)
		_, _ = d.DecodeAny()
	}
}

// BenchmarkEncodeMap benchmarks map encoding
func BenchmarkEncodeMap(b *testing.B) {
	data := map[string]any{
		"name":  "Alice",
		"age":   30,
		"email": "alice@example.com",
		"active": true,
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, _ = Marshal(data)
	}
}

// BenchmarkDecodeStruct benchmarks struct decoding
func BenchmarkDecodeStruct(b *testing.B) {
	type Person struct {
		Name   string `msgpack:"name"`
		Age    int    `msgpack:"age"`
		Email  string `msgpack:"email"`
		Active bool   `msgpack:"active"`
	}

	data := Person{
		Name:   "Alice",
		Age:    30,
		Email:  "alice@example.com",
		Active: true,
	}
	encoded, _ := Marshal(data)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		var p Person
		d := NewDecoder(encoded)
		_ = d.DecodeStruct(&p)
	}
}
