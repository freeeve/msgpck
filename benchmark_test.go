package msgpck

import (
	"testing"
)

// Test structs
type SmallStruct struct {
	Name string
	Age  int
}

type MediumStruct struct {
	ID       int64
	Name     string
	Email    string
	Age      int
	Active   bool
	Score    float64
	Tags     []string
	Metadata map[string]string
}

var (
	smallStruct  = SmallStruct{Name: "Alice", Age: 30}
	mediumStruct = MediumStruct{
		ID: 12345, Name: "Bob Smith", Email: "bob@example.com",
		Age: 42, Active: true, Score: 98.6,
		Tags:     []string{"admin", "user", "premium"},
		Metadata: map[string]string{"role": "manager", "dept": "engineering"},
	}
	smallMap  = map[string]any{"name": "Alice", "age": 30}
	mediumMap = map[string]any{
		"id": int64(12345), "name": "Bob Smith", "email": "bob@example.com",
		"age": 42, "active": true, "score": 98.6,
		"tags":     []any{"admin", "user", "premium"},
		"metadata": map[string]any{"role": "manager", "dept": "engineering"},
	}
	stringMap = map[string]string{
		"name": "Alice", "email": "alice@example.com",
		"role": "admin", "dept": "engineering",
	}
)

// Pre-registered codecs (using cached versions)
var (
	smallStructDec   = GetStructDecoder[SmallStruct](false)
	smallStructDecZC = GetStructDecoder[SmallStruct](true)
	smallStructEnc   = GetStructEncoder[SmallStruct]()

	mediumStructDec   = GetStructDecoder[MediumStruct](false)
	mediumStructDecZC = GetStructDecoder[MediumStruct](true)
	mediumStructEnc   = GetStructEncoder[MediumStruct]()
)

// ============================================================================
// Map Encoding Benchmarks
// ============================================================================

func BenchmarkMsgpckSmallMapMarshal(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Marshal(smallMap)
	}
}

func BenchmarkMsgpckSmallMapMarshalCopy(b *testing.B) {
	for i := 0; i < b.N; i++ {
		MarshalCopy(smallMap)
	}
}

func BenchmarkMsgpckMediumMapMarshal(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Marshal(mediumMap)
	}
}

func BenchmarkMsgpckMediumMapMarshalCopy(b *testing.B) {
	for i := 0; i < b.N; i++ {
		MarshalCopy(mediumMap)
	}
}

// ============================================================================
// Map Decoding Benchmarks
// ============================================================================

func BenchmarkMsgpckSmallMapUnmarshal(b *testing.B) {
	data, _ := MarshalCopy(smallMap)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		UnmarshalMapStringAny(data, false)
	}
}

func BenchmarkMsgpckSmallMapUnmarshalZeroCopy(b *testing.B) {
	data, _ := MarshalCopy(smallMap)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		UnmarshalMapStringAny(data, true)
	}
}

func BenchmarkMsgpckMediumMapUnmarshal(b *testing.B) {
	data, _ := MarshalCopy(mediumMap)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		UnmarshalMapStringAny(data, false)
	}
}

func BenchmarkMsgpckMediumMapUnmarshalZeroCopy(b *testing.B) {
	data, _ := MarshalCopy(mediumMap)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		UnmarshalMapStringAny(data, true)
	}
}

// ============================================================================
// Typed Map Decoding Benchmarks
// ============================================================================

func BenchmarkMsgpckStringMapUnmarshal(b *testing.B) {
	data, _ := MarshalCopy(stringMap)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		UnmarshalMapStringString(data, false)
	}
}

func BenchmarkMsgpckStringMapUnmarshalZeroCopy(b *testing.B) {
	data, _ := MarshalCopy(stringMap)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		UnmarshalMapStringString(data, true)
	}
}

// ============================================================================
// Struct Encoding Benchmarks
// ============================================================================

func BenchmarkMsgpckSmallStructMarshal(b *testing.B) {
	for i := 0; i < b.N; i++ {
		MarshalCopy(smallStruct)
	}
}

func BenchmarkMsgpckSmallStructMarshalPreReg(b *testing.B) {
	for i := 0; i < b.N; i++ {
		smallStructEnc.Encode(&smallStruct)
	}
}

func BenchmarkMsgpckSmallStructMarshalPreRegCopy(b *testing.B) {
	for i := 0; i < b.N; i++ {
		smallStructEnc.EncodeCopy(&smallStruct)
	}
}

func BenchmarkMsgpckMediumStructMarshal(b *testing.B) {
	for i := 0; i < b.N; i++ {
		MarshalCopy(mediumStruct)
	}
}

func BenchmarkMsgpckMediumStructMarshalPreReg(b *testing.B) {
	for i := 0; i < b.N; i++ {
		mediumStructEnc.Encode(&mediumStruct)
	}
}

func BenchmarkMsgpckMediumStructMarshalPreRegCopy(b *testing.B) {
	for i := 0; i < b.N; i++ {
		mediumStructEnc.EncodeCopy(&mediumStruct)
	}
}

// ============================================================================
// Struct Decoding Benchmarks
// ============================================================================

func BenchmarkMsgpckSmallStructUnmarshal(b *testing.B) {
	data, _ := MarshalCopy(smallStruct)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var s SmallStruct
		UnmarshalStruct(data, &s)
	}
}

func BenchmarkMsgpckSmallStructUnmarshalPreReg(b *testing.B) {
	data, _ := MarshalCopy(smallStruct)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var s SmallStruct
		smallStructDec.Decode(data, &s)
	}
}

func BenchmarkMsgpckSmallStructUnmarshalZeroCopy(b *testing.B) {
	data, _ := MarshalCopy(smallStruct)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var s SmallStruct
		smallStructDecZC.Decode(data, &s)
	}
}

func BenchmarkMsgpckMediumStructUnmarshal(b *testing.B) {
	data, _ := MarshalCopy(mediumStruct)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var s MediumStruct
		UnmarshalStruct(data, &s)
	}
}

func BenchmarkMsgpckMediumStructUnmarshalPreReg(b *testing.B) {
	data, _ := MarshalCopy(mediumStruct)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var s MediumStruct
		mediumStructDec.Decode(data, &s)
	}
}

func BenchmarkMsgpckMediumStructUnmarshalZeroCopy(b *testing.B) {
	data, _ := MarshalCopy(mediumStruct)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var s MediumStruct
		mediumStructDecZC.Decode(data, &s)
	}
}

// ============================================================================
// Callback API Benchmarks (Safe Zero-Copy)
// ============================================================================

func BenchmarkMsgpckSmallStructCallback(b *testing.B) {
	data, _ := MarshalCopy(smallStruct)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		DecodeStructFunc(data, func(s *SmallStruct) error {
			_ = s.Name
			return nil
		})
	}
}

func BenchmarkMsgpckMediumStructCallback(b *testing.B) {
	data, _ := MarshalCopy(mediumStruct)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		DecodeStructFunc(data, func(s *MediumStruct) error {
			_ = s.Name
			return nil
		})
	}
}

func BenchmarkMsgpckSmallMapCallback(b *testing.B) {
	data, _ := MarshalCopy(smallMap)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		DecodeMapFunc(data, func(m map[string]any) error {
			_ = m["name"]
			return nil
		})
	}
}

func BenchmarkMsgpckStringMapCallback(b *testing.B) {
	data, _ := MarshalCopy(stringMap)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		DecodeStringMapFunc(data, func(m map[string]string) error {
			_ = m["name"]
			return nil
		})
	}
}

// ============================================================================
// Simple Benchmarks (from msgpck_test.go)
// ============================================================================

// BenchmarkDecodeMap benchmarks map decoding (the hot path)
func BenchmarkDecodeMap(b *testing.B) {
	data := map[string]any{
		"name":   "Alice",
		"age":    30,
		"email":  "alice@example.com",
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
		"name":   "Alice",
		"age":    30,
		"email":  "alice@example.com",
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
		"name":   "Alice",
		"age":    30,
		"email":  "alice@example.com",
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
